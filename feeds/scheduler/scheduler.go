package scheduler

import (
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/feeds/crates"
	"github.com/ossf/package-feeds/feeds/npm"
	"github.com/ossf/package-feeds/feeds/pypi"
	"github.com/ossf/package-feeds/feeds/rubygems"
	log "github.com/sirupsen/logrus"
)

// ScheduledFeeds is a registry of feeds that should be run on a schedule
var ScheduledFeeds = make(map[string]feeds.ScheduledFeed)

func init() {
	ScheduledFeeds[pypi.FeedName] = pypi.Feed{}
	ScheduledFeeds[npm.FeedName] = npm.Feed{}
	ScheduledFeeds[rubygems.FeedName] = rubygems.Feed{}
	ScheduledFeeds[crates.FeedName] = crates.Feed{}
}

// PollScheduledFeeds fetches the latest packages from each registered feed
func PollScheduledFeeds(cutoff time.Time) ([]*feeds.Package, error) {
	packages := []*feeds.Package{}
	for name, feed := range ScheduledFeeds {
		logger := log.WithField("feed", name)
		pkgs, err := feed.Latest(cutoff)
		if err != nil {
			logger.WithError(err).Error("error fetching packages")
			return nil, err
		}
		processed := 0
		for _, pkg := range pkgs {
			processed++
			packages = append(packages, pkg)
		}
		logger.WithField("num_processed", processed).Print("processed packages")
	}
	return packages, nil
}
