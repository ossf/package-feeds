package pypi_critical

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/testutils"
)

func TestPypiCriticalLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HttpHandlerFunc{
		"/rss/project/foopy/releases.xml": foopyReleasesResponse,
		"/rss/project/barpy/releases.xml": barpyReleasesResponse,
	}
	packageList = []string{
		"foopy",
		"barpy",
	}
	srv := testutils.HttpServerMock(handlers)

	packageUrlFormat = srv.URL + "/rss/project/%s/releases.xml"
	feed := Feed{}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, err := feed.Latest(cutoff)
	if err != nil {
		t.Fatalf("failed to call Latest() with err: %v", err)
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
		t.Fatalf("missing foopy 2.1")
	}
	if _, ok := pkgMap["foopy"]["2.0"]; !ok {
		t.Fatalf("missing foopy 2.0")
	}
	if _, ok := pkgMap["barpy"]["1.1"]; !ok {
		t.Fatalf("missing barpy 1.1")
	}
	if _, ok := pkgMap["barpy"]["1.0"]; !ok {
		t.Fatalf("missing barpy 1.0")
	}
}

func foopyReleasesResponse(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`
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
}

func barpyReleasesResponse(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`
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
}
