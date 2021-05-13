package crates

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
	testutils "github.com/ossf/package-feeds/utils/test"
)

func TestCratesLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		activityPath: cratesSummaryResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL
	if err != nil {
		t.Fatalf("Failed to create crates feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, err := feed.Latest(cutoff)
	if err != nil {
		t.Fatalf("feed.Latest returned error: %v", err)
	}

	if pkgs[0].Name != "FooPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `FooPackage`", pkgs[0].Name)
	}
	if pkgs[1].Name != "BarPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `BarPackage`", pkgs[1].Name)
	}
	if pkgs[0].Version != "0.2.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.2.0`", pkgs[0].Version)
	}
	if pkgs[1].Version != "0.1.1" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.1.1`", pkgs[1].Version)
	}

	for _, p := range pkgs {
		if p.Type != FeedName {
			t.Errorf("Feed type not set correctly in crates package following Latest()")
		}
	}
}

func cratesSummaryResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
{
	"just_updated": [
		{
			"id": "FooPackage",
			"name": "FooPackage",
			"updated_at": "2021-03-19T13:36:33.871721+00:00",
			"versions": null,
			"keywords": null,
			"categories": null,
			"badges": null,
			"created_at": "2021-03-17T20:04:15.901201+00:00",
			"downloads": 46,
			"recent_downloads": 46,
			"max_version": "0.2.0",
			"newest_version": "0.2.0",
			"max_stable_version": "0.2.0",
			"description": "Package for foo mangement",
			"homepage": "https://github.com/Foo/Foo",
			"documentation": "https://github.com/Foo/Foo",
			"repository": "https://github.com/Foo/Foo",
			"links": {
			"version_downloads": "/api/v1/crates/Foo/downloads",
			"versions": "/api/v1/crates/Foo/versions",
			"owners": "/api/v1/crates/Foo/owners",
			"owner_team": "/api/v1/crates/Foo/owner_team",
			"owner_user": "/api/v1/crates/Foo/owner_user",
			"reverse_dependencies": "/api/v1/crates/Foo/reverse_dependencies"
			},
			"exact_match": false
		},
		{
			"id": "BarPackage",
			"name": "BarPackage",
			"updated_at": "2021-03-19T13:17:25.784319+00:00",
			"versions": null,
			"keywords": null,
			"categories": null,
			"badges": null,
			"created_at": "2021-03-13T14:24:30.835625+00:00",
			"downloads": 31,
			"recent_downloads": 31,
			"max_version": "0.1.1",
			"newest_version": "0.1.1",
			"max_stable_version": "0.1.1",
			"description": "Provides Bar functionality",
			"homepage": "https://github.com/Bar/Bar",
			"documentation": "https://github.com/Bar/Bar",
			"repository": "https://github.com/Bar/Bar",
			"links": {
			"version_downloads": "/api/v1/crates/Bar/downloads",
			"versions": "/api/v1/crates/Bar/versions",
			"owners": "/api/v1/crates/Bar/owners",
			"owner_team": "/api/v1/crates/Bar/owner_team",
			"owner_user": "/api/v1/crates/Bar/owner_user",
			"reverse_dependencies": "/api/v1/crates/Bar/reverse_dependencies"
			},
			"exact_match": false
		}
		]
}
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
