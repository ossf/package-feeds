package utils

import (
	"fmt"
	"net/url"
	"path"
)

func URLPathJoin(baseURL string, paths ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	pathParts := []string{u.Path}
	pathParts = append(pathParts, paths...)
	u.Path = path.Join(pathParts...)
	return u.String(), nil
}
