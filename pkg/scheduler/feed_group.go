package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/publisher"
)

var (
	errPoll = errors.New("error when polling for packages")
	errPub  = errors.New("error when publishing packages")
)

type feedEntry struct {
	feed     feeds.ScheduledFeed
	lastPoll time.Time
}

type FeedGroup struct {
	feeds         []*feedEntry
	publisher     publisher.Publisher
	initialCutoff time.Time
}

type groupResult struct {
	numPublished int
	pollErr      error
	pubErr       error
}

//nolint:lll
func NewFeedGroup(scheduledFeeds []feeds.ScheduledFeed, pub publisher.Publisher, initialCutoff time.Duration) *FeedGroup {
	fg := &FeedGroup{
		publisher:     pub,
		initialCutoff: time.Now().UTC().Add(-initialCutoff),
		feeds:         make([]*feedEntry, 0),
	}
	for _, feed := range scheduledFeeds {
		fg.AddFeed(feed)
	}
	return fg
}

func (fg *FeedGroup) AddFeed(feed feeds.ScheduledFeed) {
	fg.feeds = append(fg.feeds, &feedEntry{
		feed:     feed,
		lastPoll: fg.initialCutoff,
	})
}

func (fg *FeedGroup) Run() {
	result := fg.pollAndPublish()
	if result.pollErr != nil {
		log.Error(result.pollErr)
	}
	if result.pubErr != nil {
		log.Error(result.pubErr)
	}
}

func (fg *FeedGroup) pollAndPublish() groupResult {
	result := groupResult{}
	pkgs, err := fg.poll()
	result.pollErr = err
	// Return early if no packages to process
	if len(pkgs) == 0 {
		return result
	}
	log.WithField("num_packages", len(pkgs)).Printf("Publishing packages...")
	result.numPublished, result.pubErr = fg.publishPackages(pkgs)
	if result.numPublished > 0 {
		log.WithField("num_packages", result.numPublished).Printf("Successfully published packages")
	}
	return result
}

// Poll fetches the latest packages from each registered feed.
func (fg *FeedGroup) poll() ([]*feeds.Package, error) {
	results := make(chan pollResult, len(fg.feeds))
	for _, f := range fg.feeds {
		go func(f *feedEntry) {
			result := pollResult{
				name: f.feed.GetName(),
				feed: f.feed,
			}
			result.packages, f.lastPoll, result.errs = f.feed.Latest(f.lastPoll)
			results <- result
		}(f)
	}
	errs := []error{}
	packages := []*feeds.Package{}
	for i := 0; i < len(fg.feeds); i++ {
		result := <-results

		logger := log.WithField("feed", result.name)
		for _, err := range result.errs {
			logger.WithError(err).Error("Error fetching packages")
			errs = append(errs, err)
		}
		for _, pkg := range result.packages {
			log.WithFields(log.Fields{
				"feed":    result.name,
				"name":    pkg.Name,
				"version": pkg.Version,
			}).Print("Processing Package")
		}
		packages = append(packages, result.packages...)
		logger.WithField("num_processed", len(result.packages)).Print("Packages successfully processed")
	}
	err := errPoll
	if len(errs) == 0 {
		err = nil
	}

	log.WithField("time", time.Now().UTC()).Printf("%d packages processed", len(packages))
	return packages, err
}

func (fg *FeedGroup) publishPackages(pkgs []*feeds.Package) (int, error) {
	processed := 0
	errs := []error{}
	for _, pkg := range pkgs {
		log.WithFields(log.Fields{
			"name":         pkg.Name,
			"feed":         pkg.Type,
			"created_date": pkg.CreatedDate,
		}).Print("Sending package upstream")
		b, err := json.Marshal(pkg)
		if err != nil {
			log.WithField("name", pkg.Name).WithError(err).Error("Error marshaling package")
			errs = append(errs, err)
		}
		if err := (fg.publisher).Send(context.Background(), b); err != nil {
			log.WithField("name", pkg.Name).WithError(err).Error("Error sending package to upstream publisher")
			errs = append(errs, err)
		}
		processed++
	}
	err := errPub
	if len(errs) == 0 {
		err = nil
	}
	if len(pkgs)-processed != 0 {
		log.Errorf("Failed to publish %v packages", len(pkgs)-processed)
	}
	return processed, err
}
