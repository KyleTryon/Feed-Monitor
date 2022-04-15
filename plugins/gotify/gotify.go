package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"github.com/mmcdole/gofeed"
)

type notificationInfo func(*gofeed.Item, string)

func Send(item *gofeed.Item, feedName string) (string, error) {
	notificationMessage := fmt.Sprintf("%s: %s", item.Title, item.Link)
	notificationTitle := fmt.Sprintf("New item in %s", feedName)
	if len(os.Getenv("GOTIFY_URL")) > 0 && len(os.Getenv("GOTIFY_TOKEN")) > 0 {
		postURL := fmt.Sprintf("%s/message?token=%s", os.Getenv("GOTIFY_URL"), os.Getenv("GOTIFY_TOKEN"))
		resp, err := http.PostForm(postURL,
			url.Values{"message": {notificationMessage}, "title": {notificationTitle}})
		if err != nil {
			return "Error posting notification", err
		}
		return (resp.Status), err
	} else {
		return ("Environment variable(s) missing, ensure GOTIFY_URL and GOTIFY_TOKEN are set."), errors.New("Environment variable(s) missing, ensure GOTIFY_URL and GOTIFY_TOKEN are set.")
	}
}

