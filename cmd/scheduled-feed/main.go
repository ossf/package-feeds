package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

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
	pkgs, err := handler.scheduler.Poll(cutoff)
	if err != nil {
		log.Errorf("error polling for new packages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, pkg := range pkgs {
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
}

func main() {
	pubURL := os.Getenv("OSSMALWARE_TOPIC_URL")
	var pub publisher.Publisher
	var err error
	if pubURL == "" {
		pub = publisher.NewStdoutPublisher()
	} else {
		pub, err = publisher.NewPubSub(context.TODO(), pubURL)
		if err != nil {
			log.Fatal("error creating gcp pubsub topic with url %q: %v", pubURL, err)
		}
	}
	log.Infof("using %q publisher", pub.Name())
	sched := scheduler.New()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on port %s", port)
	handler := &FeedHandler{
		scheduler: sched,
		pub:       pub,
	}
	http.Handle("/", handler)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
