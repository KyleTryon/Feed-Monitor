# Hacking Feed-Monitor

A developer guide to extending the Feed-Monitor service.

## Plugin Development

Feed Monitor is configured to send notification alerts to external plugins that are written in Go. These plugins are responsible for connecting feed alerts to external services such as [Gotify](https://gotify.io). Anyone can develop a plugin for Feed Monitor, and simply place the `plugin.so` file in the `plugins` directory, like so `./plugins/my-plugin/my-plugin.so`.

### Plugin Template

To create your own plugin, create a new Go module with the following structure:

```go
package main

import (
 "github.com/mmcdole/gofeed"
)

// The Send function must expect a feed item, the feed name, and the plugin configuration.
func Send(item *gofeed.Item, feedName string, config map[string]string) (error) {
  // Access config via the map, such as: config["server"]
  // Get data from the item, such as: item.Title
  return nil
}
```

A plugin must implement the `Send` function. The `Send` function is responsible for sending the alert to the external service. The `Send` function should return an error if the alert could not be sent.
