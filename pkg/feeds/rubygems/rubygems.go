package rubygems

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/useragent"
	"github.com/ossf/package-feeds/pkg/utils"
)

const (
	FeedName     = "rubygems"
	activityPath = "/api/v1/activity"
)

var httpClient = &http.Client{
	Transport: &useragent.RoundTripper{UserAgent: feeds.DefaultUserAgent},
	Timeout:   10 * time.Second,
}

type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"version_created_at"`
}

func fetchPackages(packagesURL string) ([]*Package, error) {
	resp, err := httpClient.Get(packagesURL)
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
		baseURL:          "https://rubygems.org",
		options:          feedOptions,
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, time.Time, []error) {
	pkgs := []*feeds.Package{}
	packages := make(map[string]*Package)
	var errs []error

	newPackagesURL, err := url.JoinPath(feed.baseURL, activityPath, "latest.json")
	if err != nil {
		// Failure to construct a url should lead to a hard failure.
		return nil, cutoff, append(errs, err)
	}
	newPackages, err := fetchPackages(newPackagesURL)
	if err != nil {
		// Updated Packages could still be processed.
		errs = append(errs, err)
	} else {
		for _, pkg := range newPackages {
			packages[pkg.Name] = pkg
		}
	}
	updatedPackagesURL, err := url.JoinPath(feed.baseURL, activityPath, "just_updated.json")
	if err != nil {
		// Failure to construct a url should lead to a hard failure.
		return nil, cutoff, append(errs, err)
	}
	updatedPackages, err := fetchPackages(updatedPackagesURL)
	if err != nil {
		// New Packages could still be processed.
		errs = append(errs, err)
	} else {
		for _, pkg := range updatedPackages {
			packages[pkg.Name] = pkg
		}
	}

	for _, pkg := range packages {
		pkg := feeds.NewPackage(pkg.CreatedDate, pkg.Name, pkg.Version, FeedName)
		pkgs = append(pkgs, pkg)
	}
	feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)

	newCutoff := feeds.FindCutoff(cutoff, pkgs)
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, newCutoff, errs
}

func (feed Feed) GetName() string {
	return FeedName
}

func (feed Feed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}
