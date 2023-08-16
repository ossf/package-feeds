package npm

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	testutils "github.com/ossf/package-feeds/pkg/utils/test"
)

func TestNpmLatest(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/-/rss/":      npmLatestPackagesResponse,
		"/FooPackage":  fooVersionInfoResponse,
		"/BarPackage":  barVersionInfoResponse,
		"/BazPackage":  bazVersionInfoResponse,
		"/QuxPackage":  quxVersionInfoResponse,
		"/QuuxPackage": quuxVersionInfoResponse,
	}

	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, _, errs := feed.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatalf("feed.Latest returned error: %v", errs[len(errs)-1])
	}

	if pkgs[0].Name != "FooPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `FooPackage`", pkgs[0].Name)
	}
	if pkgs[1].Name != "BarPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `BarPackage`", pkgs[1].Name)
	}
	if pkgs[2].Name != "BazPackage" || pkgs[3].Name != "BazPackage" {
		t.Errorf("Unexpected packages `%s` & `%s` instead of both being expected as `BazPackage`",
			pkgs[2].Name, pkgs[3].Name)
	}
	if pkgs[4].Name != "QuuxPackage" {
		t.Errorf("Unexpected package `%s` found in place of expected `QuuxPackage`", pkgs[4].Name)
	}
	if pkgs[0].Version != "1.0.1" {
		t.Errorf("Unexpected version `%s` found in place of expected `1.0.1`", pkgs[0].Version)
	}
	if pkgs[1].Version != "0.5.0-alpha" {
		t.Errorf("Unexpected version `%s` found in place of expected `0.5.0-alpha`", pkgs[1].Version)
	}
	if pkgs[2].Version != "1.1" {
		t.Errorf("Unexpected version `%s` found in place of expected `1.1`", pkgs[2].Version)
	}
	if pkgs[3].Version != "1.0" {
		t.Errorf("Unexpected version `%s` found in place of expected `1.0.`", pkgs[3].Version)
	}
	if pkgs[4].Version != "1.1.1" {
		t.Errorf("Unexpected version `%s` found in place of expected `1.1.1.`", pkgs[4].Version)
	}

	fooTime, err := time.Parse(time.RFC3339, "2021-05-11T18:32:01.000Z")
	if err != nil {
		t.Fatalf("time.Parse returned error: %v", err)
	}
	if !pkgs[0].CreatedDate.Equal(fooTime) {
		t.Errorf("Unexpected created date `%s` found in place of expected `2021-05-11T18:32:01.000Z`", pkgs[0].CreatedDate)
	}

	barTime, err := time.Parse(time.RFC3339, "2021-05-11T17:23:02.000Z")
	if err != nil {
		t.Fatalf("time.Parse returned error: %v", err)
	}
	if !pkgs[1].CreatedDate.Equal(barTime) {
		t.Errorf("Unexpected created date `%s` found in place of expected `2021-05-11T17:23:02.000Z`", pkgs[1].CreatedDate)
	}

	bazLatestTime, err := time.Parse(time.RFC3339, "2021-05-11T14:19:45.000Z")
	if err != nil {
		t.Fatalf("time.Parse returned error: %v", err)
	}
	if !pkgs[2].CreatedDate.Equal(bazLatestTime) {
		t.Errorf("Unexpected created date `%s` found in place of expected `2021-05-11T14:19:45.000Z", pkgs[2].CreatedDate)
	}

	bazOldestTime, err := time.Parse(time.RFC3339, "2021-05-11T14:18:32.000Z")
	if err != nil {
		t.Fatalf("time.Parse returned error: %v", err)
	}
	if !pkgs[3].CreatedDate.Equal(bazOldestTime) {
		t.Errorf("Unexpected created date `%s` found in place of expected `2021-05-11T14:18:32.000Z`", pkgs[3].CreatedDate)
	}

	quuxTime, err := time.Parse(time.RFC3339, "2021-05-11T14:15:43.000Z")
	if err != nil {
		t.Fatalf("time.Parse returned error: %v", err)
	}
	if !pkgs[4].CreatedDate.Equal(quuxTime) {
		t.Errorf("Unexpected created date `%s` found in place of expected `2021-05-11T14:15:43.000Z`", pkgs[3].CreatedDate)
	}

	if len(pkgs) != 5 {
		t.Errorf("Unexpected amount of *feed.Package{} generated: %v", len(pkgs))
	}
}

