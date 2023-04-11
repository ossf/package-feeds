package packagist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/utils"
)

const FeedName = "packagist"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type response struct {
	Actions   []actions `json:"actions"`
	Timestamp int64     `json:"timestamp"`
}

type actions struct {
	Type    string `json:"type"`
	Package string `json:"package"`
	Time    int64  `json:"time"`
}

type Feed struct {
	updateHost  string
	versionHost string
	options     feeds.FeedOptions
}

func New(feedOptions feeds.FeedOptions) (*Feed, error) {
	if feedOptions.Packages != nil {
		return nil, feeds.UnsupportedOptionError{
			Feed:   FeedName,
			Option: "packages",
		}
	}
	return &Feed{
		updateHost:  "https://packagist.org",
		versionHost: "https://repo.packagist.org",
		options:     feedOptions,
	}, nil
}

func fetchPackages(updateHost string, since time.Time) ([]actions, error) {
	pkgURL, err := url.JoinPath(updateHost, "/metadata/changes.json")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodGet, pkgURL, nil)
	if err != nil {
		return nil, err
	}
	values := request.URL.Query()
	sinceStr := strconv.FormatInt(since.Unix()*10000, 10)
	values.Add("since", sinceStr)
	request.URL.RawQuery = values.Encode()
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.CheckResponseStatus(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packagist package data: %w", err)
	}

	apiResponse := &response{}
	err = json.NewDecoder(resp.Body).Decode(apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse.Actions, nil
}

func fetchVersionInformation(versionHost string, action actions) ([]*feeds.Package, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/p2/%s.json", versionHost, action.Package))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.CheckResponseStatus(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packagist package version data: %w", err)
	}

	versionResponse := &packages{}
	err = json.NewDecoder(resp.Body).Decode(versionResponse)
	if err != nil {
		return nil, err
	}

	pkgs := []*feeds.Package{}
	for pkgName, versions := range versionResponse.Packages {
		for _, version := range versions {
			pkg := feeds.NewPackage(version.Time, pkgName, version.Version, FeedName)
			if err != nil {
				continue
			}
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}

// Latest returns all package updates of packagist packages since cutoff.
func (f Feed) Latest(cutoff time.Time) ([]*feeds.Package, []error) {
	pkgs := []*feeds.Package{}
	var errs []error
	packages, err := fetchPackages(f.updateHost, cutoff)
	if err != nil {
		return nil, append(errs, err)
	}
	for _, pkg := range packages {
		if time.Unix(pkg.Time, 0).Before(cutoff) {
			continue
		}
		if pkg.Type == "delete" {
			continue
		}
		updates, err := fetchVersionInformation(f.versionHost, pkg)
		if err != nil {
			errs = append(errs, fmt.Errorf("error in fetching version information: %w", err))
			continue
		}
		pkgs = append(pkgs, updates...)
	}
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, errs
}

func (f Feed) GetName() string {
	return FeedName
}

func (f Feed) GetFeedOptions() feeds.FeedOptions {
	return f.options
}
