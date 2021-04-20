package scheduler

import (
	"errors"
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

var errPackage = errors.New("error fetching packages")

type mockFeed struct {
	packages []*feeds.Package
	err      error
}

func (mf mockFeed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	return mf.packages, mf.err
}

func TestPoll(t *testing.T) {
	t.Parallel()

	registry := map[string]feeds.ScheduledFeed{
		"foo-feed": mockFeed{
			packages: []*feeds.Package{
				{Name: "foo-package-1"},
				{Name: "foo-package-2"},
			},
		},
		"err-feed": mockFeed{
			err: errPackage,
		},
	}

	sched := &Scheduler{
		registry: registry,
	}
	gotPackages, gotErrs := sched.Poll(time.Now())
	if len(gotErrs) != 1 {
		t.Fatalf("incorrect number of errors received. expected %d got %d", 1, len(gotErrs))
	}
	if !errors.Is(gotErrs[0], errPackage) {
		t.Fatalf("incorrect error received. expected %v got %v", errPackage, gotErrs[0])
	}
	feed, ok := registry["foo-feed"].(mockFeed)
	if !ok {
		t.Fatalf("error retrieving feed from registry map")
	}
	expectedPackages := feed.packages
	if len(gotPackages) != len(expectedPackages) {
		t.Fatalf("incorrect number of packages received. expected %d got %d", len(expectedPackages), len(gotPackages))
	}
	for i, pkg := range gotPackages {
		if pkg.Name != expectedPackages[i].Name {
			t.Fatalf("unexpected packages received. expected %#v got %#v", gotPackages, expectedPackages)
		}
	}
}
