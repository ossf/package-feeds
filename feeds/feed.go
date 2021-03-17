package feeds

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type ScheduledFeed interface {
	Latest(cutoff time.Time) ([]*Package, error)
}

type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"created_date"`
	Type        string    `json:"type"`
}

func NewPackage(created, cutoff time.Time, name, version, feed string) (*Package, error) {
	if created.Before(cutoff) {
		return nil, fmt.Errorf("package was created before cutoff time: %s", cutoff.String())
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
	}
	return pkg, nil
}
