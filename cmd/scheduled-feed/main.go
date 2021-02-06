package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ossf/package-feeds/feeds/scheduler"

	log "github.com/sirupsen/logrus"

	_ "gocloud.dev/pubsub/gcppubsub"
)

var delta = 5 * time.Minute

func Poll(w http.ResponseWriter, r *http.Request) {
	// topicURL := os.Getenv("OSSMALWARE_TOPIC_URL")
	// topic, err := pubsub.OpenTopic(context.TODO(), topicURL)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	cutoff := time.Now().UTC().Add(-delta)
	pkgs, err := scheduler.PollScheduledFeeds(cutoff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, pkg := range pkgs {
		log.WithFields(log.Fields{
			"name":         pkg.Name,
			"feed":         pkg.Type,
			"created_date": pkg.CreatedDate,
		}).Print("sending package upstream")
		// b, err := json.Marshal(pkg)
		// if err != nil {
		// 	log.Printf("error marshaling package: %#v", pkg)
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// if err := topic.Send(context.TODO(), &pubsub.Message{
		// 	Body: b,
		// }); err != nil {
		// 	log.Printf("error sending package to upstream topic %s: %v", topicURL, err)
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on port %s", port)
	http.HandleFunc("/", Poll)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
