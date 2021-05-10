package npm

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/ossf/package-feeds/events"
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

type Feed struct {
	lossyFeedAlerter *feeds.LossyFeedAlerter
}

func New(feedOptions feeds.FeedOptions, eventHandler *events.Handler) (*Feed, error) {
	if feedOptions.Packages != nil {
		return nil, feeds.UnsupportedOptionError{
			Feed:   FeedName,
			Option: "packages",
		}
	}
	return &Feed{
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	pkgs := []*feeds.Package{}
	packageChannel := make(chan *feeds.Package)
	errs := make(chan error)

	packages, err := fetchPackages()
	if err != nil {
		return pkgs, err
	}

	for _, pkg := range packages {
		go func(pkg *Package) {
			v, err := fetchVersionInformation(pkg.Title)
			if err != nil {
				errs <- err
				return
			}
			feedPkg := feeds.NewPackage(pkg.CreatedDate.Time, pkg.Title, v, FeedName)
			packageChannel <- feedPkg
		}(pkg)
	}

	for i := 0; i < len(packages); i++ {
		select {
		case pkg := <-packageChannel:
			pkgs = append(pkgs, pkg)
		case err := <-errs:
			return pkgs, fmt.Errorf("error in fetching version information: %w", err)
		}
	}

	// Ensure packages are sorted by CreatedDate in order of most recent, as goroutine
	// concurrency isn't deterministic.
	sort.SliceStable(pkgs, func(i, j int) bool {
		return pkgs[j].CreatedDate.Before(pkgs[i].CreatedDate)
	})

	feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, nil
}
