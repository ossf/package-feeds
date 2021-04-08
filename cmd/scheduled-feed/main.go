package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ossf/package-feeds/config"
	"github.com/ossf/package-feeds/feeds/scheduler"
	"github.com/ossf/package-feeds/publisher"

	log "github.com/sirupsen/logrus"
)

const delta = 5 * time.Minute

// FeedHandler is a handler that fetches new packages from various feeds
type FeedHandler struct {
	scheduler *scheduler.Scheduler
	pub       publisher.Publisher
}

func (handler *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cutoff := time.Now().UTC().Add(-delta)
	pkgs, errs := handler.scheduler.Poll(cutoff)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Errorf("error polling for new packages: %v", err)
		}
	}
	processed := 0
	for _, pkg := range pkgs {
		processed++
		log.WithFields(log.Fields{
			"name":         pkg.Name,
			"feed":         pkg.Type,
			"created_date": pkg.CreatedDate,
		}).Print("sending package upstream")
		b, err := json.Marshal(pkg)
		if err != nil {
			log.Printf("error marshaling package: %#v", pkg)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := handler.pub.Send(context.Background(), b); err != nil {
			log.Printf("error sending package to upstream publisher %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(errs) > 0 {
		http.Error(w, "error polling for packages - see logs for more information", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%d packages processed", processed)))
}

func main() {
	configPath, useConfig := os.LookupEnv("PACKAGE_FEEDS_CONFIG_PATH")
	var err error

	var appConfig *config.ScheduledFeedConfig
	if useConfig {
		appConfig, err = config.FromFile(configPath)
	} else {
		appConfig = config.Default()
	}
	if err != nil {
		log.Fatal(err)
	}

	pub, err := appConfig.PubConfig.ToPublisher(context.TODO())
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initialize publisher from config: %w", err))
	}
	log.Infof("using %q publisher", pub.Name())

	feeds, err := appConfig.GetScheduledFeeds()
	feedNames := []string{}
	for k, _ := range feeds {
		feedNames = append(feedNames, k)
	}
	log.Infof("watching feeds: %v", strings.Join(feedNames, ", "))
	if err != nil {
		log.Fatal(err)
	}
	sched := scheduler.New(feeds)

	log.Printf("listening on port %v", appConfig.HttpPort)
	handler := &FeedHandler{
		scheduler: sched,
		pub:       pub,
	}
	http.Handle("/", handler)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", appConfig.HttpPort), nil); err != nil {
		log.Fatal(err)
	}
}
