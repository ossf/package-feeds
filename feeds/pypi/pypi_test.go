package pypi

import (
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/testutils"
)

func TestPypiLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/rss/updates.xml": updatesXMLHandle,
	}
	srv := testutils.HTTPServerMock(handlers)

	baseURL = srv.URL + "/rss/updates.xml"
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
