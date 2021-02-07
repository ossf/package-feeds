package npm

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName   = "npm"
	baseURL    = "https://registry.npmjs.org/-/rss"
	versionURL = "https://registry.npmjs.org/"
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	Title       string      `xml:"title"`
	CreatedDate rfc1123Time `xml:"pubDate"`
	Link        string      `xml:"link"`
	Version     string
}

type PackageVersion struct {
	ID       string `json:"_id"`
	Rev      string `json:"_rev"`
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

type rfc1123Time struct {
	time.Time
}

func (t *rfc1123Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var marshaledTime string
	err := d.DecodeElement(&marshaledTime, &start)
	if err != nil {
		return err
	}
	decodedTime, err := time.Parse(time.RFC1123, marshaledTime)
	if err != nil {
		return err
	}
	*t = rfc1123Time{decodedTime}
	return nil
}

func fetchPackages() ([]*Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rssResponse := &Response{}
	err = xml.NewDecoder(resp.Body).Decode(rssResponse)
	if err != nil {
		return nil, err
	}
	return rssResponse.Packages, nil
}

// Gets the package version from the NPM.
func fetchVersionInformation(packageName string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/%s", versionURL, packageName))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	v := &PackageVersion{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", err
	}
	return v.DistTags.Latest, nil
}

type Feed struct{}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packages, err := fetchPackages()
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range packages {
		if pkg.CreatedDate.Before(cutoff) {
			continue
		}
		v, err := fetchVersionInformation(pkg.Title)
		if err != nil {
			return pkgs, fmt.Errorf("error in fetching version information: %w", err)
		}
		pkgs = append(pkgs, &feeds.Package{
			Name:        pkg.Title,
			Version:     v,
			Type:        FeedName,
			CreatedDate: pkg.CreatedDate.Time,
		})
	}
	return pkgs, nil
}
