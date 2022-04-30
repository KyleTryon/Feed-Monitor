package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"time"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v3"
)

type FeedFilter struct {
	Element string `yaml:"element"`
	Matches string `yaml:"matches"`
}

type Feed struct {
	Name     string `yaml:"name"`
	Url      string `yaml:"url"`
	Interval int    `yaml:"interval"`
	Filters  []struct {
		Include FeedFilter
		Exclude FeedFilter
	} `yaml:"filters"`
	Notifiers []map[string]map[string]string `yaml:"notifiers,omitempty"`
}

type FeedActive struct {
	Feed
	lastPulled    int64
	lastItemTitle string
}

type MonitorSchema struct {
	Monitor []struct {
		Feed Feed `yaml:"feed"`
	} `yaml:"monitor"`
}

type Notification struct {
	Send func(*gofeed.Item)
}

type Notifier struct {
	Name string
	Config string
}


func main() {
	fileName := flag.String("config", "", "config.yml")
	mode := flag.String("mode", "dev", "dev or prod")
	flag.Parse()
	fmt.Println("Reading config file: ", *fileName)
	if *fileName == "" {
		fmt.Println("No config file specified")
		return
	}
	if *mode != "prod" {
		err := godotenv.Load(".env")
		if err != nil {
			fmt.Println("Error loading .env file")
		}
	}

	activeFeeds, err := parseConfig(*fileName)
	if err != nil {
		fmt.Println("Error parsing config file: ", err)
		os.Exit(1)
	}

	plugins, err := loadPlugins()
	if err != nil {
		fmt.Println("Error loading plugins: ", err)
		os.Exit(1)
	}

	runService(plugins, activeFeeds)
}

// Loads all plugins in the plugins directory.
// Set OS environment variable 'FEED_MONITOR_PLUGIN_DIR' to the directory containing the plugins, or default to "./plugins"
func loadPlugins() (map[string]plugin.Symbol, error) {
	fmt.Println("Loading plugins...")
	var pluginList []string
	pluginMap := make(map[string]plugin.Symbol)
	pluginPath := os.Getenv("FEED_MONITOR_PLUGIN_DIR")
	if pluginPath == "" {
		pluginPath = "./plugins"
	}
	err := filepath.Walk(pluginPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".so" {
			pluginList = append(pluginList, path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error loading plugins")
		return nil, err
	}
	for _, pluginPath := range pluginList {
		fmt.Println("Loading plugin: ", pluginPath)
		plug, err := plugin.Open(pluginPath)
		if err != nil {
			fmt.Println("Error loading plugin: ", err)
		}
		sendNotification, err := plug.Lookup("Send")
		if err != nil {
			fmt.Println("Error looking up Send function in plugin")
			return nil, err
		}
		pluginMap[pluginPath] = sendNotification
	}
	return pluginMap, nil
}

// The main service loop.
// Checks the feed for new items, and sends notifications if there are any.
func runService(plugins map[string]plugin.Symbol, activeFeeds []FeedActive) {
	fmt.Println("Starting service")
	ActiveFeeds := activeFeeds
	for {
		for i, feed := range ActiveFeeds {
			if time.Now().Unix()-feed.lastPulled > int64(feed.Interval) {
				fmt.Println("Last updated: ", time.Unix(feed.lastPulled, 0))
				matches, err := checkFeed(feed.Feed, feed.lastItemTitle)
				if err != nil {
					fmt.Println("Error checking feed: ", err)
					continue
				}
				if len(matches) > 0 {
					ActiveFeeds[i].lastPulled = time.Now().Unix()
					ActiveFeeds[i].lastItemTitle = matches[len(matches)-1].Title
				}
				for _, match := range matches {
					sendFeedNotifications(feed.Feed, match, plugins)
				}
			}
		}
	}
}

// Parses the YAML config file and returns a list of Feeds.
func parseConfig(fileName string) ([]FeedActive, error) {
	var config MonitorSchema
	configFile, err := ioutil.ReadFile(fileName)
	var activeFeeds []FeedActive
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		return nil, err
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("Error parsing config file: ", err)
		return nil, err
	}
	for _, feed := range config.Monitor {
		activeFeeds = append(activeFeeds, FeedActive{feed.Feed, 0, ""})
	}
	return activeFeeds, nil
}

// Checks a feed and returns any new items
func checkFeed(feed Feed, lastItem string) ([]*gofeed.Item, error){
	var fp = gofeed.NewParser()
	parsedFeed, err := fp.ParseURL(feed.Url)
	fmt.Println("Checking feed: ", feed.Name)
	var newItems []*gofeed.Item
	var matchedItems []*gofeed.Item
	if err != nil {
		fmt.Println("Error parsing feed")
		return nil, err
	}
	fmt.Println("Feed length: ", len(parsedFeed.Items))
	for _, item := range parsedFeed.Items {
		fmt.Println("Checking item: ", item.Title)
		if item.Title == lastItem {
			fmt.Println("Item already processed")
			break
		} else {
			fmt.Println("Adding item to new items")
			newItems = append(newItems, item)
		}
	}
	fmt.Println("New items: ", len(newItems))
	for _, item := range newItems {
		isMatchItem := true
		for _, filter := range feed.Filters {
			if filter.Include.Element != "" {
				if !checkFilter(item, filter.Include, true) {
					isMatchItem = false
					continue
				}
			}
			if filter.Exclude.Element != "" {
				if checkFilter(item, filter.Exclude, false) {
					isMatchItem = false
					continue
				}
			}
		}
		if isMatchItem {
			matchedItems = append(matchedItems, item)
		}
	}
	return matchedItems, nil
}

// Send a notification for a feed item via all supplied "Notifiers"
func sendFeedNotifications(feed Feed, item *gofeed.Item, plugins map[string]plugin.Symbol ) (string, error) {
	if len(plugins) == 0 {
		return "", errors.New("no plugins loaded")
	}
	if len(feed.Notifiers) == 0 {
		return "", errors.New("no notifiers configured")
	}
	for _, y := range feed.Notifiers {
		for k, v := range y {
			selectedPlugin := getPlugin(k)
			if plugins[selectedPlugin] != nil {
				pluginConfig, err := yaml.Marshal(v)
				if err != nil {
					fmt.Println("Error marshalling plugin config: ", err)
					return "", err
				}
				err = plugins[selectedPlugin].(func(*gofeed.Item, string, string) (error))(item, feed.Name, string(pluginConfig))
				if err != nil {
					return "", err
				}
			}
		}
	}
	return "ok", nil
}

// Returns the plugin path from the name.
func getPlugin(pluginName string) (string) {
	pluginPath := os.Getenv("FEED_MONITOR_PLUGIN_DIR")
	if pluginPath == "" {
		pluginPath = "plugins"
	}
	return pluginPath + "/" + pluginName + "/" + pluginName + ".so"
}

// Checks an item against a filter and returns true if it matches.
func checkFilter(item *gofeed.Item, filter FeedFilter, inclusive bool) bool {
	var elementString string
	re := regexp.MustCompile(filter.Matches)
	switch filter.Element {
	case "title":
		elementString = item.Title
	case "link":
		elementString = item.Link
	case "published":
		elementString = item.Published
	case "content":
		elementString = item.Content
	case "author":
		elementString = item.Authors[0].Name
	case "Updated":
		elementString = item.Updated
	case "description":
		elementString = item.Description
	}
	matched := re.Match([]byte(elementString))
	if inclusive {
		return matched
	} else {
		return !matched
	}
}
