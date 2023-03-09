package pypi2

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/kolo/xmlrpc"

	"github.com/ossf/package-feeds/pkg/feeds"
)

const (
	FeedName = "pypi-v2"
)

var (
	// we care about changelog entries where the action is 'add X file <filename>'
	archiveUploadAction = regexp.MustCompile("add (.*) file (.*)")
)

type Feed struct {
	packages *[]string
	baseURL  string
	options  feeds.FeedOptions
}

func New(feedOptions feeds.FeedOptions) (*Feed, error) {
	return &Feed{
		packages: feedOptions.Packages,
		baseURL:  "https://pypi.org/pypi",
		options:  feedOptions,
	}, nil
}

func (feed Feed) Latest(cutoff time.Time) ([]*feeds.Package, []error) {
	client, _ := xmlrpc.NewClient(feed.baseURL, nil)

	// Raw result structure is array[array[string, string, int, string, int]]
	// which cannot be represented in Go (struct mapping is not supported by library)
	var result [][]any

	cutoffSeconds := cutoff.UnixNano() / 1_000_000_000
	if err := client.Call("changelog", []interface{}{cutoffSeconds, true}, &result); err != nil {
		return nil, []error{err}
	}

	var errs []error
	changelogEntries := make([]pypiChangelogEntry, len(result))
	for i, r := range result {
		e, err := processRawChangelogData(r)
		if err != nil {
			// ignore for now and keep processing
			errs = append(errs, err)
		}
		changelogEntries[i] = e
	}

	var pkgs []*feeds.Package
	for _, e := range changelogEntries {
		if e.isArchiveUpload() {
			pkg := feeds.NewPackageArchive(e.Timestamp, e.Name, e.Version, e.ArchiveName, FeedName)
			pkgs = append(pkgs, pkg)
		}

	}

	return pkgs, errs
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

type pypiChangelogEntry struct {
	Name        string
	Version     string
	Timestamp   time.Time
	Action      string
	ArchiveName string
	Id          int64
}

func (e pypiChangelogEntry) isArchiveUpload() bool {
	return e.ArchiveName != ""
}

func (e pypiChangelogEntry) String() string {
	return fmt.Sprintf("#%d %s (%s): %s ts=%d", e.Id, e.Name, e.Version, e.Action, e.Timestamp)
}

func processRawChangelogData(data []any) (pypiChangelogEntry, error) {
	// Each element of the inner array consists of:
	// name: string
	// version: string
	// timestamp: int
	// action: string
	// id: int
	entry := pypiChangelogEntry{
		Name:    fmt.Sprint(data[0]),
		Version: fmt.Sprint(data[1]),
		Action:  fmt.Sprint(data[3]),
	}

	// Changelog entries corresponding to new archives being added have an
	// action string that looks like 'add X file <filename>'
	if match := archiveUploadAction.FindStringSubmatch(entry.Action); match != nil {
		// it's a new archive!
		entry.ArchiveName = match[2]
	}

	// If timestamp cannot be parsed as an integer, a nil value (0) will be recorded
	timestampString := fmt.Sprint(data[2])
	if t, err := strconv.ParseInt(timestampString, 10, 64); err != nil {
		return entry, fmt.Errorf("cannot parse %q as timestamp: %w", timestampString, err)
	} else {
		entry.Timestamp = time.Unix(t, 0)
	}

	// If ID cannot be parsed as an integer, a nil value (0) will be recorded
	idString := fmt.Sprint(data[4])
	if id, err := strconv.ParseInt(idString, 10, 64); err != nil {
		return entry, fmt.Errorf("cannot parse %q as ID: %w", idString, err)
	} else {
		entry.Id = id

	}

	return entry, nil
}
