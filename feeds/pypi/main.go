package main

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"
)

const (
	delta   = 5 * time.Minute
	baseURL = "https://pypi.org/rss/updates.xml"
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	ModifiedDate rfc1123Time `xml:"pubDate"`
	Link         string      `xml:"link"`
	Name         string      `xml:"-"`
	Version      string      `xml:"-"`
}

type rfc1123Time struct {
	time.Time
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
	packages, err := fetchPackages()
	if err != nil {
		return err
	}
	cutoff := time.Now().UTC().Add(-delta)
	for _, pkg := range packages {
		if pkg.ModifiedDate.Before(cutoff) {
			continue
		}
		// TODO: publish the package up to a cloud pub/sub for processing
		packages = append(packages, pkg)
	}
	return nil
}

func main() {
	err := Poll(context.Background(), PubSubMessage{})
	if err != nil {
		panic(err)
	}
}
