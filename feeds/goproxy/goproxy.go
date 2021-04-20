package goproxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "goproxy"
)

var (
	baseURL    = "https://index.golang.org/index"
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

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

func fetchPackages(since time.Time) ([]Package, error) {
	var packages []Package
	params := url.Values{}
	params.Add("since", since.Format(time.RFC3339))
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	resp, err := httpClient.Get(requestURL)
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

type Feed struct{}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages(cutoff)
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
