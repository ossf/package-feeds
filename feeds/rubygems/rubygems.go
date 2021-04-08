package rubygems

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "rubygems"
)

var (
	baseURL    = "https://rubygems.org/api/v1/activity"
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

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
	response := []*Package{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

type Feed struct {
	lossyFeedAlerter *feeds.LossyFeedAlerter
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
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages := make(map[string]*Package)
	newPackages, err := fetchPackages(fmt.Sprintf("%s/%s", baseURL, "latest.json"))
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range newPackages {
		packages[pkg.Name] = pkg
	}
	updatedPackages, err := fetchPackages(fmt.Sprintf("%s/%s", baseURL, "just_updated.json"))
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
