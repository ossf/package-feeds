package utils

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrUnsuccessfulRequest = errors.New("unsuccessful request")

func CheckResponseStatus(res *http.Response) error {
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("%w: %v", ErrUnsuccessfulRequest, res.Status)
	}
	return nil
}
