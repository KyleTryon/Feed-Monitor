package main

import (
	"flag"
	"fmt"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
	"regexp"
	"net/http"
	"os"
	"net/url"
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
}

type FeedActive struct {
	Feed
	lastPulled   int64
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

var ActiveFeeds []FeedActive
var fp = gofeed.NewParser()

func main() {
	fileName := flag.String("config", "", "Config file")
	flag.Parse()
	fmt.Println("Reading config file: ", *fileName)
	if *fileName == "" {
		fmt.Println("No config file specified")
		return
	}
	parseConfig(*fileName)
	runService()
}

func runService() {
	fmt.Println("Starting service")
	for true {
		for i, feed := range ActiveFeeds {
			if time.Now().Unix()-feed.lastPulled > int64(feed.Interval) {
				fmt.Println("Last updated: ", time.Unix(feed.lastPulled, 0))
				checkFeed(feed, i)
			}
		}
	}
}

func parseConfig(fileName string) {
	var config MonitorSchema
	configFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		return
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("Error parsing config file: ", err)
		return
	}
	for _, feed := range config.Monitor {
		ActiveFeeds = append(ActiveFeeds, FeedActive{feed.Feed, 0, ""})
	}
}

func checkFeed(feedItem FeedActive, FeedIndex int) {
	// Fetches the feed and grabs only the new items, stored in newItems[].
	feed, err := fp.ParseURL(feedItem.Url)
	fmt.Println("Checking feed: ", feedItem.Name)
	var newItems []*gofeed.Item
	var matchedItems []*gofeed.Item
	if err != nil {
		fmt.Println("Error parsing feed: ", err)
		return
	}
	fmt.Println("Feed length: ", len(feed.Items))
	for _, item := range feed.Items {
		fmt.Println("Checking item: ", item.Title)
		if item.Title == ActiveFeeds[FeedIndex].lastItemTitle {
			fmt.Println("Item already processed")
			break
		} else {
			fmt.Println("Adding item to new items")
			newItems = append(newItems, item)
		}
	}
	if len(newItems) != 0 {
		ActiveFeeds[FeedIndex].lastItemTitle = newItems[0].Title
	}
	fmt.Println("New items: ", len(newItems))
	for _, item := range newItems {
		isMatchItem := true
		for _, filter := range feedItem.Filters {
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
	if len(matchedItems) != 0 {
		fmt.Println("New items found: ", len(matchedItems))
		for _, item := range matchedItems {
			notificationMessage := fmt.Sprintf("%s: %s", item.Title, item.Link)
			notificationTitle := fmt.Sprintf("New item in %s", feedItem.Name)
			postURL := fmt.Sprintf("%s/message?token=%s", os.Getenv("FEED_MONITOR_POST_URL"), os.Getenv("FEED_MONITOR_POST_TOKEN") )
			resp, err := http.PostForm(postURL,
			url.Values{"message": {notificationMessage}, "title": {notificationTitle}})
			if err != nil {
				fmt.Println("Error posting notification: ", err)
			}
			fmt.Println("Notification posted: ", resp.Status)
			// Send Alert
			println("\n==============================")
			println("Title: ", item.Title)
			println("Description: ", item.Description)
			println("Link: ", item.Link)
			println("==============================")
		}
	}
	ActiveFeeds[FeedIndex].lastPulled = time.Now().Unix()
}

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
