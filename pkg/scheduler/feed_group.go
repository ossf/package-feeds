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

type FeedGroup struct {
	feeds     []feeds.ScheduledFeed
	publisher publisher.Publisher
	lastPoll  time.Time
}

type groupResult struct {
	numPublished int
	pollErr      error
	pubErr       error
}

func NewFeedGroup(scheduledFeeds []feeds.ScheduledFeed,
	pub publisher.Publisher, initialCutoff time.Duration) *FeedGroup {
	return &FeedGroup{
		feeds:     scheduledFeeds,
		publisher: pub,
		lastPoll:  time.Now().UTC().Add(-initialCutoff),
	}
}

func (fg *FeedGroup) AddFeed(feed feeds.ScheduledFeed) {
	fg.feeds = append(fg.feeds, feed)
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
	numPublished, pubErr := fg.publishPackages(pkgs)
	result.numPublished = numPublished
	if pubErr != nil {
		log.Errorf("Failed to publish %v packages due to err: %v", len(pkgs)-numPublished, pubErr)
		result.pubErr = errPub
	} else {
		log.WithField("num_packages", numPublished).Printf("Successfully published packages")
	}
	return result
}

// Poll fetches the latest packages from each registered feed.
func (fg *FeedGroup) poll() ([]*feeds.Package, error) {
	results := make(chan pollResult, len(fg.feeds))
	for _, feed := range fg.feeds {
		go func(feed feeds.ScheduledFeed) {
			result := pollResult{
				name: feed.GetName(),
				feed: feed,
			}
			result.packages, result.errs = feed.Latest(fg.lastPoll)
			results <- result
		}(feed)
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
	fg.lastPoll = time.Now().UTC()

	log.Printf("%d packages processed", len(packages))
	return packages, err
}

func (fg *FeedGroup) publishPackages(pkgs []*feeds.Package) (int, error) {
	processed := 0
	for _, pkg := range pkgs {
		log.WithFields(log.Fields{
			"name":         pkg.Name,
			"feed":         pkg.Type,
			"created_date": pkg.CreatedDate,
		}).Print("Sending package upstream")
		b, err := json.Marshal(pkg)
		if err != nil {
			log.Printf("Error marshaling package: %#v", pkg)
			return processed, err
		}
		if err := (fg.publisher).Send(context.Background(), b); err != nil {
			log.Printf("Error sending package to upstream publisher %v", err)
			return processed, err
		}
		processed++
	}
	return processed, nil
}
