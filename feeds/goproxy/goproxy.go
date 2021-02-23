package goproxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "goproxy"
	baseURL  = "https://index.golang.org/index"
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
	client := &http.Client{Timeout: 10 * time.Second}
	params := url.Values{}
	params.Add("since", since.Format(time.RFC3339))
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	resp, err := client.Get(requestURL)
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
		log.Println("Processing:", pkg.Title, pkg.Version)
		if pkg.ModifiedDate.Before(cutoff) {
			continue
		}
		pkgs = append(pkgs, &feeds.Package{
			Name:        pkg.Title,
			Version:     pkg.Version,
			Type:        FeedName,
			CreatedDate: pkg.ModifiedDate,
		})
	}
	return pkgs, nil
}
