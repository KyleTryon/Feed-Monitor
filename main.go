package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
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
}

type FeedActive struct {
	Feed
	lastPulled   int64
	lastItemGUID string
}

type MonitorSchema struct {
	Monitor []struct {
		Feed Feed `yaml:"feed"`
	} `yaml:"monitor"`
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
	var wg sync.WaitGroup
	wg.Add(len(ActiveFeeds))
	for _, feed := range ActiveFeeds {
		go func(feed FeedActive) {
			defer wg.Done()
			fmt.Println("Checking feed: ", feed.Name)
			time.Sleep(time.Duration(feed.Interval) * time.Second)
			checkFeed(&feed)
		}(feed)
	}
	wg.Wait()
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

func checkFeed(feedItem *FeedActive) {
	// Fetches the feed and grabs only the new items, stored in newItems[].
	feed, err := fp.ParseURL(feedItem.Url)
	fmt.Println("Last checked item: ", feedItem.lastItemGUID)
	var newItems []*gofeed.Item
	if err != nil {
		fmt.Println("Error parsing feed: ", err)
		return
	}
	for _, item := range feed.Items {
		// fmt.Println("Checking item: ", item.Title)
		if item.GUID == feedItem.lastItemGUID {
			fmt.Println("End of new items")
			break
		}
		newItems = append(newItems, item)
	}
	if len(newItems) != 0 {
		feedItem.lastItemGUID = newItems[0].GUID
	}
	fmt.Println("Done checking feed: ", feedItem.Name)
}

func checkFilter(item *gofeed.Item, filter FeedFilter) bool {
	switch filter.Element {
	case "title":
		return item.Title == filter.Matches
	case "link":
		return item.Link == filter.Matches
	case "published":
		return item.Published == filter.Matches
	}
	return false
}