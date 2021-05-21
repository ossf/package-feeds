package scheduler

import (
	"fmt"
	"net/http"
	"strings"
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
	errs     []error
}

// Runs several services for the operation of scheduler, this call is blocking until application exit
// or failure in the HTTP server
// Services include: Cron polling via FeedGroups, HTTP serving of FeedGroupsHandler.
func (s *Scheduler) Run(initialCutoff time.Duration, enableDefaultTimer bool) error {
	defaultSchedule := fmt.Sprintf("@every %s", initialCutoff.String())

	schedules, err := buildSchedules(s.registry, s.publisher, initialCutoff)
	if err != nil {
		return err
	}
	feedGroups := []*FeedGroup{}

	// Configure cron job for scheduled polling.
	cronJob := cron.New()
	for schedule, feedGroup := range schedules {
		feedGroups = append(feedGroups, feedGroup)

		// Undefined schedules will follow the default schedule, if the default timer is enabled.
		if schedule == "" {
			if !enableDefaultTimer {
				continue
			}
			schedule = defaultSchedule
		}

		err := cronJob.AddJob(schedule, feedGroup)
		if err != nil {
			return fmt.Errorf("failed to parse schedule `%s`: %w", schedule, err)
		}

		feedNames := []string{}
		for _, f := range feedGroup.feeds {
			feedNames = append(feedNames, f.GetName())
		}
		log.Printf("Running a timer for %s with schedule %s", strings.Join(feedNames, ", "), schedule)
	}
	cronJob.Start()

	// Start http server for polling via HTTP requests
	pollServer := NewFeedGroupsHandler(feedGroups)
	log.Infof("Listening on port %v\n", s.httpPort)
	http.Handle("/", pollServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", s.httpPort), nil); err != nil {
		return err
	}

	return nil
}

// Prepares a map of FeedGroups indexed by their appropriate cron schedule
// The resulting map may have index "" with a FeedGroup of feeds without a schedule option configured.
func buildSchedules(registry map[string]feeds.ScheduledFeed, pub publisher.Publisher,
	initialCutoff time.Duration) (map[string]*FeedGroup, error) {
	schedules := map[string]*FeedGroup{}
	for _, feed := range registry {
		options := feed.GetFeedOptions()

		pollRate := options.PollRate
		cutoff := initialCutoff
		var err error
		var schedule string

		if pollRate != "" {
			cutoff, err = time.ParseDuration(pollRate)
			if err != nil {
				return nil, fmt.Errorf("failed to parse `%s` as duration: %w", pollRate, err)
			}
			schedule = fmt.Sprintf("@every %s", pollRate)
		}

		// Initialize new schedules in map.
		if _, ok := schedules[schedule]; !ok {
			schedules[schedule] = NewFeedGroup([]feeds.ScheduledFeed{}, pub, cutoff)
		}
		schedules[schedule].AddFeed(feed)
	}
	return schedules, nil
}
