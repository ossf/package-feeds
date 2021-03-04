package scheduler

import (
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/feeds/crates"
	"github.com/ossf/package-feeds/feeds/goproxy"
	"github.com/ossf/package-feeds/feeds/npm"
	"github.com/ossf/package-feeds/feeds/nuget"
	"github.com/ossf/package-feeds/feeds/pypi"
	"github.com/ossf/package-feeds/feeds/rubygems"
	log "github.com/sirupsen/logrus"
)

// Scheduler is a registry of feeds that should be run on a schedule
type Scheduler struct {
	registry map[string]feeds.ScheduledFeed
}

// New returns a new Scheduler with available feeds registered
func New() *Scheduler {
	registry := map[string]feeds.ScheduledFeed{
		pypi.FeedName:     pypi.Feed{},
		npm.FeedName:      npm.Feed{},
		rubygems.FeedName: rubygems.Feed{},
		crates.FeedName:   crates.Feed{},
		goproxy.FeedName:  goproxy.Feed{},
		nuget.FeedName:    nuget.Feed{},
	}
	s := &Scheduler{
		registry: registry,
	}
	return s
}

type pollResult struct {
	name     string
	feed     feeds.ScheduledFeed
	packages []*feeds.Package
	err      error
}

// Poll fetches the latest packages from each registered feed
func (s *Scheduler) Poll(cutoff time.Time) ([]*feeds.Package, []error) {
	results := make(chan pollResult)
	for name, feed := range s.registry {
		go func(name string, feed feeds.ScheduledFeed) {
			result := pollResult{
				name: name,
				feed: feed,
			}
			result.packages, result.err = feed.Latest(cutoff)
			results <- result
		}(name, feed)
	}
	errs := []error{}
	packages := []*feeds.Package{}
	for i := 0; i < len(s.registry); i++ {
		result := <-results
		logger := log.WithField("feed", result.name)
		if result.err != nil {
			logger.WithError(result.err).Error("error fetching packages")
			errs = append(errs, result.err)
			continue
		}
		packages = append(packages, result.packages...)
		logger.WithField("num_processed", len(result.packages)).Print("processed packages")
	}
	return packages, errs
}
