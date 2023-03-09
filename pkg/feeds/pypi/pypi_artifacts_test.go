package pypi

import (
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
	testutils "github.com/ossf/package-feeds/pkg/utils/test"
)

func TestPypiArtifactsLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		updatesPath: updatesXMLHandle,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := NewArtifactFeed(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create new pypi feed: %v", err)
	}
	feed.baseURL = srv.URL

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, errs := feed.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatalf("feed.Latest returned error: %v", err)
	}

	if pkgs[0].Name != "FooPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `FooPackage`", pkgs[0].Name)
	}
	if pkgs[1].Name != "BarPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `BarPackage`", pkgs[1].Name)
	}
	if pkgs[0].Version != "0.0.2" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.0.2`", pkgs[0].Version)
	}
	if pkgs[1].Version != "0.7a2" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.7a2`", pkgs[1].Version)
	}

	for _, p := range pkgs {
		if p.Type != ArtifactFeedName {
			t.Errorf("ArtifactFeed type not set correctly in pypi package following Latest()")
		}
	}
}
