package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jordan-wright/ossmalware/pkg/library"
	"github.com/jordan-wright/ossmalware/pkg/processor"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	delta   = 5 * time.Minute
	baseURL = "https://pypi.org/rss/updates.xml"
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	Title        string      `xml:"title"`
	ModifiedDate rfc1123Time `xml:"pubDate"`
	Link         string      `xml:"link"`
}

type rfc1123Time struct {
	time.Time
}

func (p *Package) Name() string {
	// The XML Feed has a "Title" element that contains the package and version in it.
	return strings.Split(p.Title, " ")[0]
}

func (p *Package) Version() string {
	// The XML Feed has a "Title" element that contains the package and version in it.
	return strings.Split(p.Title, " ")[1]
}

func (t *rfc1123Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var marshaledTime string
	err := d.DecodeElement(&marshaledTime, &start)
	if err != nil {
		return err
	}
	decodedTime, err := time.Parse(time.RFC1123, marshaledTime)
	if err != nil {
		return err
	}
	*t = rfc1123Time{decodedTime}
	return nil
}

func fetchPackages() ([]*Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rssResponse := &Response{}
	err = xml.NewDecoder(resp.Body).Decode(rssResponse)
	if err != nil {
		return nil, err
	}
	return rssResponse.Packages, nil
}

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// Poll receives a message from Cloud Pub/Sub. Ideally, this will be from a
// Cloud Scheduler trigger every `delta`.
func Poll(ctx context.Context, m PubSubMessage) error {
	topicUrl := os.Getenv("OSSMALWARE_TOPIC_URL")
	topic, err := pubsub.OpenTopic(ctx, topicUrl)
	if err != nil {
		panic(err)
	}

	packages, err := fetchPackages()
	if err != nil {
		return err
	}
	cutoff := time.Now().UTC().Add(-delta)
	for _, pkg := range packages {
		log.Println("Processing:", pkg.Name(), pkg.Version())
		if pkg.ModifiedDate.Before(cutoff) {
			continue
		}
		msg := library.Package{
			Name:    pkg.Name(),
			Version: pkg.Version(),
			Type:    processor.TypePyPI,
		}
		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if err := topic.Send(ctx, &pubsub.Message{
			Body: b,
		}); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := Poll(context.Background(), PubSubMessage{}); err != nil {
		panic(err)
	}
}
