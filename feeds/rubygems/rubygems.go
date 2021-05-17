package rubygems

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/utils"
)

const (
	FeedName     = "rubygems"
	activityPath = "/api/v1/activity"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"version_created_at"`
}

func fetchPackages(url string) ([]*Package, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.CheckResponseStatus(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rubygems package data: %w", err)
	}

	response := []*Package{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response, err
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
		baseURL:          "https://rubygems.org",
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages := make(map[string]*Package)

	newPackagesURL, err := utils.URLPathJoin(feed.baseURL, activityPath, "latest.json")
	if err != nil {
		return nil, err
	}
	newPackages, err := fetchPackages(newPackagesURL)
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range newPackages {
		packages[pkg.Name] = pkg
	}
	updatedPackagesURL, err := utils.URLPathJoin(feed.baseURL, activityPath, "just_updated.json")
	if err != nil {
		return nil, err
	}
	updatedPackages, err := fetchPackages(updatedPackagesURL)
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range updatedPackages {
		packages[pkg.Name] = pkg
	}

	for _, pkg := range packages {
		pkg := feeds.NewPackage(pkg.CreatedDate, pkg.Name, pkg.Version, FeedName)
		pkgs = append(pkgs, pkg)
	}
	feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
