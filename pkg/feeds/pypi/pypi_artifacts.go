package pypi

import (
	"fmt"
	"regexp"
	"time"

	"github.com/kolo/xmlrpc"

	"github.com/ossf/package-feeds/pkg/feeds"
)

const (
	ArtifactFeedName = "pypi-artifacts"
)

var (
	// we care about changelog entries where the action is 'add X file <filename>'
	archiveUploadAction = regexp.MustCompile("add (.*) file (.*)")
)

type ArtifactFeed struct {
	baseURL string
	options feeds.FeedOptions
}

func NewArtifactFeed(feedOptions feeds.FeedOptions) (*ArtifactFeed, error) {
	return &ArtifactFeed{
		baseURL: "https://pypi.org/pypi",
		options: feedOptions,
	}, nil
}

func (feed ArtifactFeed) Latest(cutoff time.Time) ([]*feeds.Package, []error) {
	cutoffSeconds := cutoff.UnixNano() / 1_000_000_000

	client, _ := xmlrpc.NewClient(feed.baseURL, nil)

	// Raw result structure is array[array[string, string, int, string, int]]
	// which cannot be represented in Go (struct mapping is not supported by library)
	var result [][]any
	if err := client.Call("changelog", []interface{}{cutoffSeconds, true}, &result); err != nil {
		return nil, []error{err}
	}

	changelogEntries := make([]pypiChangelogEntry, len(result))
	for i, r := range result {
		changelogEntries[i] = processRawChangelogItem(r)
	}

	var pkgs []*feeds.Package
	for _, e := range changelogEntries {
		if e.isArchiveUpload() {
			pkgs = append(pkgs, feeds.NewPackageArchive(e.Timestamp, e.Name, e.Version, e.ArchiveName, ArtifactFeedName))
		}

	}

	return pkgs, nil
}

func (feed ArtifactFeed) GetFeedOptions() feeds.FeedOptions {
	return feed.options
}
func (feed ArtifactFeed) GetName() string {
	return ArtifactFeedName
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
	return fmt.Sprintf("#%d %s (%s): %s ts=%s", e.Id, e.Name, e.Version, e.Action, e.Timestamp)
}

func processRawChangelogItem(data []any) pypiChangelogEntry {
	/*
		Each item of the changelog contains the following fields
		name: string
		version: string
		timestamp: int64
		action: string
		id: int64
	*/
	entry := pypiChangelogEntry{
		Name:      data[0].(string),
		Version:   data[1].(string),
		Timestamp: time.Unix(data[2].(int64), 0),
		Action:    data[3].(string),
		Id:        data[4].(int64),
	}

	// Changelog entries corresponding to new archives being added
	// have an action string that looks like 'add <archive type> file <archive name>'
	if match := archiveUploadAction.FindStringSubmatch(entry.Action); match != nil {
		// it's a new archive!
		entry.ArchiveName = match[2]
	}

	return entry
}
