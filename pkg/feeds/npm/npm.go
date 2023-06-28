package npm

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/utils"
)

const (
	FeedName = "npm"
	rssPath  = "/-/rss"

	// rssLimit controls how many RSS results should be returned.
	// Can up to about 420 before the feed will consistently fail to return any data.
	// Lower numbers will sometimes fail too. Default value if not specified is 50.
	rssLimit = 200

	// fetchWorkers defines the totoal number of concurrent HTTP1 requests to
	// allow at any one time.
	fetchWorkers = 10

	// cacheEntryLimit defines how many responses to store in the LRU cache.
	// The value should be larger than rssLimit to ensure all rss entries can
	// be covered by a cache entry.
	cacheEntryLimit = 500
)

var (
	errJSON        = errors.New("error unmarshaling json response internally")
	errUnpublished = errors.New("package is currently unpublished")
)

type Response struct {
	PackageEvents []PackageEvent `xml:"channel>item"`
}

type Package struct {
	Title       string
	CreatedDate time.Time
	Version     string
	Unpublished bool
}

type PackageEvent struct {
	Title string `xml:"title"`
}

type cacheEntry struct {
	ETag     string
	Versions []*Package
}

// Returns a slice of PackageEvent{} structs.
func fetchPackageEvents(feed Feed) ([]PackageEvent, error) {
	pkgURL, err := url.Parse(feed.baseURL)
	if err != nil {
		return nil, err
	}

	pkgURL = pkgURL.JoinPath(rssPath)
	q := pkgURL.Query()
	q.Set("limit", fmt.Sprintf("%d", rssLimit))
	pkgURL.RawQuery = q.Encode()

	resp, err := feed.client.Get(pkgURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := utils.CheckResponseStatus(resp); err != nil {
		return nil, fmt.Errorf("failed to fetch npm package data: %w", err)
	}

	rssResponse := &Response{}
	reader := utils.NewXMLReader(resp.Body, true)
	if err := xml.NewDecoder(reader).Decode(rssResponse); err != nil {
		return nil, err
	}

	return rssResponse.PackageEvents, nil
}

// Gets the package version & corresponding created date from NPM. Returns
// a slice of {}Package.
func fetchPackage(feed Feed, pkgTitle string) ([]*Package, error) {
	versionURL, err := url.JoinPath(feed.baseURL, pkgTitle)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", versionURL, nil)
	if err != nil {
		return nil, err
	}

	e, inCache := feed.cache.Get(versionURL)
	if inCache && e != nil {
		req.Header.Add("If-None-Match", e.ETag)
	}

	resp, err := feed.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()

	if inCache && e != nil && utils.IsNotModified(resp) {
		return e.Versions, nil
	}
	if err := utils.CheckResponseStatus(resp); err != nil {
		return nil, fmt.Errorf("failed to fetch npm package version data: %w", err)
	}

	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	etag := resp.Header.Get("etag")

	// We only care about the `time` field as it contains all the versions in
	// date order, from oldest to newest.
	// Using a struct for parsing also avoids the cost of deserializing data
	// that is ultimately unused.
	var packageDetails struct {
		Time map[string]interface{} `json:"time"`
	}

	if err := json.Unmarshal(body, &packageDetails); err != nil {
		return nil, fmt.Errorf("%w : %w for package %s", errJSON, err, pkgTitle)
	}
	versions := packageDetails.Time

	// If `unpublished` exists in the version map then at a given point in time
	// the package was 'entirely' removed, the packageEvent(s) received are for package
	// versions that no longer exist. For a given 24h period no further versions can
	// be uploaded, with any previous versions never being available again.
	// https://www.npmjs.com/policies/unpublish
	_, unPublished := versions["unpublished"]

	if unPublished {
		return nil, fmt.Errorf("%s %w", pkgTitle, errUnpublished)
	}

	// Remove redundant entries in map, we're only interested in actual version pairs.
	delete(versions, "created")
	delete(versions, "modified")

	// Create slice of Package{} to allow sorting of a slice, as maps
	// are unordered.
	versionSlice := []*Package{}
	for version, timestamp := range versions {
		date, err := time.Parse(time.RFC3339, timestamp.(string))
		if err != nil {
			return nil, err
		}
		versionSlice = append(versionSlice,
			&Package{Title: pkgTitle, CreatedDate: date, Version: version})
	}

	// Sort slice of versions into order of most recent.
	sort.SliceStable(versionSlice, func(i, j int) bool {
		return versionSlice[j].CreatedDate.Before(versionSlice[i].CreatedDate)
	})

	if etag != "" {
		feed.cache.Add(versionURL, &cacheEntry{
			ETag:     etag,
			Versions: versionSlice,
		})
	}

	return versionSlice, nil
}

func fetchAllPackages(feed Feed) ([]*feeds.Package, []error) {
	pkgs := []*feeds.Package{}
	errs := []error{}
	packageChannel := make(chan []*Package)
	errChannel := make(chan error)
	packageEvents, err := fetchPackageEvents(feed)
	if err != nil {
		// If we can't generate package events then return early.
		return pkgs, append(errs, err)
	}
	// Handle the possibility of multiple releases of the same package
	// within the polled `packages` slice.
	uniquePackages := make(map[string]int)
	for _, pkg := range packageEvents {
		uniquePackages[pkg.Title]++
	}

	// Start a collection of workers to fetch all the packages.
	// This limits the number of concurrent requests to avoid flooding the NPM
	// registry API with too many simultaneous requests.
	workChannel := make(chan struct {
		pkgTitle string
		count    int
	})

	// Define the fetcher function that grabs the repos from NPM
	fetcherFn := func(pkgTitle string, count int) {
		pkgs, err := fetchPackage(feed, pkgTitle)
		if err != nil {
			if !errors.Is(err, errUnpublished) {
				err = feeds.PackagePollError{Name: pkgTitle, Err: err}
			}
			errChannel <- err
			return
		}
		// Apply count slice, guard against a given events corresponding
		// version entry being unpublished by the time the specific
		// endpoint has been processed. This seemingly happens silently
		// without being recorded in the json. An `event` could be logged
		// here.
		if len(pkgs) > count {
			packageChannel <- pkgs[:count]
		} else {
			packageChannel <- pkgs
		}
	}

	// The WaitGroup is used to ensure all the goroutines are complete before
	// returning.
	var wg sync.WaitGroup

	// Start the fetcher workers.
	for i := 0; i < fetchWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				w, more := <-workChannel
				if !more {
					// If we have no more work then return.
					return
				}
				fetcherFn(w.pkgTitle, w.count)
			}
		}()
	}

	// Start a goroutine to push work to the workers.
	go func() {
		// Populate the worker feed.
		for pkgTitle, count := range uniquePackages {
			workChannel <- struct {
				pkgTitle string
				count    int
			}{pkgTitle: pkgTitle, count: count}
		}

		// Close the channel to indicate that there is no more work.
		close(workChannel)
	}()

	// Collect all the work.
	for i := 0; i < len(uniquePackages); i++ {
		select {
		case npmPkgs := <-packageChannel:
			for _, pkg := range npmPkgs {
				feedPkg := feeds.NewPackage(pkg.CreatedDate, pkg.Title,
					pkg.Version, FeedName)
				pkgs = append(pkgs, feedPkg)
			}
		case err := <-errChannel:
			// When polling the 'firehose' unpublished packages
			// don't need to be logged as an error.
			if !errors.Is(err, errUnpublished) {
				errs = append(errs, err)
			}
		}
	}

	wg.Wait()

	return pkgs, errs
}

