package rubygems

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "rubygems"
	baseURL  = "https://rubygems.org/api/v1/activity"
)

type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"version_created_at"`
}

func fetchPackages(url string) ([]*Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := []*Package{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

type Feed struct{}

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
		if pkg.CreatedDate.Before(cutoff) {
			continue
		}
		pkgs = append(pkgs, &feeds.Package{
			Name:        pkg.Name,
			Version:     pkg.Version,
			CreatedDate: pkg.CreatedDate,
			Type:        FeedName,
		})
	}
	return pkgs, nil
}
