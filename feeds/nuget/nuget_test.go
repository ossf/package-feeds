package nuget

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var testEndpoint *url.URL

func TestCanParseFeed(t *testing.T) {
	srv := nugetAPIMock()

	testEndpoint, _ = url.Parse(srv.URL)
	feedURL = makeTestURL("v3/index.json")

	sut := Feed{}
	cutoff := time.Now().Add(-5 * time.Minute)

	results, err := sut.Latest(cutoff)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("1 result expected but %d retrieved", len(results))
	}

	const expectedName = "new.expected.package"
	const expectedVersion = "0.0.1"
	const expectedType = "nuget"
	result := results[0]

	if result.Name != expectedName {
		t.Fatalf("expected %s but %s was retrieved", expectedName, result.Name)
	}

	if result.Version != expectedVersion {
		t.Fatalf("expected version %s but %s was retrieved", expectedVersion, result.Version)
	}

	if result.Type != expectedType {
		t.Fatalf("expected type %s but %s was retrieved", expectedType, result.Type)
	}
}

func nugetAPIMock() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/v3/index.json", indexMock)
	handler.HandleFunc("/v3/catalog0/index.json", catalogMock)
	handler.HandleFunc("/v3/catalog0/page1.json", catalogPageMock)
	handler.HandleFunc("/v3/catalog0/data/somecatalog/new.expected.package.0.0.1.json", packageDetailMock)

	srv := httptest.NewServer(handler)
	return srv
}

func indexMock(w http.ResponseWriter, r *http.Request) {
	catalogEndpoint := makeTestURL("v3/catalog0/index.json")
	response := fmt.Sprintf(`{"resources": [{"@id": "%s", "@type": "Catalog/3.0.0"}]}`, catalogEndpoint)

	_, _ = w.Write([]byte(response))
}

func catalogMock(w http.ResponseWriter, r *http.Request) {
	pageEndpoint := makeTestURL("v3/catalog0/page1.json")

	response := fmt.Sprintf(`{"items": [{"@id": "%s", "commitTimeStamp": "%s"}]}`, pageEndpoint, time.Now().Format(time.RFC3339))

	_, _ = w.Write([]byte(response))
}

func catalogPageMock(w http.ResponseWriter, r *http.Request) {
	pkgAdded := "nuget:PackageDetails"
	pkgDeleted := "nuget:PackageDelete"
	pkgTemplate := `{"@id": "%s", "@type": "%s", "commitTimeStamp": "%s"}`

	addedItem := fmt.Sprintf(pkgTemplate, makeTestURL("v3/catalog0/data/somecatalog/new.expected.package.0.0.1.json"), pkgAdded, time.Now().UTC().Format(time.RFC3339))
	oldAddedItem := fmt.Sprintf(pkgTemplate, makeTestURL("v3/catalog0/data/somecatalog/old.not.expected.package.0.0.1.json"), pkgAdded, time.Now().UTC().Add(-10*time.Minute).Format(time.RFC3339))
	deletedItem := fmt.Sprintf(pkgTemplate, makeTestURL("v3/catalog0/data/somecatalog/modified.not.expected.0.0.1.json"), pkgDeleted, time.Now().UTC().Format(time.RFC3339))

	response := fmt.Sprintf(`{"items": [%s, %s, %s]}`, addedItem, deletedItem, oldAddedItem)

	_, _ = w.Write([]byte(response))
}

func packageDetailMock(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf(`{"id": "new.expected.package", "version": "0.0.1", "published": "%s"}`, time.Now().UTC().Add(-1*time.Minute).Format(time.RFC3339))

	_, _ = w.Write([]byte(response))
}

func makeTestURL(suffix string) string {
	path, _ := url.Parse(suffix)
	return testEndpoint.ResolveReference(path).String()
}
