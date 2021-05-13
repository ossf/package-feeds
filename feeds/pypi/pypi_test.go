package pypi

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/testutils"
)

func TestPypiLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		updatesPath: updatesXMLHandle,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	if err != nil {
		t.Fatalf("Failed to create new pypi feed: %v", err)
	}
	feed.baseURL = srv.URL

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
	if pkgs[0].Version != "0.0.2" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.0.2`", pkgs[0].Version)
	}
	if pkgs[1].Version != "0.7a2" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.7a2`", pkgs[1].Version)
	}

	for _, p := range pkgs {
		if p.Type != FeedName {
			t.Errorf("Feed type not set correctly in pypi package following Latest()")
		}
	}
}

func TestPypiCriticalLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/rss/project/foopy/releases.xml": foopyReleasesResponse,
		"/rss/project/barpy/releases.xml": barpyReleasesResponse,
	}
	packages := []string{
		"foopy",
		"barpy",
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{
		Packages: &packages,
	}, events.NewNullHandler())
	if err != nil {
		t.Fatalf("Failed to create pypi feed: %v", err)
	}
	feed.baseURL = srv.URL

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, err := feed.Latest(cutoff)
	if err != nil {
		t.Fatalf("Failed to call Latest() with err: %v", err)
	}

	const expectedNumPackages = 4
	if len(pkgs) != expectedNumPackages {
		t.Fatalf("Latest() produced %v packages instead of the expected %v", len(pkgs), expectedNumPackages)
	}
	pkgMap := map[string]map[string]*feeds.Package{}
	pkgMap["foopy"] = map[string]*feeds.Package{}
	pkgMap["barpy"] = map[string]*feeds.Package{}

	for _, pkg := range pkgs {
		pkgMap[pkg.Name][pkg.Version] = pkg
	}

	if _, ok := pkgMap["foopy"]["2.1"]; !ok {
		t.Fatalf("Missing foopy 2.1")
	}
	if _, ok := pkgMap["foopy"]["2.0"]; !ok {
		t.Fatalf("Missing foopy 2.0")
	}
	if _, ok := pkgMap["barpy"]["1.1"]; !ok {
		t.Fatalf("Missing barpy 1.1")
	}
	if _, ok := pkgMap["barpy"]["1.0"]; !ok {
		t.Fatalf("Missing barpy 1.0")
	}
}

// Mock data for pypi firehose with all packages.
func updatesXMLHandle(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
	<title>PyPI recent updates</title>
	<link>https://pypi.org/</link>
	<description>Recent updates to the Python Package Index</description>
	<language>en</language>
	<item>
		<title>FooPackage 0.0.2</title>
		<link>https://pypi.org/project/FooPackage/0.0.2/</link>
		<description>Python wrapper for fooing</description>
		<author>fooman@bazco.org</author>
		<pubDate>Fri, 19 Mar 2021 12:01:04 GMT</pubDate>
	</item>
	<item>
		<title>BarPackage 0.7a2</title>
		<link>https://pypi.org/project/BarPackage/0.7a2/</link>
		<description>A package full of bars</description>
		<author>barman@bazco.org</author>
		<pubDate>Fri, 19 Mar 2021 12:00:39 GMT</pubDate>
	</item>
	</channel>
</rss>
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

// Mock data response for package specific api when pypi is configured with
// a package list in FeedOptions.
func foopyReleasesResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
	<?xml version="1.0" encoding="UTF-8"?>
	<rss version="2.0">
	  <channel>
		<title>PyPI recent updates for foopy</title>
		<link>https://pypi.org/project/foopy/</link>
		<description>Recent updates to the Python Package Index for foopy</description>
		<language>en</language>
		<item>
		  <title>2.1</title>
		  <link>https://pypi.org/project/foopy/2.1/</link>
		  <pubDate>Sat, 27 Mar 2021 22:16:26 GMT</pubDate>
		</item>
		<item>
		  <title>2.0</title>
		  <link>https://pypi.org/project/foopy/2.0/</link>
		  <pubDate>Sun, 23 Sep 2018 16:50:37 GMT</pubDate>
		</item>
	  </channel>
	</rss>
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

// Mock data response for package specific api when pypi is configured with
// a package list in FeedOptions.
func barpyReleasesResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
	<?xml version="1.0" encoding="UTF-8"?>
	<rss version="2.0">
	  <channel>
		<title>PyPI recent updates for barpy</title>
		<link>https://pypi.org/project/barpy/</link>
		<description>Recent updates to the Python Package Index for barpy</description>
		<language>en</language>
		<item>
		  <title>1.1</title>
		  <link>https://pypi.org/project/barpy/1.1/</link>
		  <pubDate>Sat, 27 Mar 2021 22:16:26 GMT</pubDate>
		</item>
		<item>
		  <title>1.0</title>
		  <link>https://pypi.org/project/barpy/1.0/</link>
		  <pubDate>Sun, 23 Sep 2018 16:50:37 GMT</pubDate>
		</item>
	  </channel>
	</rss>
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
