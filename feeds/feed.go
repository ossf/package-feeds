package feeds

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const schemaVer = "1.0"

var errCutoff = errors.New("package was created before cutoff time")

type ScheduledFeed interface {
	Latest(cutoff time.Time) ([]*Package, error)
}

// Marshalled json output validated against package.schema.json.
type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"created_date"`
	Type        string    `json:"type"`
	SchemaVer   string    `json:"schema_ver"`
}

func NewPackage(created, cutoff time.Time, name, version, feed string) (*Package, error) {
	if created.Before(cutoff) {
		return nil, fmt.Errorf("%w : %s", errCutoff, cutoff.String())
	}
	log.WithFields(log.Fields{
		"feed":    feed,
		"name":    name,
		"version": version,
	}).Print("Processing Package")
	pkg := &Package{
		Name:        name,
		Version:     version,
		CreatedDate: created,
		Type:        feed,
		SchemaVer:   schemaVer,
	}
	return pkg, nil
}
