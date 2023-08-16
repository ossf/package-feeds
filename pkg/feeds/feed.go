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

// Cutoff is an interface used to define orering of package updates.
//
// The type *time.Time satisifies this interface.
type Cutoff[T any] interface {
	Before(Cutoff[T]) bool
}

type ScheduledFeed interface {
	Latest(cutoff time.Time) ([]*Package, time.Time, []error)
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
	ArtifactID  string    `json:"artifact_id"`
	SchemaVer   string    `json:"schema_ver"`
}

type PackagePollError struct {
	Err  error
	Name string
}

func (err PackagePollError) Error() string {
	return fmt.Sprintf("Polling for package %s returned error: %v", err.Name, err.Err)
}

// NewPackage creates a Package object without the artifact ID field populated.
func NewPackage(created time.Time, name, version, feed string) *Package {
	return NewArtifact(created, name, version, "", feed)
}

// NewArtifact creates a Package object with the artifact ID field populated.
func NewArtifact(created time.Time, name, version, artifactID, feed string) *Package {
	return &Package{
		Name:        name,
		Version:     version,
		CreatedDate: created,
		Type:        feed,
		ArtifactID:  artifactID,
		SchemaVer:   schemaVer,
	}
}

func ApplyCutoff(pkgs []*Package, cutoff time.Time) []*Package {
	filteredPackages := []*Package{}
	for _, pkg := range pkgs {
		if pkg.CreatedDate.After(cutoff) {
			filteredPackages = append(filteredPackages, pkg)
		}
	}
	return filteredPackages
}

func FindCutoff(cutoff time.Time, pkgs []*Package) time.Time {
	for _, pkg := range pkgs {
		if pkg.CreatedDate.After(cutoff) {
			cutoff = pkg.CreatedDate
		}
	}
	return cutoff
}

func (err UnsupportedOptionError) Error() string {
	return fmt.Sprintf("unsupported option `%v` supplied to %v feed", err.Option, err.Feed)
}
