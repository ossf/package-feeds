package goproxy

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
	testutils "github.com/ossf/package-feeds/utils/test"
)

func TestGoProxyLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		indexPath: goproxyPackageResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create goproxy feed: %v", err)
	}
	feed.baseURL = srv.URL

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, err := feed.Latest(cutoff)
	if err != nil {
		t.Fatalf("feed.Latest returned error: %v", err)
	}

	if pkgs[0].Name != "golang.org/x/foo" {
		t.Errorf("Unexpected package `%s` found in place of expected `golang.org/x/foo`", pkgs[0].Name)
	}
	if pkgs[1].Name != "golang.org/x/bar" {
		t.Errorf("Unexpected package `%s` found in place of expected `golang.org/x/bar`", pkgs[1].Name)
	}
	if pkgs[0].Version != "v0.3.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `v0.3.0`", pkgs[0].Version)
	}
	if pkgs[1].Version != "v0.4.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `v0.4.0`", pkgs[1].Version)
	}

	for _, p := range pkgs {
		if p.Type != FeedName {
			t.Errorf("Feed type not set correctly in goproxy package following Latest()")
		}
	}
}

func goproxyPackageResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`{"Path": "golang.org/x/foo","Version": "v0.3.0","Timestamp": "2019-04-10T19:08:52.997264Z"}
{"Path": "golang.org/x/bar","Version": "v0.4.0","Timestamp": "2019-04-10T20:30:02.04035Z"}
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
