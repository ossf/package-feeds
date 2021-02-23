package scheduler

import (
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/feeds/crates"
	"github.com/ossf/package-feeds/feeds/goproxy"
	"github.com/ossf/package-feeds/feeds/npm"
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
	}
	s := &Scheduler{
		registry: registry,
	}
	return s
}

// Poll fetches the latest packages from each registered feed
func (s *Scheduler) Poll(cutoff time.Time) ([]*feeds.Package, error) {
	packages := []*feeds.Package{}
	for name, feed := range s.registry {
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
