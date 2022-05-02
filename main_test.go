package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
	"os"
	"testing"
)

var exampleFeed = Feed{
	Name:     "Example",
	Url:      "https://example.com/feed.xml",
	Interval: 60,
	Filters: make([]struct {
		Include FeedFilter
		Exclude FeedFilter
	}, 0),
	Notifiers: make([]map[string]map[string]string, 0),
}

var exampleFilter = struct {
	Include FeedFilter
	Exclude FeedFilter
}{
	Include: FeedFilter{
		Element: "item",
		Matches: ".*",
	},
}

var exampleMatchItem = &gofeed.Item{
	Title:       "Example Title",
	Description: "Example Description",
	Link:        "https://example.com/",
	Links: []string{
		"https://example.com/",
	},
	Content: "Example Content",
	Updated: "Example Updated",
}

func setupTests() {
	godotenv.Load(".env")
}

func TestLoadPlugins(t *testing.T) {
	plugins, err := loadPlugins()
	if err != nil {
		t.Error("Error loading plugins: ", err)
	}
	if len(plugins) == 0 {
		t.Error("No plugins loaded")
	}
}

func TestSendFeedNotifications(t *testing.T) {

	setupTests()

	var exampleNotifier = map[string]map[string]string{
		"gotify": {
			"url":   os.Getenv("GOTIFY_URL"),
			"token": os.Getenv("GOTIFY_TOKEN"),
		},
	}
	exampleFeed.Filters = append(exampleFeed.Filters, exampleFilter)
	exampleFeed.Notifiers = append(exampleFeed.Notifiers, exampleNotifier)

	fmt.Println("TEST: Loading plugins")
	plugins, err := loadPlugins()
	if err != nil {
		t.Error("Error loading plugins: ", err)
	}

	status, err := sendFeedNotifications(exampleFeed, exampleMatchItem, plugins)
	if err != nil {
		t.Errorf("Error sending notification, Error: %s", err)
	}
	if status != "ok" {
		t.Errorf("Error sending notification, Status: %s", status)
	}
}
