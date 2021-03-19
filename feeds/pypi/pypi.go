package pypi

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "pypi"
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	Title       string      `xml:"title"`
	CreatedDate rfc1123Time `xml:"pubDate"`
	Link        string      `xml:"link"`
}

func (p *Package) Name() string {
	// The XML Feed has a "Title" element that contains the package and version in it.
	return strings.Split(p.Title, " ")[0]
}

func (p *Package) Version() string {
	// The XML Feed has a "Title" element that contains the package and version in it.
	return strings.Split(p.Title, " ")[1]
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

func fetchPackages(url string) ([]*Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
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

type Feed struct {
	baseURL string
}

func NewFeed() *Feed {
	return &Feed{
		baseURL: "https://pypi.org/rss/updates.xml",
	}
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	pypiPackages, err := fetchPackages(feed.baseURL)
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range pypiPackages {
		if pkg.CreatedDate.Before(cutoff) {
			continue
		}
		pkgs = append(pkgs, &feeds.Package{
			Name:        pkg.Name(),
			Version:     pkg.Version(),
			CreatedDate: pkg.CreatedDate.Time,
			Type:        FeedName,
		})
	}
	return pkgs, nil
}
