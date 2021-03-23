package packagist

import "time"

type versionInfo struct {
	Version           string    `json:"version"`
	VersionNormalized string    `json:"version_normalized"`
	License           []string  `json:"license,omitempty"`
	Time              time.Time `json:"time"`
	Name              string    `json:"name,omitempty"`
}
type packages struct {
	Packages map[string][]versionInfo `json:"packages"`
}