func fetchCriticalPackages(feed Feed, packages []string) ([]*feeds.Package, []error) {
	pkgs := []*feeds.Package{}
	errs := []error{}
	packageChannel := make(chan []*Package)
	errChannel := make(chan error)

	for _, pkgTitle := range packages {
		go func(pkgTitle string) {
			pkgs, err := fetchPackage(feed, pkgTitle)
			if err != nil {
				if !errors.Is(err, errUnpublished) {
					err = feeds.PackagePollError{Name: pkgTitle, Err: err}
				}
				errChannel <- err
				return
			}
			packageChannel <- pkgs
		}(pkgTitle)
	}

	for i := 0; i < len(packages); i++ {
		select {
		case npmPkgs := <-packageChannel:
			for _, pkg := range npmPkgs {
				feedPkg := feeds.NewPackage(pkg.CreatedDate, pkg.Title,
					pkg.Version, FeedName)
				pkgs = append(pkgs, feedPkg)
			}
		case err := <-errChannel:
			// Assume if a package has been unpublished that it is a valid reason
			// to log the error when polling for 'critical' packages. This could
			// be changed for a 'lossy' type event instead. Further packages should
			// be proccessed.
			errs = append(errs, err)
		}
	}
	return pkgs, errs
}

type Feed struct {
	packages         *[]string
	lossyFeedAlerter *feeds.LossyFeedAlerter
	baseURL          string
	options          feeds.FeedOptions
	client           *http.Client
	cache            *lru.Cache[string, *cacheEntry]
}

func New(feedOptions feeds.FeedOptions, eventHandler *events.Handler) (*Feed, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	// Disable HTTP2. HTTP2 flow control hurts performance for large concurrent
	// responses.
	tr.ForceAttemptHTTP2 = false
	tr.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	tr.TLSClientConfig.NextProtos = []string{"http/1.1"}

	tr.MaxIdleConns = 100
	tr.MaxIdleConnsPerHost = fetchWorkers
	tr.MaxConnsPerHost = fetchWorkers
	tr.IdleConnTimeout = 0 // No limit, try and reuse the idle connecitons.

	cache, err := lru.New[string, *cacheEntry](cacheEntryLimit)
	if err != nil {
		return nil, err
	}

	return &Feed{
		packages:         feedOptions.Packages,
		lossyFeedAlerter: feeds.NewLossyFeedAlerter(eventHandler),
		baseURL:          "https://registry.npmjs.org/",
		options:          feedOptions,
		client: &http.Client{
			Transport: tr,
			Timeout:   45 * time.Second,
		},
		cache: cache,
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, []error) {
	var pkgs []*feeds.Package
	var errs []error

	if feed.packages == nil {
		pkgs, errs = fetchAllPackages(feed)
	} else {
		pkgs, errs = fetchCriticalPackages(feed, *feed.packages)
	}

	if len(pkgs) == 0 {
		// If none of the packages were successfully polled for, return early.
		return nil, append(errs, feeds.ErrNoPackagesPolled)
	}

	// Ensure packages are sorted by CreatedDate in order of most recent, as goroutine
	// concurrency isn't deterministic.
	sort.SliceStable(pkgs, func(i, j int) bool {
		return pkgs[j].CreatedDate.Before(pkgs[i].CreatedDate)
	})

	// TODO: Add an event for checking if the previous package list contains entries
	// that do not exist in the latest package list when polling for critical packages.
	// This can highlight cases where specific versions have been unpublished.
	if feed.packages == nil {
		feed.lossyFeedAlerter.ProcessPackages(FeedName, pkgs)
	}

	pkgs = feeds.ApplyCutoff(pkgs, cutoff)
	return pkgs, errs
}

func (feed Feed) GetName() string {
	return FeedName
}

func (feed Feed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}
