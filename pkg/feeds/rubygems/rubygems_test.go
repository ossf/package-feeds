package rubygems

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/utils"
	testutils "github.com/ossf/package-feeds/pkg/utils/test"
)

func TestRubyLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/api/v1/activity/latest.json":       rubyGemsPackagesResponse,
		"/api/v1/activity/just_updated.json": rubyGemsPackagesResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL
	if err != nil {
		t.Fatalf("failed to create new ruby feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, gotCutoff, errs := feed.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatalf("feed.Latest returned error: %v", errs[len(errs)-1])
	}

	// Returned cutoff should match the newest package creation time of packages retrieved.
	wantCutoff := time.Date(2021, 3, 19, 13, 0, 43, 0, time.UTC)
	if gotCutoff.Sub(wantCutoff).Abs() > time.Second {
		t.Errorf("Latest() cutoff %v, want %v", gotCutoff, wantCutoff)
	}

	var fooPkg *feeds.Package
	var barPkg *feeds.Package

	// rubygems constructs pkgs from a dict so the order is unpredictable
	for _, pkg := range pkgs {
		switch pkg.Name {
		case "FooPackage":
			fooPkg = pkg
		case "BarPackage":
			barPkg = pkg
		default:
			t.Errorf("Unexpected package `%s` found in packages", pkg.Name)
		}
	}

	if fooPkg.Version != "0.13.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.13.0`", pkgs[0].Version)
	}
	if barPkg.Version != "0.0.3" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.0.3`", pkgs[1].Version)
	}

	for _, p := range pkgs {
		if p.Type != FeedName {
			t.Errorf("Feed type not set correctly in ruby package following Latest()")
		}
	}
}

func TestRubyGemsNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/api/v1/activity/latest.json":       testutils.NotFoundHandlerFunc,
		"/api/v1/activity/just_updated.json": testutils.NotFoundHandlerFunc,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL
	if err != nil {
		t.Fatalf("failed to create new ruby feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, gotCutoff, errs := feed.Latest(cutoff)
	if cutoff != gotCutoff {
		t.Error("feed.Latest() cutoff should be unchanged if an error is returned")
	}
	if len(errs) == 0 {
		t.Fatalf("feed.Latest() was successful when an error was expected")
	}
	if !errors.Is(errs[len(errs)-1], utils.ErrUnsuccessfulRequest) {
		t.Fatalf("feed.Latest() returned an error which did not match the expected error")
	}
}

func TestRubyGemsPartialNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/api/v1/activity/latest.json":       rubyGemsPackagesResponse,
		"/api/v1/activity/just_updated.json": testutils.NotFoundHandlerFunc,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL
	if err != nil {
		t.Fatalf("failed to create new ruby feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, _, errs := feed.Latest(cutoff)
	if len(errs) != 1 {
		t.Fatalf("feed.Latest() returned %v errors when 1 was expected", len(errs))
	}
	if !errors.Is(errs[len(errs)-1], utils.ErrUnsuccessfulRequest) {
		t.Fatalf("feed.Latest() returned an error which did not match the expected error")
	}
	// Although the just_updated (updatedPackages) endpoint failed, the two latest (newPackages)
	// should be processed.
	if len(pkgs) != 2 {
		t.Fatalf("Latest() produced %v packages instead of the expected %v", len(pkgs), 2)
	}
}

func rubyGemsPackagesResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
[
	{
		"name": "FooPackage",
		"downloads": 35,
		"version": "0.13.0",
		"version_created_at": "2021-03-19T13:00:43.260Z",
		"version_downloads": 35,
		"platform": "ruby",
		"authors": "FooMan",
		"info": "A package to support Foo",
		"licenses": [
			"MIT"
		],
		"metadata": {},
		"yanked": false,
		"sha": "8649253fb98b8ed0f733e2fc723b2435ead35cb1a70004ebff821abe7abaf131",
		"project_uri": "https://rubygems.org/gems/FooPackage",
		"gem_uri": "https://rubygems.org/gems/FooPackage-0.13.0.gem",
		"homepage_uri": "http://github.com/FooMan/FooPackage/",
		"wiki_uri": null,
		"documentation_uri": "https://www.rubydoc.info/gems/FooPackage/0.13.0",
		"mailing_list_uri": null,
		"source_code_uri": null,
		"bug_tracker_uri": null,
		"changelog_uri": null,
		"funding_uri": null
	},
	{
		"name": "BarPackage",
		"downloads": 41,
		"version": "0.0.3",
		"version_created_at": "2021-03-19T12:52:15.157Z",
		"version_downloads": 41,
		"platform": "ruby",
		"authors": "BarMan",
		"info": "A package to add Bar support.",
		"licenses": [
			"MIT"
		],
		"metadata": {},
		"yanked": false,
		"sha": "fd38fbd77499eb494fd84e710034314287d6895460253aec4a7d105e3199a0fb",
		"project_uri": "https://rubygems.org/gems/BarPackage",
		"gem_uri": "https://rubygems.org/gems/BarPackage-0.0.3.gem",
		"homepage_uri": "http://github.com/BarMan/BarPackage/",
		"wiki_uri": null,
		"documentation_uri": "https://www.rubydoc.info/gems/BarPackage/0.0.3",
		"mailing_list_uri": null,
		"source_code_uri": null,
		"bug_tracker_uri": null,
		"changelog_uri": null,
		"funding_uri": null
	}
]

`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
