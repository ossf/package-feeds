package pypi

import (
	"fmt"
	"regexp"
	"time"

	"github.com/kolo/xmlrpc"

	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/useragent"
)

const (
	ArtifactFeedName = "pypi-artifacts"
)

// We care about changelog entries where the action is 'add X file <filename>'.
var archiveUploadAction = regexp.MustCompile("add (.*) file (.*)")

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

func (feed ArtifactFeed) Latest(cutoff time.Time) ([]*feeds.Package, time.Time, []error) {
	client, err := xmlrpc.NewClient(feed.baseURL, &useragent.RoundTripper{UserAgent: feeds.DefaultUserAgent})
	if err != nil {
		return nil, cutoff, []error{err}
	}

	changelogEntries, err := getPyPIChangeLog(client, cutoff)
	if err != nil {
		return nil, cutoff, []error{err}
	}

	pkgs := getUploadedArtifacts(changelogEntries)
	return pkgs, feeds.FindCutoff(cutoff, pkgs), nil
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
}

func (e *pypiChangelogEntry) isArchiveUpload() bool {
	return e.ArchiveName != ""
}

func (e *pypiChangelogEntry) String() string {
	return fmt.Sprintf("%s (%s): %s ts=%s", e.Name, e.Version, e.Action, e.Timestamp)
}

// getPyPIChangeLog returns a list of PyPI changelog entries since the given timestamp
// defined by https://warehouse.pypa.io/api-reference/xml-rpc.html#changelog-since-with-ids-false
func getPyPIChangeLog(client *xmlrpc.Client, since time.Time) ([]pypiChangelogEntry, error) {
	// Raw result structure is array[array[string, string|nil, int64, string (, int64 if with_ids=true) ]]
	// which cannot be represented in Go (struct mapping is not supported by library)
	var result [][]interface{}
	if err := client.Call("changelog", []interface{}{since.Unix(), false}, &result); err != nil {
		return nil, err
	}

	return processRawChangelog(result), nil
}

func processRawChangelog(apiResult [][]interface{}) []pypiChangelogEntry {
	changelogEntries := make([]pypiChangelogEntry, len(apiResult))
	for i, r := range apiResult {
		changelogEntries[i] = processRawChangelogItem(r)
	}

	return changelogEntries
}

func processRawChangelogItem(data []interface{}) pypiChangelogEntry {
	/*
		Each item of the changelog contains the following fields:
		name: string
		version: string (nullable)
		timestamp: int64
		action: string
	*/
	name, ok := data[0].(string)
	if !ok {
		name = ""
	}
	version, ok := data[1].(string)
	if !ok {
		version = ""
	}
	unixTimestamp, ok := data[2].(int64)
	if !ok {
		unixTimestamp = 0
	}
	action, ok := data[3].(string)
	if !ok {
		action = ""
	}

	archiveName := ""
	// Changelog entries corresponding to new archives being added have an action string
	// that looks like 'add <archive type> file <archive name>'. This code is generated by
	// github.com/pypi/warehouse/blob/3bfd3e0d32e7396582d9635a316f1a47e407304d/warehouse/forklift/legacy.py#L1355
	if match := archiveUploadAction.FindStringSubmatch(action); match != nil {
		// it's a new archive!
		archiveName = match[2]
	}

	return pypiChangelogEntry{
		Name:        name,
		Version:     version,
		Timestamp:   time.Unix(unixTimestamp, 0),
		Action:      action,
		ArchiveName: archiveName,
	}
}

func getUploadedArtifacts(changelogEntries []pypiChangelogEntry) []*feeds.Package {
	var pkgs []*feeds.Package
	for _, e := range changelogEntries {
		if e.isArchiveUpload() {
			pkgs = append(pkgs, feeds.NewArtifact(e.Timestamp, e.Name, e.Version, e.ArchiveName, ArtifactFeedName))
		}
	}

	return pkgs
}
