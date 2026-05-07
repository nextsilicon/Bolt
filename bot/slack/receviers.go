package slack

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/shlex"
	"github.com/oriser/bolt/service"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (s *SlackBot) ListenAndServe(ctx context.Context) error {
	for i := 0; i < s.mentionsWorkers; i++ {
		go s.mentionsWorker(ctx)
	}
	for i := 0; i < s.linksWorkers; i++ {
		go s.linksWorker(ctx)
	}
	for i := 0; i < s.reactionsWorkers; i++ {
		go s.reactionsAddWorker(ctx)
	}

	go func() {
		if err := s.socketClient.Run(); err != nil {
			log.Printf("socketmode: connection closed: %v\n", err)
		}
	}()

	log.Println("Listening for Slack events via Socket Mode")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-s.socketClient.Events:
			if !ok {
				return nil
			}
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("unexpected data type for EventsAPI: %T\n", event.Data)
					continue
				}
				s.socketClient.Ack(*event.Request)
				s.dispatchEvent(eventsAPIEvent)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := event.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("unexpected data type for slash command: %T\n", event.Data)
					continue
				}
				go s.handleSlashCommand(ctx, event, cmd)
			}
		}
	}
}

func (s *SlackBot) dispatchEvent(event slackevents.EventsAPIEvent) {
	if event.Type != slackevents.CallbackEvent {
		return
	}
	switch ev := event.InnerEvent.Data.(type) {
	case *slackevents.ReactionAddedEvent:
		select {
		case s.reactionsAddCh <- ev:
		default:
			log.Println("reactions channel full, dropping event")
		}
	case *slackevents.AppMentionEvent:
		select {
		case s.mentionsCh <- ev:
		default:
			log.Println("mentions channel full, dropping event")
		}
	case *slackevents.LinkSharedEvent:
		select {
		case s.linksCh <- ev:
		default:
			log.Println("links channel full, dropping event")
		}
	}
}

func (s *SlackBot) handleMention(_ *slackevents.AppMentionEvent) error {
	// Currently unimplemented
	return nil
}

func (s *SlackBot) mentionsWorker(ctx context.Context) {
	for {
		select {
		case event := <-s.mentionsCh:
			if err := s.handleMention(event); err != nil {
				log.Println("Error handling mention:", err)
			}
		case <-ctx.Done():
			log.Println("Finishing mention worker due to context cancellation")
			return
		}
	}
}

func (s *SlackBot) handleLink(linkEvent *slackevents.LinkSharedEvent) error {
	if linkEvent.Channel == "COMPOSER" {
		// COMPOSER link events fire when a link is unfurled in the composer but not yet sent;
		// there is no valid channel to respond to.
		return nil
	}

	links := make([]service.Link, len(linkEvent.Links))
	for i, l := range linkEvent.Links {
		links[i] = service.Link{
			Domain: l.Domain,
			URL:    l.URL,
		}
	}

	response, err := s.service.HandleLinkMessage(service.LinksRequest{
		Links:     links,
		MessageID: linkEvent.MessageTimeStamp,
		Channel:   linkEvent.Channel,
	})
	if err != nil {
		return fmt.Errorf("link handler: %w", err)
	}

	if response != "" {
		if _, _, err := s.PostMessage(linkEvent.Channel, slack.MsgOptionText(response, false)); err != nil {
			return fmt.Errorf("post message: %w", err)
		}
	}
	return nil
}

func (s *SlackBot) linksWorker(ctx context.Context) {
	for {
		select {
		case event := <-s.linksCh:
			if err := s.handleLink(event); err != nil {
				log.Println("Error handling link:", err)
			}
		case <-ctx.Done():
			log.Println("Finishing links worker due to context cancellation")
			return
		}
	}
}

func (s *SlackBot) handleReactionAdd(event *slackevents.ReactionAddedEvent) error {
	msgs, _, _, err := s.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: event.Item.Channel,
		Timestamp: event.Item.Timestamp,
		Limit:     1,
	})
	if err != nil {
		return fmt.Errorf("GetConversationReplies: %w", err)
	}
	if len(msgs) == 0 {
		return fmt.Errorf("no messages found for conversation with ts %s", event.Item.Timestamp)
	}

	response, err := s.service.HandleReactionAdded(service.ReactionAddRequest{
		Reaction:      event.Reaction,
		FromUserID:    event.User,
		Channel:       event.Item.Channel,
		MessageUserID: event.ItemUser,
		MessageText:   msgs[0].Text,
	})
	if err != nil {
		return fmt.Errorf("reaction add handler: %w", err)
	}

	if response != "" {
		if _, _, err := s.PostMessage(event.Item.Channel, slack.MsgOptionText(response, false)); err != nil {
			return fmt.Errorf("post message: %w", err)
		}
	}
	return nil
}

func (s *SlackBot) reactionsAddWorker(ctx context.Context) {
	for {
		select {
		case event := <-s.reactionsAddCh:
			if err := s.handleReactionAdd(event); err != nil {
				log.Println("Error handling reaction:", err)
			}
		case <-ctx.Done():
			log.Println("Finishing reaction worker due to context cancellation")
			return
		}
	}
}

func (s *SlackBot) getUserByUserName(ctx context.Context, userName string) (slack.User, error) {
	var err error
	paginatedUsers := s.GetUsersPaginated()
	for {
		paginatedUsers, err = paginatedUsers.Next(ctx)
		if err != nil {
			break
		}
		for _, user := range paginatedUsers.Users {
			if user.Name == userName {
				return user, nil
			}
		}
	}
	if err = paginatedUsers.Failure(err); err != nil {
		return slack.User{}, fmt.Errorf("list users: %w", err)
	}
	return slack.User{}, fmt.Errorf("user %q not found", userName)
}

func (s *SlackBot) handleSlashCommand(ctx context.Context, event socketmode.Event, cmd slack.SlashCommand) {
	var responseText string
	switch cmd.Command {
	case "/add-user":
		responseText = s.processAddUserCommand(ctx, cmd)
	default:
		responseText = fmt.Sprintf("unknown command: %s", cmd.Command)
	}
	s.socketClient.Ack(*event.Request, map[string]interface{}{
		"text": responseText,
	})
}

func (s *SlackBot) processAddUserCommand(ctx context.Context, cmd slack.SlashCommand) string {
	if _, ok := s.adminsUserIds[cmd.UserID]; !ok {
		return "Unauthorized"
	}

	splitted, err := shlex.Split(cmd.Text)
	if err != nil {
		return fmt.Sprintf("error parsing command: %v", err)
	}
	if len(splitted) != 2 || !strings.HasPrefix(splitted[1], "@") {
		return `USAGE: "<name>" @<user>`
	}

	user, err := s.getUserByUserName(ctx, splitted[1][1:])
	if err != nil {
		return err.Error()
	}

	if err := s.service.HandleAddUser(splitted[0], user); err != nil {
		return fmt.Sprintf("error adding user: %v", err)
	}
	return fmt.Sprintf("OK, got you. I added <@%s> as %q", user.ID, splitted[0])
}
