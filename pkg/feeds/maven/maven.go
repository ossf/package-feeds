package maven

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/useragent"
)

const (
	FeedName  = "maven-central"
	indexPath = "/api/internal/browse/components"
)

type Feed struct {
	baseURL string
	options feeds.FeedOptions
}

var (
	httpClient = &http.Client{
		Transport: &useragent.RoundTripper{UserAgent: feeds.DefaultUserAgent},
		Timeout:   10 * time.Second,
	}

	ErrMaxRetriesReached = errors.New("maximum retries reached due to rate limiting")
)

func New(feedOptions feeds.FeedOptions) (*Feed, error) {
	if feedOptions.Packages != nil {
		return nil, feeds.UnsupportedOptionError{
			Feed:   FeedName,
			Option: "packages",
		}
	}
	return &Feed{
		baseURL: "https://central.sonatype.com/" + indexPath,
		options: feedOptions,
	}, nil
}

// Package represents package information.
type LatestVersionInfo struct {
	Version             string `json:"version"`
	TimestampUnixWithMS int64  `json:"timestampUnixWithMS"`
}

type Package struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	LatestVersionInfo LatestVersionInfo `json:"latestVersionInfo"`
}

// Response represents the response structure from Sonatype API.
type Response struct {
	Components []Package `json:"components"`
}

// fetchPackages fetches packages from Sonatype API for the given page.
func (feed Feed) fetchPackages(page int) ([]Package, error) {
	maxRetries := 5
	retryDelay := 5 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Define the request payload
		payload := map[string]interface{}{
			"page":          page,
			"size":          20,
			"sortField":     "publishedDate",
			"sortDirection": "desc",
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error encoding JSON: %w", err)
		}

		// Send POST request to Sonatype API.
		resp, err := httpClient.Post(feed.baseURL+"?repository=maven-central", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			// Check if maximum retries have been reached
			if attempt == maxRetries {
				return nil, fmt.Errorf("error sending request: %w", err)
			}
			time.Sleep(retryDelay) // Wait before retrying
			continue
		}
		defer resp.Body.Close()

		// Handle rate limiting (HTTP status code 429).
		if resp.StatusCode == http.StatusTooManyRequests {
			// Check if maximum retries have been reached
			if attempt == maxRetries {
				return nil, ErrMaxRetriesReached
			}
			time.Sleep(retryDelay) // Wait before retrying
			continue
		}

		// Decode response.
		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, fmt.Errorf("error decoding response: %w", err)
		}
		return response.Components, nil
	}
	return nil, ErrMaxRetriesReached
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, time.Time, []error) {
	pkgs := []*feeds.Package{}
	var errs []error

	page := 0
	for {
		// Fetch packages from Sonatype API for the current page.
		packages, err := feed.fetchPackages(page)
		if err != nil {
			errs = append(errs, err)
			break
		}

		// Iterate over packages
		hasToCut := false
		for _, pkg := range packages {
			// convert published to date to compare with cutoff
			if pkg.LatestVersionInfo.TimestampUnixWithMS > cutoff.UnixMilli() {
				// Append package to pkgs
				timestamp := time.Unix(pkg.LatestVersionInfo.TimestampUnixWithMS/1000, 0)
				packageName := pkg.Namespace + ":" + pkg.Name

				newPkg := feeds.NewPackage(timestamp, packageName, pkg.LatestVersionInfo.Version, FeedName)
				pkgs = append(pkgs, newPkg)
			} else {
				// Break the loop if the cutoff date is reached
				hasToCut = true
			}
		}

		// Move to the next page
		page++

		// Check if the loop should be terminated
		if len(pkgs) == 0 || hasToCut {
			break
		}
	}

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
