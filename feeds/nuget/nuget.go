package nuget

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName           = "nuget"
	feedURL            = "https://api.nuget.org/v3/index.json"
	catalogServiceType = "Catalog/3.0.0"
)

type serviceIndex struct {
	Services []*nugetService `json:"resources"`
}

type nugetService struct {
	URI  string `json:"@id"`
	Type string `json:"@type"`
}

type catalog struct {
	Pages []*catalogPage `json:"items"`
}

type catalogPage struct {
	URI      string         `json:"@id"`
	Created  time.Time      `json:"commitTimeStamp"`
	Packages []*catalogLeaf `json:"items"`
}

type catalogLeaf struct {
	URI            string    `json:"@id"`
	CatalogCreated time.Time `json:"commitTimeStamp"`
	Type           string    `json:"@type"`
}

type nugetPackageDetails struct {
	PackageID string    `json:"id"`
	Version   string    `json:"version"`
	Created   time.Time `json:"published"`
}

var httpClient = http.Client{
	Timeout: 10 * time.Second,
}

func fetchCatalogService() (*nugetService, error) {
	resp, err := httpClient.Get(feedURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	directory := &serviceIndex{}
	err = json.NewDecoder(resp.Body).Decode(directory)
	if err != nil {
		return nil, err
	}

	for _, service := range directory.Services {
		if service.Type == catalogServiceType {
			return service, nil
		}
	}

	return nil, fmt.Errorf("Could not locate catalog service for nuget feed (%s)", feedURL)
}

func fetchCatalogPages(catalogURL string) ([]*catalogPage, error) {
	resp, err := httpClient.Get(catalogURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	c := &catalog{}
	err = json.NewDecoder(resp.Body).Decode(c)
	if err != nil {
		return nil, err
	}

	return c.Pages, nil
}

func fetchCatalogPage(url string) ([]*catalogLeaf, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	page := &catalogPage{}
	err = json.NewDecoder(resp.Body).Decode(page)
	if err != nil {
		return nil, err
	}

	return page.Packages, nil
}

func fetchPackageInfo(url string) (*nugetPackageDetails, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	packageDetail := &nugetPackageDetails{}
	err = json.NewDecoder(resp.Body).Decode(packageDetail)
	if err != nil {
		return nil, err
	}

	return packageDetail, nil
}

type Feed struct{}

// Latest will parse all creation events for packages in the nuget.org catalog feed
// for packages that have been published since the cutoff
// https://docs.microsoft.com/en-us/nuget/api/catalog-resource
func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}

	catalogService, err := fetchCatalogService()
	if err != nil {
		return nil, err
	}

	catalogPages, err := fetchCatalogPages(catalogService.URI)
	if err != nil {
		return nil, err
	}

	for _, catalogPage := range catalogPages {
		if catalogPage.Created.Before(cutoff) {
			continue
		}

		page, err := fetchCatalogPage(catalogPage.URI)
		if err != nil {
			return nil, err
		}

		for _, catalogLeafNode := range page {
			if catalogLeafNode.CatalogCreated.Before(cutoff) {
				continue
			}

			if catalogLeafNode.Type != "nuget:PackageDetails" {
				continue // Not currently interested in package deletion events
			}

			packageCreationDetail, err := fetchPackageInfo(catalogLeafNode.URI)
			if err != nil {
				return nil, err
			}

			if packageCreationDetail.Created.Before(cutoff) {
				continue
			}

			pkgs = append(pkgs, &feeds.Package{
				Name:        packageCreationDetail.PackageID,
				CreatedDate: packageCreationDetail.Created,
				Version:     packageCreationDetail.Version,
				Type:        FeedName,
			})
		}
	}

	return pkgs, nil
}
