package slack

import (
	"fmt"

	"github.com/oriser/bolt/service"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Config struct {
	AppToken               string   `env:"SLACK_APP_TOKEN,required" json:"-"`
	ClientSecret           string   `env:"SLACK_OAUTH_TOKEN,required" json:"-"`
	MaxConcurrentLinks     int      `env:"SLACK_MAX_CONCURRENT_LINKS" envDefault:"100"`
	MaxConcurrentMentions  int      `env:"SLACK_MAX_CONCURRENT_MENTIONS" envDefault:"100"`
	MaxConcurrentReactions int      `env:"SLACK_MAX_CONCURRENT_REACTIONS" envDefault:"100"`
	AdminSlackUserID       []string `env:"ADMIN_SLACK_USER_IDS"`
	SlackAPIUrl            string   `env:"SLACK_API_URL"` // only for testing
}

type SlackBot struct {
	*slack.Client
	socketClient     *socketmode.Client
	service          *service.Service
	mentionsWorkers  int
	linksWorkers     int
	reactionsWorkers int
	adminsUserIds    map[string]interface{}
	mentionsCh       chan *slackevents.AppMentionEvent
	linksCh          chan *slackevents.LinkSharedEvent
	reactionsAddCh   chan *slackevents.ReactionAddedEvent
}

type Client struct {
	*slack.Client
	socketClient *socketmode.Client
	cfg          Config
}

func NewClient(cfg Config, slackOptions ...slack.Option) *Client {
	slackOptions = append(slackOptions, slack.OptionAppLevelToken(cfg.AppToken))
	if cfg.SlackAPIUrl != "" {
		slackOptions = append(slackOptions, slack.OptionAPIURL(cfg.SlackAPIUrl))
	}
	api := slack.New(cfg.ClientSecret, slackOptions...)
	socketClient := socketmode.New(api)
	return &Client{
		Client:       api,
		socketClient: socketClient,
		cfg:          cfg,
	}
}

func (c *Client) GetSelfID() (string, error) {
	res, err := c.AuthTest()
	if err != nil {
		return "", fmt.Errorf("auth test: %w", err)
	}
	return res.UserID, nil
}

func (c *Client) SendMessage(receiver, event, messageID string) (string, error) {
	options := []slack.MsgOption{slack.MsgOptionText(event, false)}
	if messageID != "" {
		options = append(options, slack.MsgOptionTS(messageID))
	}
	_, ts, err := c.PostMessage(receiver, options...)
	if err != nil {
		return "", fmt.Errorf("posting message: %w", err)
	}
	return ts, nil
}

func (c *Client) EditMessage(receiver, event, messageID string) error {
	if messageID == "" {
		return fmt.Errorf("empty message ID")
	}
	options := []slack.MsgOption{slack.MsgOptionText(event, false)}
	_, _, _, err := c.UpdateMessage(receiver, messageID, options...)
	if err != nil {
		return fmt.Errorf("editing message %s: %w", messageID, err)
	}
	return nil
}

func (c *Client) AddReaction(receiver, messageID, reaction string) error {
	if err := c.Client.AddReaction(reaction, slack.ItemRef{
		Channel:   receiver,
		Timestamp: messageID,
	}); err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}
	return nil
}

func (c *Client) ServiceBot(serviceHandler *service.Service) *SlackBot {
	sb := &SlackBot{
		Client:           c.Client,
		socketClient:     c.socketClient,
		mentionsWorkers:  c.cfg.MaxConcurrentMentions,
		linksWorkers:     c.cfg.MaxConcurrentLinks,
		reactionsWorkers: c.cfg.MaxConcurrentReactions,
		mentionsCh:       make(chan *slackevents.AppMentionEvent),
		linksCh:          make(chan *slackevents.LinkSharedEvent),
		reactionsAddCh:   make(chan *slackevents.ReactionAddedEvent),
		adminsUserIds:    make(map[string]interface{}),
		service:          serviceHandler,
	}
	for _, userID := range c.cfg.AdminSlackUserID {
		sb.adminsUserIds[userID] = nil
	}
	return sb
}
