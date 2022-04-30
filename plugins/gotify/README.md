# Gotify Plugin for Feed-Monitor

This plugin allows `Feed-Monitor` to send notifications to a [Gotify](https://gotify.io) server to receive real-time push notifications when a match is found.

## Usage

Add the following config to your feed-monitor config file:

```yaml
notifiers:
    - gotify:
        server: https://gotify.example.com
        token: 'your_token'
```

You can specify different tokens for different feeds to categorize your notifications.
