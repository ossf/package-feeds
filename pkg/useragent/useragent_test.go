package useragent_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ossf/package-feeds/pkg/useragent"
)

func TestRoundTripper(t *testing.T) {
	t.Parallel()
	want := "test user agent string"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("user-agent")
		if got != want {
			t.Errorf("User Agent = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := http.Client{
		Transport: &useragent.RoundTripper{UserAgent: want},
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Fatalf("Get() = %v; want no error", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get() status = %v; want 200", resp.StatusCode)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

func TestRoundTripper_Parent(t *testing.T) {
	t.Parallel()
	want := "test user agent string"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("user-agent")
		if got != want {
			t.Errorf("User Agent = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	calledParent := false
	c := http.Client{
		Transport: &useragent.RoundTripper{
			UserAgent: want,
			Parent: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				calledParent = true
				return http.DefaultTransport.RoundTrip(r)
			}),
		},
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Fatalf("Get() = %v; want no error", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get() status = %v; want 200", resp.StatusCode)
	}
	if !calledParent {
		t.Errorf("Failed to call Parent RoundTripper")
	}
}
