package goproxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/utils"
)

const (
	FeedName  = "goproxy"
	indexPath = "/index"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type PackageJSON struct {
	Path      string `json:"Path"`
	Version   string `json:"Version"`
	Timestamp string `json:"Timestamp"`
}

type Package struct {
	Title        string
	ModifiedDate time.Time
	Version      string
}

func fetchPackages(baseURL string, since time.Time) ([]Package, error) {
	var packages []Package
	indexURL, err := utils.URLPathJoin(baseURL, indexPath)
	if err != nil {
		return nil, err
	}
	pkgURL, err := url.Parse(indexURL)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("since", since.Format(time.RFC3339))
	pkgURL.RawQuery = params.Encode()

	resp, err := httpClient.Get(pkgURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var packageJSON PackageJSON
		err = json.Unmarshal([]byte(scanner.Text()), &packageJSON)
		if err != nil {
			return nil, err
		}
		modifiedDate, err := time.Parse(time.RFC3339, packageJSON.Timestamp)
		if err != nil {
			return nil, err
		}

		pkg := Package{
			Title:        packageJSON.Path,
			ModifiedDate: modifiedDate,
			Version:      packageJSON.Version,
		}
		packages = append(packages, pkg)
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return packages, nil
}

type Feed struct {
	baseURL string
}

func New(feedOptions feeds.FeedOptions) (*Feed, error) {
	if feedOptions.Packages != nil {
		return nil, feeds.UnsupportedOptionError{
			Feed:   FeedName,
			Option: "packages",
		}
	}
	return &Feed{
		baseURL: "https://index.golang.org/",
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages(feed.baseURL, cutoff)
	if err != nil {
		return pkgs, fmt.Errorf("error fetching packages: %w", err)
	}
	for _, pkg := range packages {
		pkg := feeds.NewPackage(pkg.ModifiedDate, pkg.Title, pkg.Version, FeedName)
		pkgs = append(pkgs, pkg)
	}
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
