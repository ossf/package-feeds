package crates

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "crates"
)

var (
	baseURL    = "https://crates.io/api/v1/summary"
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

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
func fetchPackages() ([]*Package, error) {
	resp, err := httpClient.Get(baseURL)
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
}

func New(eventHandler *events.Handler) *Feed {
	return &Feed{
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
	}
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages()
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
