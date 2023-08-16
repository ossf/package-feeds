package crates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/utils"
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
	pkgURL, err := url.JoinPath(baseURL, activityPath)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Get(pkgURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.CheckResponseStatus(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crates package data: %w", err)
	}

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
	options          feeds.FeedOptions
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
		baseURL:          "https://crates.io",
		options:          feedOptions,
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, time.Time, []error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages(feed.baseURL)
	if err != nil {
		return pkgs, cutoff, []error{err}
	}
	for _, pkg := range packages {
		pkg := feeds.NewPackage(pkg.UpdatedAt, pkg.Name, pkg.NewestVersion, FeedName)
		pkgs = append(pkgs, pkg)
	}
	feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)

	newCutoff := feeds.FindCutoff(cutoff, pkgs)
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, newCutoff, []error{}
}

func (feed Feed) GetName() string {
	return FeedName
}

func (feed Feed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}