func TestNpmCritical(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/FooPackage": fooVersionInfoResponse,
		"/BarPackage": barVersionInfoResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	packages := []string{
		"FooPackage",
		"BarPackage",
	}

	feed, err := New(feeds.FeedOptions{Packages: &packages}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, gotCutoff, errs := feed.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatalf("Failed to call Latest() with err: %v", errs[len(errs)-1])
	}

	if len(pkgs) != 5 {
		t.Fatalf("Latest() produced %v packages instead of the expected 7", len(pkgs))
	}

	wantCutoff := time.Date(2021, 5, 11, 18, 32, 1, 0, time.UTC)
	if gotCutoff != wantCutoff {
		t.Errorf("Latest() cutoff %v, want %v", gotCutoff, wantCutoff)
	}

	pkgMap := map[string]map[string]*feeds.Package{}
	pkgMap["FooPackage"] = map[string]*feeds.Package{}
	pkgMap["BarPackage"] = map[string]*feeds.Package{}

	for _, pkg := range pkgs {
		pkgMap[pkg.Name][pkg.Version] = pkg
	}

	if _, ok := pkgMap["FooPackage"]["1.0.0"]; !ok {
		t.Fatalf("Missing FooPackage 1.0.0")
	}
	if _, ok := pkgMap["FooPackage"]["0.9.1"]; !ok {
		t.Fatalf("Missing FooPackage 0.9.1")
	}
	if _, ok := pkgMap["FooPackage"]["1.0.1"]; !ok {
		t.Fatalf("Missing FooPackage 1.0.1")
	}
	if _, ok := pkgMap["BarPackage"]["0.4.0"]; !ok {
		t.Fatalf("Missing BarPackage 0.4.0")
	}
	if _, ok := pkgMap["BarPackage"]["0.5.0-alpha"]; !ok {
		t.Fatalf("Missing barpy 0.5.0-alpha")
	}
}

func TestNpmCriticalUnpublished(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/FooPackage": fooVersionInfoResponse,
		"/QuxPackage": quxVersionInfoResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	packages := []string{
		"FooPackage",
		"QuxPackage",
	}

	feed, err := New(feeds.FeedOptions{Packages: &packages}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, _, errs := feed.Latest(cutoff)

	if len(errs) != 1 {
		t.Fatalf("feed.Latest() returned %v errors when 1 was expected", len(errs))
	}

	if !errors.Is(errs[len(errs)-1], errUnpublished) {
		t.Fatalf("Failed to return unpublished error when polling for an unpublished package, instead: %v", err)
	}

	if !strings.Contains(errs[len(errs)-1].Error(), "QuxPackage") {
		t.Fatalf("Failed to correctly include the package name in unpublished error, instead: %v", errs[len(errs)-1])
	}

	// Even though QuxPackage is unpublished, the error should be
	// logged and FooPackage should still be processed.
	if len(pkgs) != 3 {
		t.Fatalf("Latest() produced %v packages instead of the expected 3", len(pkgs))
	}
}

func TestNpmNonUtf8Response(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		rssPath: nonUtf8Response,
	}
	srv := testutils.HTTPServerMock(handlers)

	pkgs, err := fetchPackageEvents(Feed{client: http.DefaultClient, baseURL: srv.URL})
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

func TestNpmNonXMLResponse(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		rssPath: nonXMLResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	pkgs, err := fetchPackageEvents(Feed{client: http.DefaultClient, baseURL: srv.URL})
	if err != nil {
		t.Fatalf("Failed to fetch packages: %v", err)
	}

	if len(pkgs) != 1 {
		t.Fatalf("Expected a single package but found %v packages", len(pkgs))
	}

	expectedTitle := "@abcdefg-software/pog-api"
	if pkgs[0].Title != expectedTitle {
		t.Errorf("Package name '%v' does not match expected '%v'", pkgs[0].Title, expectedTitle)
	}
}

func TestNpmNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/-/rss/": testutils.NotFoundHandlerFunc,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, gotCutoff, errs := feed.Latest(cutoff)
	if cutoff != gotCutoff {
		t.Error("feed.Latest() cutoff should be unchanged if an error is returned")
	}
	if len(errs) != 2 {
		t.Fatalf("feed.Latest() returned %v errors when 2 were expected", len(errs))
	}
	if !errors.Is(errs[len(errs)-1], feeds.ErrNoPackagesPolled) {
		t.Fatalf("feed.Latest() returned an error which did not match the expected error")
	}
}

func TestNpmPartialNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/-/rss/":      npmLatestPackagesResponse,
		"/FooPackage":  fooVersionInfoResponse,
		"/BarPackage":  barVersionInfoResponse,
		"/BazPackage":  bazVersionInfoResponse,
		"/QuxPackage":  testutils.NotFoundHandlerFunc,
		"/QuuxPackage": quuxVersionInfoResponse,
	}
	srv := testutils.HTTPServerMock(handlers)

	feed, err := New(feeds.FeedOptions{}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, _, errs := feed.Latest(cutoff)
	if len(errs) != 1 {
		t.Fatalf("feed.Latest() returned %v errors when 1 was expected", len(errs))
	}
	if !strings.Contains(errs[len(errs)-1].Error(), "QuxPackage") {
		t.Fatalf("Failed to correctly include the package name in feeds.PackagePollError, instead: %v", errs[len(errs)-1])
	}
	if !strings.Contains(errs[len(errs)-1].Error(), "404") {
		t.Fatalf("Failed to wrapped expected 404 error in feeds.PackagePollError, instead: %v", errs[len(errs)-1])
	}
	// Even though QuxPackage returns a 404, the error should be
	// logged and the rest of the packages should still be processed.
	if len(pkgs) != 5 {
		t.Fatalf("Latest() produced %v packages instead of the expected 5", len(pkgs))
	}
}

func TestNpmCriticalPartialNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/FooPackage": fooVersionInfoResponse,
		"/BarPackage": testutils.NotFoundHandlerFunc,
	}
	srv := testutils.HTTPServerMock(handlers)

	packages := []string{
		"FooPackage",
		"BarPackage",
	}

	feed, err := New(feeds.FeedOptions{Packages: &packages}, events.NewNullHandler())
	feed.baseURL = srv.URL

	if err != nil {
		t.Fatalf("Failed to create new npm feed: %v", err)
	}

	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	pkgs, _, errs := feed.Latest(cutoff)
	if len(errs) != 1 {
		t.Fatalf("feed.Latest() returned %v errors when 1 was expected", len(errs))
	}
	if !strings.Contains(errs[len(errs)-1].Error(), "BarPackage") {
		t.Fatalf("Failed to correctly include the package name in feeds.PackagePollError, instead: %v", errs[len(errs)-1])
	}
	if !strings.Contains(errs[len(errs)-1].Error(), "404") {
		t.Fatalf("Failed to wrapped expected 404 error in feeds.PackagePollError, instead: %v", errs[len(errs)-1])
	}
	// Even though BarPackage returns a 404, the error should be
	// logged and FooPackage should still be processed.
	if len(pkgs) != 3 {
		t.Fatalf("Latest() produced %v packages instead of the expected 3", len(pkgs))
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
		<item>
			<title><![CDATA[BazPackage]]></title>
			<dc:creator><![CDATA[BazMan]]></dc:creator>
			<pubDate>Tue, 11 May 2021 14:19:45 GMT</pubDate>
		</item>
		<item>
			<title><![CDATA[BazPackage]]></title>
			<dc:creator><![CDATA[BazMan]]></dc:creator>
			<pubDate>Tue, 11 May 2021 14:18.32 GMT</pubDate>
		</item>
		<item>
			<title><![CDATA[QuxPackage]]></title>
			<dc:creator><![CDATA[QuxMan]]></dc:creator>
			<pubDate>Tue, 11 May 2021 14:17.12 GMT</pubDate>
		</item>
		<item>
			<title><![CDATA[QuuxPackage]]></title>
			<dc:creator><![CDATA[QuuxMan]]></dc:creator>
			<pubDate>Tue, 11 May 2021 14:15.43 GMT</pubDate>
		</item>
		<item>
			<title><![CDATA[QuuxPackage]]></title>
			<dc:creator><![CDATA[QuuxMan]]></dc:creator>
			<pubDate>Tue, 11 May 2021 14:14.31 GMT</pubDate>
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
		"latest": "1.0.1",
		"release-0.9.x": "0.9.1"
	},
	"time": {
		"created" : "2021-03-22T13:07:29.000Z",
		"1.0.0": "2021-03-22T13:07:29.000Z",
		"modified": "2021-05-11T18:34:12.000Z",
		"0.9.1": "2021-03-23T05:17:43.000Z",
		"1.0.1": "2021-05-11T18:32:01.000Z"
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
		"latest": "0.4.0",
		"next": "0.5.0-alpha"
	},
	"time": {
		"created": "2021-03-22T13:45:16.000Z",
		"0.4.0": "2021-03-22T13:45:16.000Z",
		"modified": "2021-05-11T17:24:14.000Z",
		"0.5.0-alpha": "2021-05-11T17:23:02.000Z"
	}
}
`))
	if err != nil {
		fmt.Println("Unexpected error during mock http server write: %w", err)
	}
}

// BazPackage has 2 entries in the registry rss, as such it should result
// in both tags being resolved, in date order.
func bazVersionInfoResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
{
	"name": "BazPackage",
	"dist-tags": {
		"latest": "1.1"
	},
	"time": {
		"created": "2021-05-11T14:18:32.000Z",
		"1.0": "2021-05-11T14:18:32.000Z",
		"modified": "2021-05-11T14:19:46.000Z",
		"1.1": "2021-05-11T14:19:45.000Z"
	}
}
`))
	if err != nil {
		fmt.Println("Unexpected error during mock http server write: %w", err)
	}
}

