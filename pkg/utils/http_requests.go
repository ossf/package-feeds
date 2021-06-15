package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

var ErrUnsuccessfulRequest = errors.New("unsuccessful request")

func CheckResponseStatus(res *http.Response) error {
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("%w: %v", ErrUnsuccessfulRequest, res.Status)
	}
	return nil
}

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
