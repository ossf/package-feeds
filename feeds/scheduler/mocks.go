package scheduler

import (
	"context"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

type mockFeed struct {
	packages []*feeds.Package
	errs     []error
	options  feeds.FeedOptions
}

func (feed mockFeed) GetName() string {
	return "mockFeed"
}

func (feed mockFeed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}

func (feed mockFeed) Latest(cutoff time.Time) ([]*feeds.Package, []error) {
	return feed.packages, feed.errs
}

type mockPublisher struct {
	sendCallback func(string) error
}

func (pub mockPublisher) Send(ctx context.Context, body []byte) error {
	if pub.sendCallback != nil {
		return pub.sendCallback(string(body))
	}
	return nil
}

func (pub mockPublisher) Name() string {
	return "mockPublisher"
}
