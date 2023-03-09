package feeds

import (
	"errors"
	"fmt"
	"time"
)

const schemaVer = "1.1"

var ErrNoPackagesPolled = errors.New("no packages were successfully polled")

type UnsupportedOptionError struct {
	Option string
	Feed   string
}

type ScheduledFeed interface {
	Latest(cutoff time.Time) ([]*Package, []error)
	GetFeedOptions() FeedOptions
	GetName() string
}

// General configuration options for feeds.
type FeedOptions struct {
	// A collection of package names to poll instead of standard firehose behaviour.
	// Not supported by all feeds.
	Packages *[]string `yaml:"packages"`

	// Cron string for scheduling the polling for the feed.
	PollRate string `yaml:"poll_rate"`
}

// Marshalled json output validated against package.schema.json.
type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"created_date"`
	Type        string    `json:"type"`
	ArchiveName string    `json:"archive_name"`
	SchemaVer   string    `json:"schema_ver"`
}

type PackagePollError struct {
	Err  error
	Name string
}

func (err PackagePollError) Error() string {
	return fmt.Sprintf("Polling for package %s returned error: %v", err.Name, err.Err)
}

func NewPackage(created time.Time, name, version, feed string) *Package {
	return NewPackageArchive(created, name, version, "", feed)
}

func NewPackageArchive(created time.Time, name, version, archiveName, feed string) *Package {
	return &Package{
		Name:        name,
		Version:     version,
		CreatedDate: created,
		Type:        feed,
		ArchiveName: archiveName,
		SchemaVer:   schemaVer,
	}
}

func ApplyCutoff(pkgs []*Package, cutoff time.Time) []*Package {
	filteredPackages := []*Package{}
	for _, pkg := range pkgs {
		if !pkg.CreatedDate.Before(cutoff) {
			filteredPackages = append(filteredPackages, pkg)
		}
	}
	return filteredPackages
}

func (err UnsupportedOptionError) Error() string {
	return fmt.Sprintf("unsupported option `%v` supplied to %v feed", err.Option, err.Feed)
}
