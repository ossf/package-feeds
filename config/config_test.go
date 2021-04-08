package config_test

import (
	"context"
	"testing"

	"github.com/ossf/package-feeds/config"
	"github.com/ossf/package-feeds/feeds/scheduler"
	"github.com/ossf/package-feeds/publisher/stdout"

	"github.com/ossf/package-feeds/feeds/pypi_critical"
)

const (
	TestConfigStr = `
enabled_feeds:
- rubygems
- goproxy
- npm

publisher:
  type: "gcp"
  config:
    endpoint: "https://foobaz.com"

http_port: 8080
`
	TestConfigStrUnknownFeedType = `
enabled_feeds:
- foo
`
	TestConfigStrUnknownField = `
foo:
- bar
- baz
`
)

func TestDefault(t *testing.T) {
	t.Parallel()

	c := config.Default()
	feeds, err := c.GetScheduledFeeds()
	if err != nil {
		t.Fatalf("failed to initialize feeds: %v", err)
	}
	_ = scheduler.New(feeds)
}

func TestGetScheduledFeeds(t *testing.T) {
	t.Parallel()

	c, err := config.NewConfigFromBytes([]byte(TestConfigStr))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.EnabledFeeds) != 3 {
		t.Fatalf("EnabledFeeds is expected to be 3 but was `%v`", len(c.EnabledFeeds))
	}
	feeds, err := c.GetScheduledFeeds()
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range c.EnabledFeeds {
		if _, ok := feeds[val]; !ok {
			t.Errorf("expected `%v` feed was not found in scheduled feeds after GetScheduledFeeds()", val)
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

func TestTypeConfigPairToPublisherStdout(t *testing.T) {
	t.Parallel()

	c := config.TypeConfigPair{
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

func TestTypeConfigPairToFeed(t *testing.T) {
	t.Parallel()

	packages := []string{
		"foo",
		"bar",
		"baz",
	}

	c := config.TypeConfigPair{
		Type: pypi_critical.FeedName,
		Config: map[string]interface{}{
			"packages": packages,
		},
	}
	feed, err := c.ToFeed()
	if err != nil {
		t.Fatalf("failed to create pypi critical feed from configuration: %v", err)
	}

	pypiFeed, ok := feed.(*pypi_critical.Feed)
	if !ok {
		t.Fatal("failed to cast feed as pypi critical feed")
	}

	if len(pypiFeed.PackageList) != len(packages) {
		t.Errorf("pypi critical package list does not match config provided package list")
	} else {
		for i := 0; i < len(packages); i++ {
			if pypiFeed.PackageList[i] != packages[i] {
				t.Errorf("pypi critical package '%v' does not match configured package '%v'", pypiFeed.PackageList[i], packages[i])
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
