# Hacking Feed-Monitor

A developer guide to extending the Feed-Monitor service.

## Plugin Development

Feed Monitor is configured to send notification alerts to external plugins that are written in Go. These plugins are responsible for connecting feed alerts to external services such as [Gotify](https://gotify.io). Anyone can develop a plugin for Feed Monitor, and simply place the `plugin.so` file in the `plugins` directory, like so `./plugins/my-plugin/my-plugin.so`.

### Plugin Template

To create your own plugin, create a new Go  module with the following structure:

```go
package main

import (
 "github.com/mmcdole/gofeed"
 "gopkg.in/yaml.v3"
)

// Plugins can accept custom configuration YAML.
type pluginConfig struct {
 Url string `yaml:"url"`
 Token string `yaml:"token"`
}

// Parse the configuration and return the config values.
// You may need the user to provide values via config for the Send function
func getParams(config string) (string, string, error) {
 var pluginConfig pluginConfig
 err := yaml.Unmarshal([]byte(config), &pluginConfig)
 if err != nil {
  return "", "", err
 }
 return pluginConfig.Url, pluginConfig.Token, nil
}

// The Send function must expect a feed item, the feed name, and the plugin configuration.
func Send(item *gofeed.Item, feedName string, config string) (string, error) {
    url, token, err := getParams(config)
    if err != nil {
        return "", err
    }
  return fmt.Sprintf("%s: %s", item.Title, item.Link), nil
}
```
