package pypi

import (
	"reflect"
	"testing"
	"time"

	"github.com/ossf/package-feeds/pkg/feeds"
)

// testAPIData is a sample of PyPI changelog raw data (as a Go struct).
var testAPIData = [][]interface{}{
	{
		"supertemplater",
		"1.4.0",
		int64(1678414652),
		"new release",
	},
	{
		"supertemplater",
		"1.4.0",
		int64(1678414652),
		"add py3 file supertemplater-1.4.0-py3-none-any.whl",
	},
	{
		"supertemplater",
		"1.4.0",
		int64(1678414654),
		"add source file supertemplater-1.4.0.tar.gz",
	},
	{
		"OpenVisus",
		"2.2.96",
		int64(1678414663),
		"add cp310 file OpenVisus-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
	},
	{
		"OpenVisusNoGui",
		"2.2.96",
		int64(1678414694),
		"add cp310 file OpenVisusNoGui-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
	},
	{
		"benchling-api-client",
		"2.0.118",
		int64(1678414736),
		"new release",
	},
	{
		"benchling-api-client",
		"2.0.118",
		int64(1678414736),
		"add py3 file benchling_api_client-2.0.118-py3-none-any.whl",
	},
	{
		"benchling-api-client",
		"2.0.118",
		int64(1678414738),
		"add source file benchling_api_client-2.0.118.tar.gz",
	},
	{
		"adbutils",
		"1.2.9",
		int64(1678415159),
		"new release",
	},
	{
		"adbutils",
		"1.2.9",
		int64(1678415161),
		"add py3 file adbutils-1.2.9-py3-none-manylinux1_x86_64.whl",
	},
	{
		"adbutils",
		"1.2.9",
		int64(1678415164),
		"add py3 file adbutils-1.2.9-py3-none-win32.whl",
	},
	{
		"adbutils",
		"1.2.9",
		int64(1678415166),
		"add py3 file adbutils-1.2.9-py3-none-win_amd64.whl",
	},
	{
		"adbutils",
		"1.2.9",
		int64(1678415167),
		"add source file adbutils-1.2.9.tar.gz",
	},
	{
		"qiskit-qasm2",
		"0.5.1",
		int64(1678415182),
		"new release",
	},
	{
		"qiskit-qasm2",
		"0.5.1",
		int64(1678415182),
		"add source file qiskit_qasm2-0.5.1.tar.gz",
	},
	{
		"dsp-py",
		nil,
		int64(1678415258),
		"remove project",
	},
	{
		"genai",
		"0.12.0a0",
		int64(1678415278),
		"new release",
	},
	{
		"genai",
		"0.12.0a0",
		int64(1678415278),
		"add py3 file genai-0.12.0a0-py3-none-any.whl",
	},
	{
		"genai",
		"0.12.0a0",
		int64(1678415281),
		"add source file genai-0.12.0a0.tar.gz",
	},
	{
		"chia-blockchain",
		"1.7.1rc1",
		int64(1678415319),
		"new release",
	},
	{
		"chia-blockchain",
		"1.7.1rc1",
		int64(1678415319),
		"add source file chia-blockchain-1.7.1rc1.tar.gz",
	},
	{
		"ScraperFC",
		"2.6.3",
		int64(1678415386),
		"new release",
	},
	{
		"ScraperFC",
		"2.6.3",
		int64(1678415386),
		"add py3 file ScraperFC-2.6.3-py3-none-any.whl",
	},
	{
		"ScraperFC",
		"2.6.3",
		int64(1678415389),
		"add source file ScraperFC-2.6.3.tar.gz",
	},
	{
		"callpyfile",
		nil,
		int64(1678415402),
		"create",
	},
	{
		"callpyfile",
		nil,
		int64(1678415402),
		"add Owner hansalemao",
	},
	{
		"callpyfile",
		"0.10",
		int64(1678415402),
		"new release",
	},
	{
		"callpyfile",
		"0.10",
		int64(1678415402),
		"add py3 file callpyfile-0.10-py3-none-any.whl",
	},
	{
		"callpyfile",
		"0.10",
		int64(1678415403),
		"add source file callpyfile-0.10.tar.gz",
	},
}

