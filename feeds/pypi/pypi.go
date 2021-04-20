package pypi

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "pypi"
)

var (
	baseURL    = "https://pypi.org/rss/updates.xml"
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

type Feed struct {
	lossyFeedAlerter *feeds.LossyFeedAlerter
}

func New(eventHandler *events.Handler) *Feed {
	return &Feed{
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
	}
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	pypiPackages, err := fetchPackages()
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range pypiPackages {
		pkg := feeds.NewPackage(pkg.CreatedDate.Time, pkg.Name(), pkg.Version(), FeedName)
		pkgs = append(pkgs, pkg)
	}
	feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
