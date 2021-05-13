package crates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName     = "crates"
	activityPath = "/api/v1/summary"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type crates struct {
	JustUpdated []*Package `json:"just_updated"`
}

// Package stores the information from crates.io updates.
type Package struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	UpdatedAt        time.Time `json:"updated_at"`
	NewestVersion    string    `json:"newest_version"`
	MaxStableVersion string    `json:"max_stable_version"`
	Repository       string    `json:"repository"`
}

// Gets crates.io packages.
func fetchPackages(baseURL string) ([]*Package, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s%s", baseURL, activityPath))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	v := &crates{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return nil, err
	}
	// TODO: We should check both the NewCrates as well.
	return v.JustUpdated, nil
}

type Feed struct {
	lossyFeedAlerter *feeds.LossyFeedAlerter
	baseURL          string
}

func New(feedOptions feeds.FeedOptions, eventHandler *events.Handler) (*Feed, error) {
	if feedOptions.Packages != nil {
		return nil, feeds.UnsupportedOptionError{
			Feed:   FeedName,
			Option: "packages",
		}
	}
	return &Feed{
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
		baseURL:          "https://crates.io/api/v1/summary",
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages(feed.baseURL)
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range packages {
		pkg := feeds.NewPackage(pkg.UpdatedAt, pkg.Name, pkg.NewestVersion, FeedName)
		pkgs = append(pkgs, pkg)
	}
	feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