// testChangelogEntries is a sample of PyPI changelog data after processing, corresponding to the above raw data.
var testChangelogEntries = []pypiChangelogEntry{
	{
		Name:        "supertemplater",
		Version:     "1.4.0",
		Timestamp:   time.Unix(1678414652, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "supertemplater",
		Version:     "1.4.0",
		Timestamp:   time.Unix(1678414652, 0),
		Action:      "add py3 file supertemplater-1.4.0-py3-none-any.whl",
		ArchiveName: "supertemplater-1.4.0-py3-none-any.whl",
	},
	{
		Name:        "supertemplater",
		Version:     "1.4.0",
		Timestamp:   time.Unix(1678414654, 0),
		Action:      "add source file supertemplater-1.4.0.tar.gz",
		ArchiveName: "supertemplater-1.4.0.tar.gz",
	},
	{
		Name:        "OpenVisus",
		Version:     "2.2.96",
		Timestamp:   time.Unix(1678414663, 0),
		Action:      "add cp310 file OpenVisus-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
		ArchiveName: "OpenVisus-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
	},
	{
		Name:        "OpenVisusNoGui",
		Version:     "2.2.96",
		Timestamp:   time.Unix(1678414694, 0),
		Action:      "add cp310 file OpenVisusNoGui-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
		ArchiveName: "OpenVisusNoGui-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
	},
	{
		Name:        "benchling-api-client",
		Version:     "2.0.118",
		Timestamp:   time.Unix(1678414736, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "benchling-api-client",
		Version:     "2.0.118",
		Timestamp:   time.Unix(1678414736, 0),
		Action:      "add py3 file benchling_api_client-2.0.118-py3-none-any.whl",
		ArchiveName: "benchling_api_client-2.0.118-py3-none-any.whl",
	},
	{
		Name:        "benchling-api-client",
		Version:     "2.0.118",
		Timestamp:   time.Unix(1678414738, 0),
		Action:      "add source file benchling_api_client-2.0.118.tar.gz",
		ArchiveName: "benchling_api_client-2.0.118.tar.gz",
	},
	{
		Name:        "adbutils",
		Version:     "1.2.9",
		Timestamp:   time.Unix(1678415159, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "adbutils",
		Version:     "1.2.9",
		Timestamp:   time.Unix(1678415161, 0),
		Action:      "add py3 file adbutils-1.2.9-py3-none-manylinux1_x86_64.whl",
		ArchiveName: "adbutils-1.2.9-py3-none-manylinux1_x86_64.whl",
	},
	{
		Name:        "adbutils",
		Version:     "1.2.9",
		Timestamp:   time.Unix(1678415164, 0),
		Action:      "add py3 file adbutils-1.2.9-py3-none-win32.whl",
		ArchiveName: "adbutils-1.2.9-py3-none-win32.whl",
	},
	{
		Name:        "adbutils",
		Version:     "1.2.9",
		Timestamp:   time.Unix(1678415166, 0),
		Action:      "add py3 file adbutils-1.2.9-py3-none-win_amd64.whl",
		ArchiveName: "adbutils-1.2.9-py3-none-win_amd64.whl",
	},
	{
		Name:        "adbutils",
		Version:     "1.2.9",
		Timestamp:   time.Unix(1678415167, 0),
		Action:      "add source file adbutils-1.2.9.tar.gz",
		ArchiveName: "adbutils-1.2.9.tar.gz",
	},
	{
		Name:        "qiskit-qasm2",
		Version:     "0.5.1",
		Timestamp:   time.Unix(1678415182, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "qiskit-qasm2",
		Version:     "0.5.1",
		Timestamp:   time.Unix(1678415182, 0),
		Action:      "add source file qiskit_qasm2-0.5.1.tar.gz",
		ArchiveName: "qiskit_qasm2-0.5.1.tar.gz",
	},
	{
		Name:        "dsp-py",
		Version:     "",
		Timestamp:   time.Unix(1678415258, 0),
		Action:      "remove project",
		ArchiveName: "",
	},
	{
		Name:        "genai",
		Version:     "0.12.0a0",
		Timestamp:   time.Unix(1678415278, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "genai",
		Version:     "0.12.0a0",
		Timestamp:   time.Unix(1678415278, 0),
		Action:      "add py3 file genai-0.12.0a0-py3-none-any.whl",
		ArchiveName: "genai-0.12.0a0-py3-none-any.whl",
	},
	{
		Name:        "genai",
		Version:     "0.12.0a0",
		Timestamp:   time.Unix(1678415281, 0),
		Action:      "add source file genai-0.12.0a0.tar.gz",
		ArchiveName: "genai-0.12.0a0.tar.gz",
	},
	{
		Name:        "chia-blockchain",
		Version:     "1.7.1rc1",
		Timestamp:   time.Unix(1678415319, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "chia-blockchain",
		Version:     "1.7.1rc1",
		Timestamp:   time.Unix(1678415319, 0),
		Action:      "add source file chia-blockchain-1.7.1rc1.tar.gz",
		ArchiveName: "chia-blockchain-1.7.1rc1.tar.gz",
	},
	{
		Name:        "ScraperFC",
		Version:     "2.6.3",
		Timestamp:   time.Unix(1678415386, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "ScraperFC",
		Version:     "2.6.3",
		Timestamp:   time.Unix(1678415386, 0),
		Action:      "add py3 file ScraperFC-2.6.3-py3-none-any.whl",
		ArchiveName: "ScraperFC-2.6.3-py3-none-any.whl",
	},
	{
		Name:        "ScraperFC",
		Version:     "2.6.3",
		Timestamp:   time.Unix(1678415389, 0),
		Action:      "add source file ScraperFC-2.6.3.tar.gz",
		ArchiveName: "ScraperFC-2.6.3.tar.gz",
	},
	{
		Name:        "callpyfile",
		Version:     "",
		Timestamp:   time.Unix(1678415402, 0),
		Action:      "create",
		ArchiveName: "",
	},
	{
		Name:        "callpyfile",
		Version:     "",
		Timestamp:   time.Unix(1678415402, 0),
		Action:      "add Owner hansalemao",
		ArchiveName: "",
	},
	{
		Name:        "callpyfile",
		Version:     "0.10",
		Timestamp:   time.Unix(1678415402, 0),
		Action:      "new release",
		ArchiveName: "",
	},
	{
		Name:        "callpyfile",
		Version:     "0.10",
		Timestamp:   time.Unix(1678415402, 0),
		Action:      "add py3 file callpyfile-0.10-py3-none-any.whl",
		ArchiveName: "callpyfile-0.10-py3-none-any.whl",
	},
	{
		Name:        "callpyfile",
		Version:     "0.10",
		Timestamp:   time.Unix(1678415403, 0),
		Action:      "add source file callpyfile-0.10.tar.gz",
		ArchiveName: "callpyfile-0.10.tar.gz",
	},
}

// TestPyPIArtifactsLive is a live test of the XML-RPC call with basic sanity checks on the final output data.
func TestPyPIArtifactsLive(t *testing.T) {
	t.Parallel()

	feed, err := NewArtifactFeed(feeds.FeedOptions{})
	if err != nil {
		t.Fatalf("Failed to create new pypi-artifacts feed: %v", err)
	}

	cutoff := time.Now().AddDate(0, 0, -1)
	pkgs, gotCutoff, errs := feed.Latest(cutoff)
	if len(errs) != 0 {
		t.Fatalf("feed.Latest returned error: %v", errs)
	}

	if cutoff == gotCutoff {
		t.Errorf("Latest() cutoff did not change")
	}

	if len(pkgs) == 0 {
		t.Fatalf("Latest() returned no packages")
	}

	for _, p := range pkgs {
		if p.Name == "" {
			t.Errorf("Package has no name: %s", p)
		}
		if p.Version == "" {
			t.Errorf("Package has no version: %s", p)
		}
		if p.ArtifactID == "" {
			t.Errorf("Package has no artifact ID: %s", p)
		}
		if p.CreatedDate.Unix() < cutoff.Unix() {
			t.Errorf("Package create date (%s) is before cutoff (%s)", p.CreatedDate, cutoff)
		}

		if p.Type != ArtifactFeedName {
			t.Errorf("Feed type not set correctly in pypi package following Latest()")
		}
	}
}

func TestProcessRawChangelog(t *testing.T) {
	t.Parallel()

	expectedResults := testChangelogEntries
	actualResults := processRawChangelog(testAPIData)

	if len(actualResults) != len(expectedResults) {
		t.Errorf("expected %d changelog entries, got %d", len(expectedResults), len(actualResults))
	}

	for i, e := range expectedResults {
		if i < len(actualResults) {
			a := actualResults[i]
			if a != e {
				t.Errorf("Mismatch between expectedResult[%d] and actualResult[%d]: want %s, got %s", i, i, e, a)
			}
		} else {
			t.Errorf("expectedResult[%d] with value %s but got no actual result with this index", i, e)
		}
	}
}

func TestCachedLatest(t *testing.T) {
	t.Parallel()

	expectedResults := []*feeds.Package{
		{
			Name:        "supertemplater",
			Version:     "1.4.0",
			CreatedDate: time.Unix(1678414652, 0),
			ArtifactID:  "supertemplater-1.4.0-py3-none-any.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "supertemplater",
			Version:     "1.4.0",
			CreatedDate: time.Unix(1678414654, 0),
			ArtifactID:  "supertemplater-1.4.0.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "OpenVisus",
			Version:     "2.2.96",
			CreatedDate: time.Unix(1678414663, 0),
			ArtifactID:  "OpenVisus-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "OpenVisusNoGui",
			Version:     "2.2.96",
			CreatedDate: time.Unix(1678414694, 0),
			ArtifactID:  "OpenVisusNoGui-2.2.96-cp310-none-macosx_10_9_x86_64.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "benchling-api-client",
			Version:     "2.0.118",
			CreatedDate: time.Unix(1678414736, 0),
			ArtifactID:  "benchling_api_client-2.0.118-py3-none-any.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "benchling-api-client",
			Version:     "2.0.118",
			CreatedDate: time.Unix(1678414738, 0),
			ArtifactID:  "benchling_api_client-2.0.118.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "adbutils",
			Version:     "1.2.9",
			CreatedDate: time.Unix(1678415161, 0),
			ArtifactID:  "adbutils-1.2.9-py3-none-manylinux1_x86_64.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "adbutils",
			Version:     "1.2.9",
			CreatedDate: time.Unix(1678415164, 0),
			ArtifactID:  "adbutils-1.2.9-py3-none-win32.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "adbutils",
			Version:     "1.2.9",
			CreatedDate: time.Unix(1678415166, 0),
			ArtifactID:  "adbutils-1.2.9-py3-none-win_amd64.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "adbutils",
			Version:     "1.2.9",
			CreatedDate: time.Unix(1678415167, 0),
			ArtifactID:  "adbutils-1.2.9.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "qiskit-qasm2",
			Version:     "0.5.1",
			CreatedDate: time.Unix(1678415182, 0),
			ArtifactID:  "qiskit_qasm2-0.5.1.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "genai",
			Version:     "0.12.0a0",
			CreatedDate: time.Unix(1678415278, 0),
			ArtifactID:  "genai-0.12.0a0-py3-none-any.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "genai",
			Version:     "0.12.0a0",
			CreatedDate: time.Unix(1678415281, 0),
			ArtifactID:  "genai-0.12.0a0.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "chia-blockchain",
			Version:     "1.7.1rc1",
			CreatedDate: time.Unix(1678415319, 0),
			ArtifactID:  "chia-blockchain-1.7.1rc1.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "ScraperFC",
			Version:     "2.6.3",
			CreatedDate: time.Unix(1678415386, 0),
			ArtifactID:  "ScraperFC-2.6.3-py3-none-any.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "ScraperFC",
			Version:     "2.6.3",
			CreatedDate: time.Unix(1678415389, 0),
			ArtifactID:  "ScraperFC-2.6.3.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "callpyfile",
			Version:     "0.10",
			CreatedDate: time.Unix(1678415402, 0),
			ArtifactID:  "callpyfile-0.10-py3-none-any.whl",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
		{
			Name:        "callpyfile",
			Version:     "0.10",
			CreatedDate: time.Unix(1678415403, 0),
			ArtifactID:  "callpyfile-0.10.tar.gz",
			Type:        ArtifactFeedName,
			SchemaVer:   "1.1",
		},
	}

	actualResults := getUploadedArtifacts(testChangelogEntries)

	if len(actualResults) != len(expectedResults) {
		t.Errorf("expected %d changelog entries, got %d", len(expectedResults), len(actualResults))
	}

	for i, e := range expectedResults {
		if i < len(actualResults) {
			a := actualResults[i]
			if !reflect.DeepEqual(a, e) {
				t.Errorf("Mismatch between expectedResult[%d] and actualResult[%d]: want %s, got %s", i, i, e, a)
			}
		} else {
			t.Errorf("expectedResult[%d] with value %s but got no actual result with this index", i, e)
		}
	}
}