// QuxPackage has an `unpublished` field, this should't cause an error if polling
// the 'firehose' but a *feeds.Package{} should not be generated. Completely
// unpublishing a package entails there's a minimum of 24hours before a new version
// of it may be published.
func quxVersionInfoResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
{
	"name": "QuxPackage",
	"time": {
		"created": "2021-05-10T14:38:14.000Z",
		"1.0": "2021-05-10T14:38:14.000Z",
		"modified": "2021-05-11T14:17:12.000Z",
		"1.1": "2021-05-11T11:19:43.000Z",
		"unpublished": {
			"name": "Quxman",
			"time": "2021-05-11T14:17:12.000Z",
			"versions": ["1.0", "1.1"]
		}
	}
}
`))
	if err != nil {
		fmt.Println("Unexpected error during mock http server write: %w", err)
	}
}

// QuuxPackage has 2 entries in the registry rss, however one of the
// versions was 'unpublished' (not all, unlike Qux). This could occur
// for instance if a broken version was pushed incorrectly and subsequently
// unpublished, in the delta between the package being reported to the
// rss feed and the scraping of the package specific endpoint. If this
// were to happen for a single registry entry (count == 1), then the
// `cutoff` should handle older than expected versions being mistakenly
// processed, assuming it was not completely unpublished.
func quuxVersionInfoResponse(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
{
	"name": "QuuxPackage",
	"dist-tags": {
		"latest": "1.0.1"
	},
	"time": {
		"created": "2021-05-011T14:14:02.000Z",
		"modified": "2021-05-11T14:15:49.000Z",
		"1.1.1": "2021-05-11T14:15:43.000Z"
	}
}
`))
	if err != nil {
		fmt.Println("Unexpected error during mock http server write: %w", err)
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

func nonXMLResponse(w http.ResponseWriter, r *http.Request) {
	nonXMLChars := "\u0002\u0010\u0014\u0016\u001b\u0000"
	xml := `
<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">
	<channel>
		<item>
			<title><![CDATA[@abcdefg-software/pog-api]]></title>
			<description><![CDATA[API client for @abcdefg-software/pog-api` + nonXMLChars + `]]></description>
		</item>
	</channel>
</rss>
`
	if _, err := w.Write([]byte(xml)); err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}
