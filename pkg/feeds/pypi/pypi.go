package pypi

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/useragent"
	"github.com/ossf/package-feeds/pkg/utils"
)

const (
	FeedName          = "pypi"
	updatesPath       = "/rss/updates.xml"
	packagePathFormat = "/rss/project/%s/releases.xml"
)

var (
	httpClient = &http.Client{
		Transport: &useragent.RoundTripper{UserAgent: feeds.DefaultUserAgent},
		Timeout:   10 * time.Second,
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
	pkgURL, err := url.JoinPath(baseURL, updatesPath)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Get(pkgURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.CheckResponseStatus(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pypi package data: %w", err)
	}

	rssResponse := &Response{}
	reader := utils.NewXMLReader(resp.Body, true)
	err = xml.NewDecoder(reader).Decode(rssResponse)
	if err != nil {
		return nil, err
	}
	return rssResponse.Packages, nil
}

func fetchCriticalPackages(baseURL string, packageList []string) ([]*Package, []error) {
	responseChannel := make(chan *Response)
	errChannel := make(chan error)

	for _, pkgName := range packageList {
		go func(pkgName string) {
			packageDataPath := fmt.Sprintf(packagePathFormat, pkgName)
			pkgURL, err := url.JoinPath(baseURL, packageDataPath)
			if err != nil {
				errChannel <- feeds.PackagePollError{Name: pkgName, Err: err}
				return
			}
			resp, err := httpClient.Get(pkgURL)
			if err != nil {
				errChannel <- feeds.PackagePollError{Name: pkgName, Err: err}
				return
			}
			defer resp.Body.Close()

			err = utils.CheckResponseStatus(resp)
			if err != nil {
				errChannel <- feeds.PackagePollError{Name: pkgName, Err: fmt.Errorf("failed to fetch pypi package data: %w", err)}
				return
			}

			rssResponse := &Response{}
			reader := utils.NewXMLReader(resp.Body, true)
			err = xml.NewDecoder(reader).Decode(rssResponse)
			if err != nil {
				errChannel <- feeds.PackagePollError{Name: pkgName, Err: err}
				return
			}

			responseChannel <- rssResponse
		}(pkgName)
	}

	pkgs := []*Package{}
	errs := []error{}
	for i := 0; i < len(packageList); i++ {
		select {
		case response := <-responseChannel:
			pkgs = append(pkgs, response.Packages...)
		case err := <-errChannel:
			errs = append(errs, err)
		}
	}
	return pkgs, errs
}

type Feed struct {
	packages *[]string

	lossyFeedAlerter *feeds.LossyFeedAlerter
	baseURL          string

	options feeds.FeedOptions
}

func New(feedOptions feeds.FeedOptions, eventHandler *events.Handler) (*Feed, error) {
	return &Feed{
		packages:         feedOptions.Packages,
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
		baseURL:          "https://pypi.org/",
		options:          feedOptions,
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, time.Time, []error) {
	pkgs := []*feeds.Package{}
	var pypiPackages []*Package
	var errs []error
	var err error

	if feed.packages == nil {
		// Firehose fetch all packages.
		// If this fails then we need to return, as it's the only source of
		// data.
		pypiPackages, err = fetchPackages(feed.baseURL)
		if err != nil {
			return nil, cutoff, append(errs, err)
		}
	} else {
		// Fetch specific packages individually from configured packages list.
		pypiPackages, errs = fetchCriticalPackages(feed.baseURL, *feed.packages)
		if len(pypiPackages) == 0 {
			// If none of the packages were successfully polled for, return early.
			return nil, cutoff, append(errs, feeds.ErrNoPackagesPolled)
		}
	}

	for _, pkg := range pypiPackages {
		pkgName, err := pkg.Name()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		pkgVersion, err := pkg.Version()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		pkg := feeds.NewPackage(pkg.CreatedDate.Time, pkgName, pkgVersion, FeedName)
		pkgs = append(pkgs, pkg)
	}

	// Lossy feed detection is only necessary for firehose fetching
	if feed.packages == nil {
		feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)
	}

	newCutoff := feeds.FindCutoff(cutoff, pkgs)
	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, newCutoff, errs
}

func (feed Feed) GetPackageList() *[]string {
	return feed.packages
}

func (feed Feed) GetName() string {
	return FeedName
}

func (feed Feed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}
