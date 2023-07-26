package scheduler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/publisher"
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
	var feedGroups []*FeedGroup
	var pollFeedNames []string

	// Configure cron job for scheduled polling.
	cronJob := cron.New(
		cron.WithLogger(cron.PrintfLogger(log.StandardLogger())),
		cron.WithParser(cron.NewParser(
			cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
		)))
	for schedule, feedGroup := range schedules {
		var feedNames []string
		for _, f := range feedGroup.feeds {
			feedNames = append(feedNames, f.GetName())
		}

		if schedule == "" {
			if !enableDefaultTimer {
				// Without the default timer enabled, undefined schedules depend on HTTP request polling.
				// This avoids race conditions where the cron based request is in flight when an HTTP
				// request is made (or visa-versa).
				feedGroups = append(feedGroups, feedGroup)
				pollFeedNames = append(pollFeedNames, feedNames...)
				continue
			}
			// Undefined schedules will follow the default schedule, if the default timer is enabled.
			schedule = defaultSchedule
		}

		_, err := cronJob.AddJob(schedule, cron.NewChain(
			cron.SkipIfStillRunning(cron.VerbosePrintfLogger(log.StandardLogger())),
		).Then(feedGroup))
		if err != nil {
			return fmt.Errorf("failed to parse schedule `%s`: %w", schedule, err)
		}

		log.Printf("Running a timer for %s with schedule %s", strings.Join(feedNames, ", "), schedule)
	}
	cronJob.Start()

	// Start http server for polling via HTTP requests
	pollServer := NewFeedGroupsHandler(feedGroups)
	log.Infof("Listening on port %v for %s", s.httpPort, strings.Join(pollFeedNames, ", "))
	http.Handle("/", pollServer)

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", s.httpPort),
		// default 60s timeout used from nginx
		// https://medium.com/a-journey-with-go/go-understand-and-mitigate-slowloris-attack-711c1b1403f6
		ReadHeaderTimeout: 60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// buildSchedules prepares a map of FeedGroups indexed by their appropriate cron schedule
// The resulting map may have index "" with a FeedGroup of feeds without a schedule option configured.
//
//nolint:lll
func buildSchedules(registry map[string]feeds.ScheduledFeed, pub publisher.Publisher, initialCutoff time.Duration) (map[string]*FeedGroup, error) {
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
