package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"github.com/mmcdole/gofeed"
)

func Send(item *gofeed.Item, feedName string, config map[string]string) (error) {
	notificationMessage := fmt.Sprintf("%s: %s", item.Title, item.Link)
	notificationTitle := fmt.Sprintf("New item in %s", feedName)
	if len(config["url"]) > 0 && len(config["token"]) > 0 {
		postURL := fmt.Sprintf("%s/message?token=%s", config["url"], config["token"])
		response, err := http.PostForm(postURL,
			url.Values{"message": {notificationMessage}, "title": {notificationTitle}})
		if err != nil {
			return err
		}
		fmt.Println("Gotify response: ", response.Status)
		return err
	} else {
		return errors.New("gotify: Token or URL is not set")
	}
}
