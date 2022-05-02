# Feed-Monitor

<p align="center">
  <img src=".github/img/feed-monitor-logo-color.svg" width="100" height="100">
</p>

_**Note:** This repo is not yet ready for use and is in very early stages of development._

- [Feed-Monitor](#feed-monitor)
  * [Basic Usage](#basic-usage)
    + [Run in Docker](#run-in-docker)
    + [Run in Docker-Compose](#run-in-docker-compose)
  * [Plugins](#plugins)
    + [Official Plugins](#official-plugins)
    + [Community Plugins](#community-plugins)

A configurable service written in Go to monitors RSS, Atom, and JSON feeds. Create a simple configuration file to define your feeds and associated filters, and begin receiving notifications when matches are found. Connect your notification to many services through plugins.

## Basic Usage

Begin by defining a configuration file, multiple examples can be found in the [sample_recipes](./sample_recipes/) directory.

Here is a sample config, which will monitor the NASA "breaking news" RSS feed and alert us to potential articles of interest involving aliens, based on the feed item title and excluding articles posted on April first.

```yaml
monitor:
  - feed:
      name: 'Aliens are real'
      url: https://www.nasa.gov/rss/dyn/breaking_news.rss
      interval: 60
      filters:
        - include:
            element: title
            matches: '/alien[s]?.*[life|real]/i'
        - exclude:
            element: published
            matches: '/Apr, 1/'
      notifiers:
        - gotify:
            server: https://gotify.example.com
            token: 'your_token'
```

All filters must return true for the item to be considered a match. The example above states the title must contain the word "alien" and "life" or "real" to be considered a match _and_ the published date must not contain "April 1".

```shell
go run main.go -config=config.yml
```

### Run in Docker

> Not yet published

```shell
docker run \
  --name-feed-monitor \
  -v /host/path/to/config:/config `# path to config file` \
  --restart unless-stopped \
  techsquidtv/feed-monitor:latest
```

### Run in Docker-Compose

> Not yet published

```yaml
services:
  feed-monitor:
    image: techsquidtv/feed-monitor:latest
    restart: unless-stopped
    volumes:
      - /host/path/to/config:/config
```

## Plugins

Feed Monitor supports plugins to push notifications to other services.

### Official Plugins

| Plugin | Description |
| --- | --- |
| [gotify](./plugins/gotify/) | Send notifications to a [gotify](https://gotify.io) server to receive real-time push notifications when a match is found. |

### Community Plugins

Write a plugin for Feed-Monitor! Get started by reading the  [HACKING.md](./HACKING.md) page.

| Plugin | Description |
| --- | --- |
| - awaiting - | - |
