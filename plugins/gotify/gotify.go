package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v3"
)

type gotifyConfig struct {
	Url string `yaml:"url"`
	Token string `yaml:"token"`
}

func getParams(config string) (string, string, error) {
	var gotifyConfig gotifyConfig
	err := yaml.Unmarshal([]byte(config), &gotifyConfig)
	if err != nil {
		return "", "", err
	}
	return gotifyConfig.Url, gotifyConfig.Token, nil
}

// A successful status will return "200 OK"
func Send(item *gofeed.Item, feedName string, config string) (error) {
	fmt.Println("Sending to Gotify")
	notificationMessage := fmt.Sprintf("%s: %s", item.Title, item.Link)
	notificationTitle := fmt.Sprintf("New item in %s", feedName)
	gotifyUrl, gotifyToken, err := getParams(config)
	if err != nil {
		return err
	}
	if len(gotifyUrl) > 0 && len(gotifyToken) > 0 {
		postURL := fmt.Sprintf("%s/message?token=%s", gotifyUrl, gotifyToken)
		_, err := http.PostForm(postURL,
			url.Values{"message": {notificationMessage}, "title": {notificationTitle}})
		if err != nil {
			return err
		}
		return err
	} else {
		return errors.New("environment variable(s) missing, ensure GOTIFY_URL and GOTIFY_TOKEN are set")
	}
}

