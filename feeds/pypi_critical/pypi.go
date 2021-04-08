package pypi_critical

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

const (
	FeedName = "pypi_critical"
)

var (
	packageUrlFormat = "https://pypi.org/rss/project/%s/releases.xml"
	httpClient       = &http.Client{
		Timeout: 10 * time.Second,
	}
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
	pkgName  string
}

type Package struct {
	Title       string      `xml:"title"`
	CreatedDate rfc1123Time `xml:"pubDate"`
	Link        string      `xml:"link"`
}

func (p *Package) Name() (string, error) {
	// The XML Link ends with /packageName/Version/
	parts := strings.Split(p.Link, "/")
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid link provided by pypi releases API")
	}
	return parts[len(parts)-3], nil
}

func (p *Package) Version() string {
	// The XML Feed has a "Title" element that contains the release version.
	return p.Title
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

func fetchPackages(packageList []string) ([]*Package, error) {
	responseChannel := make(chan *Response)
	errs := make(chan error)
	var wg sync.WaitGroup

	for _, pkgName := range packageList {
		wg.Add(1)
		go func(pkgName string) {
			resp, err := httpClient.Get(fmt.Sprintf(packageUrlFormat, pkgName))
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()
			rssResponse := &Response{}
			err = xml.NewDecoder(resp.Body).Decode(rssResponse)
			if err != nil {
				errs <- err
				return
			}
			wg.Done()
			responseChannel <- rssResponse
		}(pkgName)
	}
	wg.Wait()
	select {
	case err := <-errs:
		return nil, err
	default:
		close(errs)
	}

	pkgs := []*Package{}
	for i := 0; i < len(packageList); i++ {
		select {
		case response := <-responseChannel:
			for _, pkg := range response.Packages {
				pkgs = append(pkgs, pkg)
			}
		default:
			return nil, fmt.Errorf("unexpected missing package release response")
		}
	}
	return pkgs, nil
}

type Feed struct {
	PackageList []string `mapstructure:"packages"`
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	pypiPackages, err := fetchPackages(feed.PackageList)
	if err != nil {
		return pkgs, err
	}
	for _, pkg := range pypiPackages {
		if pkg.CreatedDate.Before(cutoff) {
			continue
		}
		pkgName, err := pkg.Name()
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &feeds.Package{
			Name:        pkgName,
			Version:     pkg.Version(),
			CreatedDate: pkg.CreatedDate.Time,
			Type:        FeedName,
		})
	}
	return pkgs, nil
}
