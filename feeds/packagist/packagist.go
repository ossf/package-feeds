package packagist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const FeedName = "packagist"

var (
	updateHost  = "https://packagist.org"
	versionHost = "https://repo.packagist.org"
)

type response struct {
	Actions   []actions `json:"actions"`
	Timestamp int64     `json:"timestamp"`
}

type actions struct {
	Type    string `json:"type"`
	Package string `json:"package"`
	Time    int64  `json:"time"`
}

type Feed struct{}

func fetchPackages(since time.Time) ([]actions, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/metadata/changes.json", updateHost), nil)
	if err != nil {
		return nil, err
	}
	values := request.URL.Query()
	sinceStr := strconv.FormatInt(since.Unix()*10000, 10)
	values.Add("since", sinceStr)
	request.URL.RawQuery = values.Encode()
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiResponse := &response{}
	err = json.NewDecoder(resp.Body).Decode(apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse.Actions, nil
}

func fetchVersionInformation(action actions) ([]*feeds.Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/p2/%s.json", versionHost, action.Package))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
func (f Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages(cutoff)
	if err != nil {
		return nil, err
	}
	for _, pkg := range packages {
		if time.Unix(pkg.Time, 0).Before(cutoff) {
			continue
		}
		if pkg.Type == "delete" {
			continue
		}
		updates, err := fetchVersionInformation(pkg)
		if err != nil {
			return nil, fmt.Errorf("error in fetching version information: %w", err)
		}
		pkgs = append(pkgs, updates...)
	}
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
