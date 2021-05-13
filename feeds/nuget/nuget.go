package nuget

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName           = "nuget"
	catalogServiceType = "Catalog/3.0.0"
	indexPath          = "/v3/index.json"
)

var (
	httpClient = http.Client{
		Timeout: 10 * time.Second,
	}
	errCatalogService = errors.New("error fetching catalog service")
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

func fetchCatalogService(baseURL string) (*nugetService, error) {
	var err error
	catalogServiceURL := fmt.Sprintf("%s/%s", baseURL, indexPath)
	resp, err := httpClient.Get(catalogServiceURL)
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
	return nil, fmt.Errorf("%w : could not locate catalog service for nuget feed %s",
		errCatalogService, catalogServiceURL)
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
		baseURL: "https://api.nuget.org/",
	}, nil
}

// Latest will parse all creation events for packages in the nuget.org catalog feed
// for packages that have been published since the cutoff
// https://docs.microsoft.com/en-us/nuget/api/catalog-resource
func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}

	catalogService, err := fetchCatalogService(feed.baseURL)
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

			pkgInfo, err := fetchPackageInfo(catalogLeafNode.URI)
			if err != nil {
				return nil, err
			}

			pkg := feeds.NewPackage(pkgInfo.Created, pkgInfo.PackageID, pkgInfo.Version, FeedName)
			if err != nil {
				continue
			}
			pkgs = append(pkgs, pkg)
		}
	}
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)

	return pkgs, nil
}
