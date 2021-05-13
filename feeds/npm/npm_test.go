package npm

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
	testutils "github.com/ossf/package-feeds/utils/test"
)

func TestNpmLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/-/rss/":     npmLatestPackagesResponse,
		"/FooPackage": fooVersionInfoResponse,
		"/BarPackage": barVersionInfoResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
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
	if pkgs[0].Version != "1.0.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `1.0.0`", pkgs[0].Version)
	}
	if pkgs[1].Version != "0.4.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.4.0`", pkgs[1].Version)
	}

	for _, p := range pkgs {
		if p.Type != FeedName {
			t.Errorf("Feed type not set correctly in npm package following Latest()")
		}
	}
}

func TestNpmNonUtf8Response(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		rssPath: nonUtf8Response,
	}
	srv := testutils.HTTPServerMock(handlers)

	pkgs, err := fetchPackages(srv.URL)
	if err != nil {
		t.Fatalf("Failed to fetch packages: %v", err)
	}

	if len(pkgs) != 1 {
		t.Fatalf("Expected a single package but found %v packages", len(pkgs))
	}

	if pkgs[0].Title != "BarPackage" {
		t.Errorf("Package name '%v' does not match expected '%v'", pkgs[0].Title, "BarPackage")
	}
}

func npmLatestPackagesResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
<?xml version="1.0" encoding="UTF-8"?><rss>
    <channel>
        <title><![CDATA[npm recent updates]]></title>
        <lastBuildDate>Mon, 22 Mar 2021 13:45:33 GMT</lastBuildDate>
        <pubDate>Mon, 22 Mar 2021 13:45:33 GMT</pubDate>
        <item>
            <title><![CDATA[FooPackage]]></title>
            <dc:creator><![CDATA[FooMan]]></dc:creator>
            <pubDate>Mon, 22 Mar 2021 13:45:16 GMT</pubDate>
        </item>
        <item>
            <title><![CDATA[BarPackage]]></title>
            <dc:creator><![CDATA[BarMan]]></dc:creator>
            <pubDate>Mon, 22 Mar 2021 13:07:29 GMT</pubDate>
        </item>
    </channel>
</rss>
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func fooVersionInfoResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
{
	"name": "FooPackage",
	"dist-tags": {
		"latest": "1.0.0"
	}
}
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func barVersionInfoResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
{
	"name": "BarPackage",
	"dist-tags": {
		"latest": "0.4.0"
	}
}
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func nonUtf8Response(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
<?xml version="1.0" encoding="UTF-8"?><rss>
    <channel>
        <title><![CDATA[npm recent updates]]></title>
        <lastBuildDate>Mon, 22 Mar 2021 13:45:33 GMT</lastBuildDate>
        <pubDate>Mon, 22 Mar 2021 13:45:33 GMT</pubDate>
        <item>
            <title><![CDATA[BarPackage���]]></title>
            <dc:creator><![CDATA[Bar���Man]]></dc:creator>
            <pubDate>Mon, 22 Mar 2021 13:07:29 GMT</pubDate>
        </item>
    </channel>
</rss>
`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
