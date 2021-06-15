package nuget

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
	testutils "github.com/ossf/package-feeds/pkg/utils/test"
)

var testEndpoint *url.URL

func TestCanParseFeed(t *testing.T) {
	t.Parallel()
	var err error
	handlers := map[string]testutils.HTTPHandlerFunc{
		indexPath:                 indexMock,
		"/v3/catalog0/index.json": catalogMock,
		"/v3/catalog0/page1.json": catalogPageMock,
		"/v3/catalog0/data/somecatalog/new.expected.package.0.0.1.json": packageDetailMock,
	}
	srv := testutils.HTTPServerMock(handlers)
	testEndpoint, err = url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("Unexpected error during test url parsing: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}

	sut, err := New(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create nuget feed: %v", err)
	}
	sut.baseURL = srv.URL

	cutoff := time.Now().Add(-5 * time.Minute)

	results, errs := sut.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatal(errs[len(errs)-1])
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

func indexMock(w http.ResponseWriter, r *http.Request) {
	var err error
	catalogEndpoint, err := makeTestURL("v3/catalog0/index.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		response := fmt.Sprintf(`{"resources": [{"@id": "%s", "@type": "Catalog/3.0.0"}]}`,
			catalogEndpoint)
		_, err = w.Write([]byte(response))
		if err != nil {
			http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
		}
	}
}

func catalogMock(w http.ResponseWriter, r *http.Request) {
	var err error
	pageEndpoint, err := makeTestURL("v3/catalog0/page1.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		response := fmt.Sprintf(`{"items": [{"@id": "%s", "commitTimeStamp": "%s"}]}`,
			pageEndpoint, time.Now().Format(time.RFC3339))
		_, err = w.Write([]byte(response))
		if err != nil {
			http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
		}
	}
}

func catalogPageMock(w http.ResponseWriter, r *http.Request) {
	var err error
	pkgAdded := "nuget:PackageDetails"
	pkgDeleted := "nuget:PackageDelete"
	pkgTemplate := `{"@id": "%s", "@type": "%s", "commitTimeStamp": "%s"}`

	addedItemURL, err := makeTestURL("v3/catalog0/data/somecatalog/new.expected.package.0.0.1.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	addedItem := fmt.Sprintf(pkgTemplate,
		addedItemURL,
		pkgAdded, time.Now().UTC().Format(time.RFC3339))

	oldAddedItemURL, err := makeTestURL("v3/catalog0/data/somecatalog/old.not.expected.package.0.0.1.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	oldAddedItem := fmt.Sprintf(pkgTemplate,
		oldAddedItemURL,
		pkgAdded, time.Now().UTC().Add(-10*time.Minute).Format(time.RFC3339))

	deletedItemURL, err := makeTestURL("v3/catalog0/data/somecatalog/modified.not.expected.0.0.1.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	deletedItem := fmt.Sprintf(pkgTemplate,
		deletedItemURL,
		pkgDeleted, time.Now().UTC().Format(time.RFC3339))

	response := fmt.Sprintf(`{"items": [%s, %s, %s]}`, addedItem, deletedItem, oldAddedItem)

	_, err = w.Write([]byte(response))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func packageDetailMock(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf(`{"id": "new.expected.package", "version": "0.0.1", "published": "%s"}`,
		time.Now().UTC().Add(-1*time.Minute).Format(time.RFC3339))

	_, err := w.Write([]byte(response))
	if err != nil {
		http.Error(w, testutils.UnexpectedWriteError(err), http.StatusInternalServerError)
	}
}

func makeTestURL(suffix string) (string, error) {
	path, err := url.Parse(suffix)
	if err != nil {
		return "", err
	}
	return testEndpoint.ResolveReference(path).String(), nil
}
