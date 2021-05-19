package scheduler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/publisher"
)

// Scheduler is a registry of feeds that should be run on a schedule.
type Scheduler struct {
	registry  map[string]feeds.ScheduledFeed
	publisher publisher.Publisher
	httpPort  int
}

// New returns a new Scheduler with a publisher and feeds configured for polling.
func New(feedsMap map[string]feeds.ScheduledFeed, pub publisher.Publisher, httpPort int) *Scheduler {
	return &Scheduler{
		registry:  feedsMap,
		publisher: pub,
		httpPort:  httpPort,
	}
}

type pollResult struct {
	name     string
	feed     feeds.ScheduledFeed
	packages []*feeds.Package
	err      error
}

// Runs several services for the operation of scheduler, this call is blocking until application exit
// or failure in the HTTP server
// Services include: Cron polling via FeedGroups, HTTP serving of FeedGroupsHandler.
func (s *Scheduler) Run(initialCutoff time.Duration, enableDefaultTimer bool) error {
	defaultSchedule := fmt.Sprintf("@every %s", initialCutoff.String())

	allFeeds := []feeds.ScheduledFeed{}
	for _, feed := range s.registry {
		allFeeds = append(allFeeds, feed)
	}

	// Configure cron job for scheduled polling
	cronJob := cron.New()
	feedGroup := NewFeedGroup(allFeeds, s.publisher, initialCutoff)
	if enableDefaultTimer {
		err := cronJob.AddJob(defaultSchedule, feedGroup)
		if err != nil {
			return fmt.Errorf("failed to parse schedule `%s`: %w", defaultSchedule, err)
		}
		log.Printf("Running a timer %s", defaultSchedule)
	}
	cronJob.Start()

	// Start http server for polling via HTTP requests
	pollServer := NewFeedGroupsHandler([]*FeedGroup{feedGroup})
	log.Infof("Listening on port %v\n", s.httpPort)
	http.Handle("/", pollServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", s.httpPort), nil); err != nil {
		return err
	}

	return nil
}
