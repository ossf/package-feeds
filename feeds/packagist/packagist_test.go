package packagist

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/utils"
	testutils "github.com/ossf/package-feeds/utils/test"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/p2/":                   versionMock,
		"/metadata/changes.json": changesMock,
	}
	srv := testutils.HTTPServerMock(handlers)

	defer srv.Close()
	feed, err := New(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create packagist feed: %v", err)
	}
	feed.updateHost = srv.URL
	feed.versionHost = srv.URL

	cutoff := time.Unix(1614513658, 0)
	latest, errs := feed.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatalf("got error: %v", errs[len(errs)-1])
	}
	if len(latest) == 0 {
		t.Fatalf("did not get any updates")
	}
	for _, pkg := range latest {
		if pkg.CreatedDate.Before(cutoff) {
			t.Fatalf("package returned that was updated before cutoff %f", pkg.CreatedDate.Sub(cutoff).Minutes())
		}
		if pkg.Name == "to-delete/deleted-package" {
			t.Fatalf("pkg to-delete/deleted-package was deleted and should not be included here")
		}
	}
}

func TestPackagistNotFound(t *testing.T) {
	t.Parallel()

	handlers := map[string]testutils.HTTPHandlerFunc{
		"/p2/":                   testutils.NotFoundHandlerFunc,
		"/metadata/changes.json": testutils.NotFoundHandlerFunc,
	}
	srv := testutils.HTTPServerMock(handlers)

	defer srv.Close()
	feed, err := New(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create packagist feed: %v", err)
	}
	feed.updateHost = srv.URL
	feed.versionHost = srv.URL

	cutoff := time.Unix(1614513658, 0)
	_, errs := feed.Latest(cutoff)
	if len(errs) != 1 {
		t.Fatalf("feed.Latest() returned %v errors when 1 was expected", len(errs))
	}
	if !errors.Is(errs[len(errs)-1], utils.ErrUnsuccessfulRequest) {
		t.Fatalf("feed.Latest() returned an error which did not match the expected error")
	}
}

func changesMock(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("since") == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error":"Invalid or missing \u0022since\u0022 query parameter,
		make sure you store the timestamp at the initial point you started mirroring,
		then send that to begin receiving changes,
		e.g. https:\/\/packagist.org\/metadata\/changes.json?since=16145146222715 for example."
		,"timestamp":16145146222715}`))
		if err != nil {
			http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
		}
	}
	_, err := w.Write([]byte(`{"actions":[{"type":"delete","package":"to-delete/deleted-package",
	"time":1614513806},{"type":"update","package":"ossf/package","time":1614514502},
	{"type":"update","package":"ossf/package~dev","time":1614514502}],"timestamp":16145145025048}`))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func versionMock(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{
		"/p2/ossf/package.json": `{"packages":{"ossf/package":[{"name":"ossf/package",
		"description":"Lorem Ipsum","keywords":["Lorem Ipsum 1","Lorem Ipsum 2"],"homepage":"",
		"version":"v1.0.0","version_normalized":"1.0.0.0","license":["MIT"],
		"authors":[{"name":"John Doe","email":"john.doe@local"}],
		"source":{"type":"git","url":"https://github.com/ossf/package.git","reference":
		"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f"},"dist":{"type":"zip","url":
		"https://api.github.com/repos/ossf/package/zipball/c3afaa087afb42bfaf40fc3cda1252cd9a653d7f",
		"reference":"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f","shasum":""},"type":"library",
		"time":"2021-02-28T12:20:03+00:00","autoload":{"psr-4":{"package\\":"src/"}},
		"require":{"php":"^8.0","guzzlehttp/guzzle":"^7.2"},
		"require-dev":{"codeception/codeception":"^4.1","codeception/module-phpbrowser":"^1.0.0",
		"codeception/module-asserts":"^1.0.0","hoa/console":"^3.17"},
		"support":{"issues":"https://github.com/ossf/package/issues",
		"source":"https://github.com/ossf/package/tree/v1.0.0"}}]},"minified":"composer/2.0"}`,
		"/p2/ossf/package~dev.json": `{"packages":{"ossf/package":[{"name":"ossf/package",
		"description":"Lorem Ipsum","keywords":["Lorem Ipsum 1","Lorem Ipsum 2"],"homepage":"",
		"version":"dev-master","version_normalized":"dev-master","license":["MIT"],
		"authors":[{"name":"John Doe","email":"john.doe@local"}],
		"source":{"type":"git","url":"https://github.com/ossf/package.git","reference":
		"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f"},"dist":{"type":"zip","url":
		"https://api.github.com/repos/ossf/package/zipball/c3afaa087afb42bfaf40fc3cda1252cd9a653d7f",
		"reference":"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f","shasum":""},"type":"library",
		"time":"2021-02-28T12:20:03+00:00","autoload":{"psr-4":{"package\\":"src/"}},
		"default-branch":true,"require":{"php":"^8.0","guzzlehttp/guzzle":"^7.2"},
		"require-dev":{"codeception/codeception":"^4.1","codeception/module-phpbrowser":"^1.0.0",
		"codeception/module-asserts":"^1.0.0","hoa/console":"^3.17"},
		"support":{"issues":"https://github.com/ossf/package/issues",
		"source":"https://github.com/ossf/package/tree/v1.0.0"}}]},"minified":"composer/2.0"}`,
	}
	for u, response := range m {
		if u == r.URL.String() {
			_, err := w.Write([]byte(response))
			if err != nil {
				http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
			}
			return
		}
	}
	http.NotFound(w, r)
}
