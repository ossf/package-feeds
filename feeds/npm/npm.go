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
	FeedName = "npm"
)

var (
	baseURL    = "https://registry.npmjs.org/-/rss"
	versionURL = "https://registry.npmjs.org/"
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
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
	resp, err := httpClient.Get(baseURL)
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
	resp, err := httpClient.Get(fmt.Sprintf("%s/%s", versionURL, packageName))
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
		v, err := fetchVersionInformation(pkg.Title)
		if err != nil {
			return pkgs, fmt.Errorf("error in fetching version information: %w", err)
		}
		pkg := feeds.NewPackage(pkg.CreatedDate.Time, pkg.Title, v, FeedName)
		pkgs = append(pkgs, pkg)
	}
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
