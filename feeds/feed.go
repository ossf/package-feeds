package feeds

import (
	"fmt"
	"time"
)

const schemaVer = "1.0"

type UnsupportedOptionError struct {
	Option string
	Feed   string
}

type ScheduledFeed interface {
	Latest(cutoff time.Time) ([]*Package, error)
	GetFeedOptions() FeedOptions
	GetName() string
}

// General configuration options for feeds.
type FeedOptions struct {
	// A collection of package names to poll instead of standard firehose behaviour.
	// Not supported by all feeds.
	Packages *[]string `yaml:"packages"`
}

// Marshalled json output validated against package.schema.json.
type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"created_date"`
	Type        string    `json:"type"`
	SchemaVer   string    `json:"schema_ver"`
}

func NewPackage(created time.Time, name, version, feed string) *Package {
	return &Package{
		Name:        name,
		Version:     version,
		CreatedDate: created,
		Type:        feed,
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
