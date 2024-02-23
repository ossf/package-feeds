package useragent

import "net/http"

type RoundTripper struct {
	UserAgent string
	Parent    http.RoundTripper
}

func (rt *RoundTripper) RoundTrip(ireq *http.Request) (*http.Response, error) {
	req := ireq.Clone(ireq.Context())
	req.Header.Set("User-Agent", rt.UserAgent)
	return rt.parent().RoundTrip(req)
}

func (rt *RoundTripper) parent() http.RoundTripper {
	if rt.Parent != nil {
		return rt.Parent
	}
	return http.DefaultTransport
}
