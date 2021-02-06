package feeds

import "time"

type ScheduledFeed interface {
	Latest(cutoff time.Time) ([]*Package, error)
}

type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedDate time.Time `json:"created_date"`
	Type        string    `json:"type"`
}
