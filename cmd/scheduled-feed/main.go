package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/config"
	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/feeds/scheduler"
	"github.com/ossf/package-feeds/publisher"
)

// FeedHandler is a handler that fetches new packages from various feeds.
type FeedHandler struct {
	scheduler *scheduler.Scheduler
	pub       publisher.Publisher
	pollRate  time.Duration
	lastPoll  time.Time
}

func (handler *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	pkgs, pollErrors := handler.pollFeeds()
	processedPackages, err := handler.publishPackages(pkgs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pollErrors {
		http.Error(w, "error polling for packages - see logs for more information", http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf("%d packages processed", processedPackages)))
	if err != nil {
		http.Error(w, "unexpected error during http server write: %w", http.StatusInternalServerError)
	}
}

func (handler FeedHandler) getCutoff() time.Time {
	var cutoff time.Time
	if handler.lastPoll.IsZero() {
		cutoff = time.Now().UTC().Add(-handler.pollRate)
	} else {
		cutoff = handler.lastPoll
	}
	return cutoff
}

func (handler *FeedHandler) pollFeeds() ([]*feeds.Package, bool) {
	cutoff := handler.getCutoff()
	handler.lastPoll = time.Now().UTC()
	pkgs, errs := handler.scheduler.Poll(cutoff)
	errors := false
	if len(errs) > 0 {
		errors = true
		for _, err := range errs {
			log.Errorf("error polling for new packages: %v", err)
		}
	}
	return pkgs, errors
}

func (handler FeedHandler) publishPackages(pkgs []*feeds.Package) (int, error) {
	processed := 0
	for _, pkg := range pkgs {
		log.WithFields(log.Fields{
			"name":         pkg.Name,
			"feed":         pkg.Type,
			"created_date": pkg.CreatedDate,
		}).Print("sending package upstream")
		b, err := json.Marshal(pkg)
		if err != nil {
			log.Printf("error marshaling package: %#v", pkg)
			return processed, err
		}
		if err := handler.pub.Send(context.Background(), b); err != nil {
			log.Printf("error sending package to upstream publisher %v", err)
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func main() {
	configPath, useConfig := os.LookupEnv("PACKAGE_FEEDS_CONFIG_PATH")
	var err error

	var appConfig *config.ScheduledFeedConfig
	if useConfig {
		appConfig, err = config.FromFile(configPath)
		log.Infof("Using config from file: %v", configPath)
	} else {
		appConfig = config.Default()
		log.Info("No config specified, using default configuration")
	}
	if err != nil {
		log.Fatal(err)
	}

	pub, err := appConfig.PubConfig.ToPublisher(context.TODO())
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initialize publisher from config: %w", err))
	}
	log.Infof("using %q publisher", pub.Name())

	scheduledFeeds, err := appConfig.GetScheduledFeeds()
	feedNames := []string{}
	for k := range scheduledFeeds {
		feedNames = append(feedNames, k)
	}
	log.Infof("watching feeds: %v", strings.Join(feedNames, ", "))
	if err != nil {
		log.Fatal(err)
	}
	sched := scheduler.New(scheduledFeeds)

	log.Printf("listening on port %v", appConfig.HTTPPort)
	pollRate, err := time.ParseDuration(appConfig.PollRate)
	if err != nil {
		log.Fatal(err)
	}
	handler := &FeedHandler{
		scheduler: sched,
		pub:       pub,
		pollRate:  pollRate,
	}

	if appConfig.Timer {
		cronjob := cron.New()
		crontab := fmt.Sprintf("@every %s", pollRate.String())
		log.Printf("Running a timer %s", crontab)
		err := cronjob.AddFunc(crontab, func() { cronPoll(handler) })
		if err != nil {
			log.Fatal(err)
		}
		cronjob.Start()
	}

	http.Handle("/", handler)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", appConfig.HTTPPort), nil); err != nil {
		log.Fatal(err)
	}
}

func cronPoll(handler *FeedHandler) {
	pkgs, pollErrors := handler.pollFeeds()
	processedPackages, err := handler.publishPackages(pkgs)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	if pollErrors {
		// pollFeeds already logs with ErrorF.
		return
	}
	log.Printf("%d packages processed", processedPackages)
}
