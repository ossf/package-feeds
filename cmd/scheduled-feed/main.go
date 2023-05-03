package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/pkg/config"
	"github.com/ossf/package-feeds/pkg/scheduler"
)

func main() {
	// Increase idle conns per host to increase the reuse of existing
	// connections between requests.
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 8

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
		log.Fatalf("Failed to initialize publisher from config: %v", err)
	}
	log.Infof("Using %q publisher", pub.Name())

	scheduledFeeds, err := appConfig.GetScheduledFeeds()
	feedNames := []string{}
	for k := range scheduledFeeds {
		feedNames = append(feedNames, k)
	}
	log.Infof("Watching feeds: %v", strings.Join(feedNames, ", "))
	if err != nil {
		log.Fatal(err)
	}

	pollRate, err := time.ParseDuration(appConfig.PollRate)
	if err != nil {
		log.Fatalf("Failed to parse poll_rate to duration: %v", err)
	}
	sched := scheduler.New(scheduledFeeds, pub, appConfig.HTTPPort)
	err = sched.Run(pollRate, appConfig.Timer)
	if err != nil {
		log.Fatal(err)
	}
}
