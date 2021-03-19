package npm

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/testutils"
)

func TestNpmLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HttpHandlerFunc{
		"/-/rss/":     npmLatestPackagesResponse,
		"/FooPackage": fooVersionInfoResponse,
		"/BarPackage": barVersionInfoResponse,
	}
	srv := testutils.HttpServerMock(handlers)

	baseURL = srv.URL + "/-/rss/"
	versionURL = srv.URL + "/"
	feed := Feed{}

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
            <pubDate>Mon, 22 Mar 2021 13:07:29 GMT</pubDate>
        </item>
        <item>
            <title><![CDATA[BarPackage]]></title>
            <dc:creator><![CDATA[BarMan]]></dc:creator>
            <pubDate>Mon, 22 Mar 2021 13:45:16 GMT</pubDate>
        </item>
    </channel>
</rss>
`))
	if err != nil {
		fmt.Println("Unexpected error during mock http server write: %w", err)
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
		fmt.Println("Unexpected error during mock http server write: %w", err)
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
		fmt.Println("Unexpected error during mock http server write: %w", err)
	}
}
