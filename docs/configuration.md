# Configuration
Bolt is configured using environment variables

## Required Configuration
* `SLACK_APP_TOKEN` - App-level token for the Slack app (starts with `xapp-`). Used to establish the Socket Mode WebSocket connection.
* `SLACK_OAUTH_TOKEN` - OAuth token of the installed Slack app in a workspace (starts with `xoxb-`).

## Optional Configuration
* `DONT_JOIN_AFTER` - If defined, Bolt won't join orders after that time. Time is defined in HH:MM format. Default is None (will always join).
* `DONT_JOIN_AFTER_TZ` - Timezone for the hour defined in `DONT_JOIN_AFTER`. For example: `Europe/London`. Default is none (will be the local time where Bolt is running). 
* `ORDER_READY_TIMEOUT` - Timeout for waiting for the Wolt group order to be sent, in duration format (e.g. `1m`/`1h`). After that duration, Bolt will stop tracking that order. Default is `1h`.
* `ORDER_DONE_TIMEOUT` - Timeout for waiting for the Wolt group order to be delivered after payment. After that duration, Bolt will stop tracking that order. Default is `3h`.
* `TIME_TILL_GET_READY_MESSAGE` - How long before the delivery ETA the "get ready" message is sent. Default is `7m`.
* `ORDER_DESTINATION_EMOJI` - Emoji used to represent the order's destination in the progress message. Default is `:house:`.
* `JOINED_ORDER_EMOJI` - Emoji Bolt adds to the link message once it has joined the order. Default is `:eyes:`.
* `DEBT_REMINDER_INTERVAL` - Time to wait between each reminder of unpaid debt, in duration format. Default is `3h`.
* `DEBT_MAXIMUM_DURATION` - Maximum duration for which unpaid debt reminders are sent. After this time no more reminders are sent. Default is `24h`.
* `WAIT_BETWEEN_STATUS_CHECK` - Duration between polling for Wolt order status. Default is `20s`.
* `ADMIN_SLACK_USER_IDS` - Comma-separated list of Slack user IDs treated as Bolt admins who can add custom user mappings using the `/add-user` slash command.
* `SLACK_MAX_CONCURRENT_LINKS` - Maximum concurrent Slack link-shared event handlers. A Wolt group link holds a handler until the group finishes. Default is `100`.
* `SLACK_MAX_CONCURRENT_MENTIONS` - Maximum concurrent Slack mention handlers. Default is `100`.
* `SLACK_MAX_CONCURRENT_REACTIONS` - Maximum concurrent Slack reaction handlers. Default is `100`.
* `SLACK_STORE_MAX_CACHE_ENTRY_TIME` - Cache timeout for Wolt-name-to-Slack-user lookups, in duration format. Default is `144h` (6 days).
