package pypi

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/utils"
)

const (
	FeedName          = "pypi"
	updatesPath       = "/rss/updates.xml"
	packagePathFormat = "/rss/project/%s/releases.xml"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	errInvalidLinkForPackage = errors.New("invalid link provided by pypi API")
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	Title       string      `xml:"title"`
	CreatedDate rfc1123Time `xml:"pubDate"`
	Link        string      `xml:"link"`
}

func (p *Package) Name() (string, error) {
	// The XML Link splits to: []string{"https:", "", "pypi.org", "project", "foopy", "2.1", ""}
	parts := strings.Split(p.Link, "/")
	if len(parts) < 5 {
		return "", errInvalidLinkForPackage
	}
	return parts[len(parts)-3], nil
}

func (p *Package) Version() (string, error) {
	// The XML Link splits to: []string{"https:", "", "pypi.org", "project", "foopy", "2.1", ""}
	parts := strings.Split(p.Link, "/")
	if len(parts) < 5 {
		return "", errInvalidLinkForPackage
	}
	return parts[len(parts)-2], nil
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

func fetchPackages(baseURL string) ([]*Package, error) {
	pkgURL, err := utils.URLPathJoin(baseURL, updatesPath)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Get(pkgURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rssResponse := &Response{}
	reader := utils.NewUTF8OnlyReader(resp.Body)
	err = xml.NewDecoder(reader).Decode(rssResponse)
	if err != nil {
		return nil, err
	}
	return rssResponse.Packages, nil
}

func fetchCriticalPackages(baseURL string, packageList []string) ([]*Package, error) {
	responseChannel := make(chan *Response)
	errChannel := make(chan error)

	for _, pkgName := range packageList {
		go func(pkgName string) {
			packageDataPath := fmt.Sprintf(packagePathFormat, pkgName)
			pkgURL, err := utils.URLPathJoin(baseURL, packageDataPath)
			if err != nil {
				errChannel <- err
				return
			}
			resp, err := httpClient.Get(pkgURL)
			if err != nil {
				errChannel <- err
				return
			}
			defer resp.Body.Close()
			rssResponse := &Response{}
			reader := utils.NewUTF8OnlyReader(resp.Body)
			err = xml.NewDecoder(reader).Decode(rssResponse)
			if err != nil {
				errChannel <- err
				return
			}

			responseChannel <- rssResponse
		}(pkgName)
	}

	pkgs := []*Package{}
	for i := 0; i < len(packageList); i++ {
		select {
		case response := <-responseChannel:
			pkgs = append(pkgs, response.Packages...)
		case err := <-errChannel:
			return nil, err
		}
	}
	return pkgs, nil
}

type Feed struct {
	packages *[]string

	lossyFeedAlerter *feeds.LossyFeedAlerter
	baseURL          string
}

func New(feedOptions feeds.FeedOptions, eventHandler *events.Handler) (*Feed, error) {
	return &Feed{
		packages:         feedOptions.Packages,
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
		baseURL:          "https://pypi.org/",
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	var pypiPackages []*Package
	var err error

	if feed.packages == nil {
		// Firehose fetch all packages.
		pypiPackages, err = fetchPackages(feed.baseURL)
	} else {
		// Fetch specific packages individually from configured packages list.
		pypiPackages, err = fetchCriticalPackages(feed.baseURL, *feed.packages)
	}

	if err != nil {
		return nil, err
	}
	for _, pkg := range pypiPackages {
		pkgName, err := pkg.Name()
		if err != nil {
			return nil, err
		}
		pkgVersion, err := pkg.Version()
		if err != nil {
			return nil, err
		}
		pkg := feeds.NewPackage(pkg.CreatedDate.Time, pkgName, pkgVersion, FeedName)
		pkgs = append(pkgs, pkg)
	}

	// Lossy feed detection is only necessary for firehose fetching
	if feed.packages == nil {
		feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)
	}

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}

func (feed Feed) GetPackageList() *[]string {
	return feed.packages
}
