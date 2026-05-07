# Creating a Slack app

1. Go to [Slack Apps Dashboard](https://api.slack.com/apps)
2. Click `Create New App`
3. Choose [`From an app manifest`](../assets/slack/3_create.png)
4. Select the workspace where you want to deploy the app
5. Copy the content of the [app manifest](../../deploy/app_manifest.yaml) — no placeholders need replacing
6. Click `Next` and then `Create`
7. Click `Install To Workspace` and then `Allow` when asking for permissions (if you are not managing the workspace you may need admin approval)
8. Go to `Basic Information` → `App-Level Tokens` → `Generate Token and Scopes`.
   Name it anything (e.g. `bolt-socket`), add the `connections:write` scope, and click `Generate`.
   Copy the token (starts with `xapp-`) and save it — this is `SLACK_APP_TOKEN`.
9. Go to `Install App` and copy the [`OAuth Token`](../assets/slack/9_token.png) (starts with `xoxb-`) — this is `SLACK_OAUTH_TOKEN`.
10. You may change the app icon to [Bolt's icon](../assets/bolt_logo_slack.png) — go to `Basic Information` and find the icon under `Display Information`
11. Now configure `SLACK_OAUTH_TOKEN` and `SLACK_APP_TOKEN` for Bolt and run it. [See here how to run it using Kubernetes](k8s.md)
12. [Invite Bolt](../assets/slack/13_add.png) to any channel you want it to join for Wolt food links (use `/add` Slack command)
13. Send a Wolt group link to a channel where Bolt is invited and [see it in action](../assets/slack/14_working.png)

> **Note:** Socket Mode means Bolt opens an outbound WebSocket connection to Slack.
> No inbound port or public IP address is required.
