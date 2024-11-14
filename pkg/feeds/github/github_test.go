package github

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
	testutils "github.com/ossf/package-feeds/pkg/utils/test"
)

func TestGithubLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/repos/fooOrg/bar/releases": barResponse,
		"/repos/fooOrg/baz/releases": bazResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	urlFormatString = srv.URL + "/repos/%v/releases?per_page=%v"

	packages := []string{
		"fooOrg/bar",
		"fooOrg/baz",
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	feed, err := New(feeds.FeedOptions{Packages: &packages})
	if err != nil {
		t.Fatalf("Unexpected error during feed creation: %v", err)
	}

	pkgs, errs := feed.Latest(cutoff)
	if len(errs) > 0 {
		t.Fatalf("Failed to poll latest packages from feed: %v", errs[0])
	}
	if len(pkgs) != 6 {
		t.Fatalf("Polling feed did not return the expected number of packages")
	}
	expectedVersions := []string{"2.1.0", "1.1.0", "2.0.5", "1.0.5", "1.0.0", "2.0.0"}
	expectedRepos := []string{"fooOrg/baz", "fooOrg/bar", "fooOrg/baz", "fooOrg/bar", "fooOrg/bar", "fooOrg/baz"}

	for i := range pkgs {
		if pkgs[i].Version != expectedVersions[i] {
			t.Errorf("Unexpected version %v found when expecting %v", pkgs[i].Version, expectedVersions[i])
		}
		if pkgs[i].Type != FeedName {
			t.Errorf("Type set incorrectly for feed type")
		}
		if pkgs[i].Name != expectedRepos[i] {
			t.Errorf("Unexpected name %v found when expecting %v", pkgs[i].Name, expectedRepos[i])
		}
	}
}

func TestConfigurationErrors(t *testing.T) {
	t.Parallel()

	_, err := New(feeds.FeedOptions{})
	if !errors.Is(err, errPackageOptionsUnset) {
		t.Fatalf("Expected to fail due to missing packages option")
	}

	packages := &[]string{}
	_, err = New(feeds.FeedOptions{Packages: packages})
	if !errors.Is(err, errMinimumPackagesRequired) {
		t.Fatalf("Expected to fail due to missing packages option")
	}
}

func barResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
[
	{
		"tag_name": "1.1.0",
		"draft": false,
		"prerelease": false,
		"published_at": "2021-02-04T10:22:41Z"
	},
	{
		"tag_name": "1.0.5",
		"draft": false,
		"prerelease": false,
		"published_at": "2021-01-04T10:22:41Z"
	},
	{
		"tag_name": "1.0.0",
		"draft": false,
		"prerelease": false,
		"published_at": "2021-01-02T10:22:41Z"
	}
]
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func bazResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
[
	{
		"tag_name": "2.1.0",
		"draft": false,
		"prerelease": false,
		"published_at": "2021-04-05T10:22:41Z"
	},
	{
		"tag_name": "2.0.5",
		"draft": false,
		"prerelease": false,
		"published_at": "2021-02-01T10:22:41Z"
	},
	{
		"tag_name": "2.0.0",
		"draft": false,
		"prerelease": false,
		"published_at": "2021-01-01T10:22:41Z"
	}
]
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
