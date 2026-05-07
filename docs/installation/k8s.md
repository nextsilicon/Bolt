# Run using Kubernetes
If you have a Kubernetes cluster, it's very easy to get Bolt up and running.
No external IP address or inbound firewall rules are needed — Bolt connects to Slack via an outbound WebSocket.

### Prerequisites
* A Kubernetes cluster and configured `kubectl`
* Slack App Token (`xapp-...`) and OAuth Token (`xoxb-...`)

Make sure you created a Slack app and have both tokens. See instructions [here](./slack_app.md).

## Running Bolt
 
### Configuration
You will need at least one Slack user configured as Bolt's admin.
Bolt's admin can map Wolt users to Slack users when Bolt cannot do it automatically (using the `/add-user` slash command).

To get the user ID, go to the user's profile, click the 3 dots, and choose [`Copy member ID`](../assets/slack/0_copy_member_id.png).

1. Go to [deployment.yaml](../../deploy/deployment.yaml)
2. Paste Slack's OAuth token and App-level token **as Base64 encoded** instead of `<slack_oauth_token_base64>` and `<slack_app_token_base64>` respectively
3. Configure `ADMIN_SLACK_USER_IDS` with admin users for Bolt (separated by a comma)
4. Change other configuration as needed in the `ConfigMap` section (see [configuration](../configuration.md) for all available configurations)
5. Apply deployment: `kubectl apply -f deploy/deployment.yaml`
6. Run `kubectl get pods -n bolt` and make sure Bolt is up and running (may take a few minutes)
