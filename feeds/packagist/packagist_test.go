package packagist

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	srv := serverMock()
	defer srv.Close()
	updateHost = srv.URL
	versionHost = srv.URL
	feed := Feed{}
	cutoff := time.Unix(1614513658, 0)
	latest, err := feed.Latest(cutoff)
	if err != nil {
		t.Fatalf("got error: %s", err)
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

func serverMock() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/p2/", versionMock)
	handler.HandleFunc("/metadata/changes.json", changesMock)
	srv := httptest.NewServer(handler)

	return srv
}
func changesMock(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("since") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid or missing \u0022since\u0022 query parameter, make sure you store the timestamp at the initial point you started mirroring, then send that to begin receiving changes, e.g. https:\/\/packagist.org\/metadata\/changes.json?since=16145146222715 for example.","timestamp":16145146222715}`))
	}
	_, _ = w.Write([]byte(`{"actions":[{"type":"delete","package":"to-delete/deleted-package","time":1614513806},{"type":"update","package":"ossf/package","time":1614514502},{"type":"update","package":"ossf/package~dev","time":1614514502}],"timestamp":16145145025048}`))
}

func versionMock(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{
		"/p2/ossf/package.json":     `{"packages":{"ossf/package":[{"name":"ossf/package","description":"Lorem Ipsum","keywords":["Lorem Ipsum 1","Lorem Ipsum 2"],"homepage":"","version":"v1.0.0","version_normalized":"1.0.0.0","license":["MIT"],"authors":[{"name":"John Doe","email":"john.doe@local"}],"source":{"type":"git","url":"https://github.com/ossf/package.git","reference":"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f"},"dist":{"type":"zip","url":"https://api.github.com/repos/ossf/package/zipball/c3afaa087afb42bfaf40fc3cda1252cd9a653d7f","reference":"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f","shasum":""},"type":"library","time":"2021-02-28T12:20:03+00:00","autoload":{"psr-4":{"package\\":"src/"}},"require":{"php":"^8.0","guzzlehttp/guzzle":"^7.2"},"require-dev":{"codeception/codeception":"^4.1","codeception/module-phpbrowser":"^1.0.0","codeception/module-asserts":"^1.0.0","hoa/console":"^3.17"},"support":{"issues":"https://github.com/ossf/package/issues","source":"https://github.com/ossf/package/tree/v1.0.0"}}]},"minified":"composer/2.0"}`,
		"/p2/ossf/package~dev.json": `{"packages":{"ossf/package":[{"name":"ossf/package","description":"Lorem Ipsum","keywords":["Lorem Ipsum 1","Lorem Ipsum 2"],"homepage":"","version":"dev-master","version_normalized":"dev-master","license":["MIT"],"authors":[{"name":"John Doe","email":"john.doe@local"}],"source":{"type":"git","url":"https://github.com/ossf/package.git","reference":"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f"},"dist":{"type":"zip","url":"https://api.github.com/repos/ossf/package/zipball/c3afaa087afb42bfaf40fc3cda1252cd9a653d7f","reference":"c3afaa087afb42bfaf40fc3cda1252cd9a653d7f","shasum":""},"type":"library","time":"2021-02-28T12:20:03+00:00","autoload":{"psr-4":{"package\\":"src/"}},"default-branch":true,"require":{"php":"^8.0","guzzlehttp/guzzle":"^7.2"},"require-dev":{"codeception/codeception":"^4.1","codeception/module-phpbrowser":"^1.0.0","codeception/module-asserts":"^1.0.0","hoa/console":"^3.17"},"support":{"issues":"https://github.com/ossf/package/issues","source":"https://github.com/ossf/package/tree/v1.0.0"}}]},"minified":"composer/2.0"}`,
	}
	for u, response := range m {
		if u == r.URL.String() {
			_, _ = w.Write([]byte(response))
			return
		}
	}
	http.NotFound(w, r)
}
