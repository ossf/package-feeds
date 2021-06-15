package config_test

import (
	"context"
	"testing"

	"github.com/ossf/package-feeds/pkg/config"
	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/feeds/pypi"
	"github.com/ossf/package-feeds/pkg/publisher/stdout"
	"github.com/ossf/package-feeds/pkg/scheduler"
)

const (
	TestConfigStr = `
feeds:
- type: rubygems
- type: goproxy
- type: npm

publisher:
  type: "gcp"
  config:
    endpoint: "https://foobaz.com"

http_port: 8080
poll_rate: 5m
timer: true
`
	TestConfigStrUnknownFeedType = `
feeds:
- type: foo
`
	TestConfigStrUnknownField = `
foo:
- bar
- baz
`
	TestEventsConfig = `
events:
  sink: stdout
  filter:
    enabled_event_types:
    - foo
    disabled_event_types:
    - bar
    enabled_components:
    - baz
`
)

func TestDefault(t *testing.T) {
	t.Parallel()

	c := config.Default()
	scheduledFeeds, err := c.GetScheduledFeeds()
	if err != nil {
		t.Fatalf("failed to initialize feeds: %v", err)
	}
	pub, err := c.PubConfig.ToPublisher(context.TODO())
	if err != nil {
		t.Fatalf("Failed to initialise publisher from config")
	}
	_ = scheduler.New(scheduledFeeds, pub, c.HTTPPort)
}

func TestGetScheduledFeeds(t *testing.T) {
	t.Parallel()

	c, err := config.NewConfigFromBytes([]byte(TestConfigStr))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Feeds) != 3 {
		t.Fatalf("Feeds is expected to be 3 but was `%v`", len(c.Feeds))
	}
	scheduledFeeds, err := c.GetScheduledFeeds()
	if err != nil {
		t.Fatal(err)
	}
	for _, feed := range c.Feeds {
		if _, ok := scheduledFeeds[feed.Type]; !ok {
			t.Errorf("expected `%v` feed was not found in scheduled feeds after GetScheduledFeeds()", feed.Type)
		}
	}
}

func TestLoadFeedConfigUnknownFeedType(t *testing.T) {
	t.Parallel()

	c, err := config.NewConfigFromBytes([]byte(TestConfigStrUnknownFeedType))
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	_, err = c.GetScheduledFeeds()
	if err == nil {
		t.Error("unknown feed type was successfully parsed when it should've failed")
	}
}

func TestPublisherConfigToPublisherStdout(t *testing.T) {
	t.Parallel()

	c := config.PublisherConfig{
		Type: stdout.PublisherType,
	}
	pub, err := c.ToPublisher(context.TODO())
	if err != nil {
		t.Fatal("failed to create stdout publisher from config")
	}
	if pub.Name() != stdout.PublisherType {
		t.Errorf("stdout sub config produced a publisher with an unexpected name: '%v' != '%v'",
			pub.Name(), stdout.PublisherType)
	}
}

func TestPublisherConfigToFeed(t *testing.T) {
	t.Parallel()

	packages := []string{
		"foo",
		"bar",
		"baz",
	}

	c := config.FeedConfig{
		Type: pypi.FeedName,
		Options: feeds.FeedOptions{
			Packages: &packages,
		},
	}
	feed, err := c.ToFeed(events.NewNullHandler())
	if err != nil {
		t.Fatalf("failed to create pypi feed from configuration: %v", err)
	}

	pypiFeed, ok := feed.(*pypi.Feed)
	if !ok {
		t.Fatal("failed to cast feed as pypi feed")
	}

	feedPackages := pypiFeed.GetPackageList()
	if feedPackages == nil {
		t.Fatalf("failed to initialize pypi feed package list to poll")
	}
	if feedPackages != nil && len(*feedPackages) != len(packages) {
		t.Errorf("pypi package list does not match config provided package list")
	} else {
		for i := 0; i < len(packages); i++ {
			if (*feedPackages)[i] != packages[i] {
				t.Errorf("pypi package '%v' does not match configured package '%v'", (*feedPackages)[i], packages[i])
			}
		}
	}
}

func TestStrictConfigDecoding(t *testing.T) {
	t.Parallel()

	_, err := config.NewConfigFromBytes([]byte(TestConfigStrUnknownField))
	if err == nil {
		t.Fatal("config successfully parsed despite invalid top level configuration field")
	}
}

func TestEventHandlerConfiguration(t *testing.T) {
	t.Parallel()

	c, err := config.NewConfigFromBytes([]byte(TestEventsConfig))
	if err != nil {
		t.Fatalf("failed to load config from bytes: %v", err)
	}

	handler, err := c.GetEventHandler()
	if err != nil || handler == nil {
		t.Fatalf("failed to initialize event handler from config")
	}

	_, ok := handler.GetSink().(*events.LoggingEventSink)
	if !ok {
		t.Fatalf("sink is not configured as stdout as config file expects")
	}

	filter := handler.GetFilter()

	fooEvent := events.MockEvent{
		Type:      "foo",
		Component: "qux",
	}
	barEvent := events.MockEvent{
		Type:      "bar",
		Component: "baz",
	}
	bazEvent := events.MockEvent{
		Type:      "qux",
		Component: "baz",
	}
	if !filter.ShouldDispatch(fooEvent) {
		t.Errorf("configured filter incorrectly rejects type `foo` from being dispatched")
	}
	if filter.ShouldDispatch(barEvent) {
		t.Errorf("configured filter incorrectly allows type `bar` to be dispatched")
	}
	if !filter.ShouldDispatch(bazEvent) {
		t.Errorf("configured filter incorrectly rejects component `baz` from being dispatched")
	}
}
