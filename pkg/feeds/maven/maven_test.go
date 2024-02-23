package maven

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
	testutils "github.com/ossf/package-feeds/pkg/utils/test"
)

func TestMavenLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		indexPath: mavenPackageResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create Maven feed: %v", err)
	}
	feed.baseURL = srv.URL + "/api/internal/browse/components"

	cutoff := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, gotCutoff, errs := feed.Latest(cutoff)

	if len(errs) != 0 {
		t.Fatalf("feed.Latest returned error: %v", err)
	}

	// Returned cutoff should match the newest package creation time of packages retrieved.
	wantCutoff := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if gotCutoff.UTC().Sub(wantCutoff).Abs() > time.Second {
		t.Errorf("Latest() cutoff %v, want %v", gotCutoff, wantCutoff)
	}
	if pkgs[0].Name != "com.github.example:project" {
		t.Errorf("Unexpected package `%s` found in place of expected `com.github.example:project`", pkgs[0].Name)
	}
	if pkgs[0].Version != "1.0.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `1.0.0`", pkgs[0].Version)
	}

	for _, p := range pkgs {
		if p.Type != FeedName {
			t.Errorf("Feed type not set correctly in goproxy package following Latest()")
		}
	}
}

func TestMavenNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		indexPath: testutils.NotFoundHandlerFunc,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create Maven feed: %v", err)
	}
	feed.baseURL = srv.URL + "/api/internal/browse/components"

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	_, gotCutoff, errs := feed.Latest(cutoff)
	if cutoff != gotCutoff {
		t.Error("feed.Latest() cutoff should be unchanged if an error is returned")
	}
	if len(errs) == 0 {
		t.Fatalf("feed.Latest() was successful when an error was expected")
	}
}

func mavenPackageResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseJSON := `
	{
			"components": [
					{
							"id": "pkg:maven/com.github.example/project",
							"type": "COMPONENT",
							"namespace": "com.github.example",
							"name": "project",
							"version": "1.0.0",
							"publishedEpochMillis": 946684800000,
							"latestVersionInfo": {
									"version": "1.0.0",
									"timestampUnixWithMS": 946684800000
							}
					},
					{
						"id": "pkg:maven/com.github.example/project1",
						"type": "COMPONENT",
						"namespace": "com.github.example",
						"name": "project",
						"version": "1.0.0",
						"publishedEpochMillis": null,
						"latestVersionInfo": {
								"version": "1.0.0",
								"timestampUnixWithMS": 0
						}
				}
			]
	}
	`
	_, err := w.Write([]byte(responseJSON))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
