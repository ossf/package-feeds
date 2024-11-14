package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
)

const (
	FeedName = "github"
)

var (
	urlFormatString = "https://api.github.com/repos/%v/releases?per_page=%v"
	httpClient      = &http.Client{
		Timeout: 10 * time.Second,
	}
	releasesPerQuery           = 20
	errPackageOptionsUnset     = errors.New("github feed requires packages to be configured as a feed option")
	errMinimumPackagesRequired = errors.New("github feed requires a minimum of 1 package supplied under options")
)

type Feed struct {
	repositories []string
	options      feeds.FeedOptions
}

func New(feedOptions feeds.FeedOptions) (*Feed, error) {
	if feedOptions.Packages == nil {
		return nil, errPackageOptionsUnset
	}
	if len(*feedOptions.Packages) == 0 {
		return nil, errMinimumPackagesRequired
	}
	return &Feed{
		repositories: *feedOptions.Packages,
		options:      feedOptions,
	}, nil
}

func (feed Feed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}

func (feed Feed) GetName() string {
	return "github"
}

type releaseList []*release

type release struct {
	TagName string `json:"tag_name"`

	// TODO: Add optional filter of Draft/Prerelease
	Draft      bool `json:"draft"`
	Prerelease bool `json:"prerelease"`

	PublishedAt time.Time `json:"published_at"`
}

func fetchReleases(repository string) (releaseList, error) {
	releases := releaseList{}
	resp, err := httpClient.Get(fmt.Sprintf(urlFormatString, repository, releasesPerQuery))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, []error) {
	pkgs := []*feeds.Package{}
	pkgChannel := make(chan []*feeds.Package)
	errs := []error{}
	errChannel := make(chan error)

	for _, repo := range feed.repositories {
		go func(repo string) {
			repoReleases, err := fetchReleases(repo)
			if err != nil {
				errChannel <- err
				return
			}
			pkgChannel <- releasesToPackages(repo, repoReleases)
		}(repo)
	}

	for i := 0; i < len(feed.repositories); i++ {
		select {
		case pkgSlice := <-pkgChannel:
			pkgs = append(pkgs, pkgSlice...)
		case err := <-errChannel:
			errs = append(errs, err)
		}
	}
	// Sort packages to ensure deterministic ordering.
	feeds.SortPackages(pkgs)

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, errs
}

func releasesToPackages(repo string, releases releaseList) []*feeds.Package {
	pkgs := []*feeds.Package{}
	for _, release := range releases {
		pkgs = append(pkgs, feeds.NewPackage(release.PublishedAt, repo, release.TagName, FeedName))
	}
	return pkgs
}
