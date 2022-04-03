# Feed-Monitor

_**Note:** This repo is not yet ready for use and is in very early stages of development._

A configurable service written in Go to monitors RSS, Atom, and JSON feeds. Create a simple configuration file to define your feeds and associated filters, and begin receiving notifications when matches are found.

## Basic Usage

Begin by defining a configuration file, multiple examples can be found in the `sample_recipes` directory.

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
```

```shell
$ go run main.go -config=config.yml
```
