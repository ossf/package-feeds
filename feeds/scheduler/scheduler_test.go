package scheduler

import (
	"errors"
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

type mockFeed struct {
	packages []*feeds.Package
	err      error
}

func (mf mockFeed) Latest(cutoff time.Time) ([]*feeds.Package, error) {
	return mf.packages, mf.err
}

func TestPoll(t *testing.T) {
	packageErr := errors.New("error fetching packages")
	registry := map[string]feeds.ScheduledFeed{
		"foo-feed": mockFeed{
			packages: []*feeds.Package{
				{Name: "foo-package-1"},
				{Name: "foo-package-2"},
			},
		},
		"err-feed": mockFeed{
			err: packageErr,
		},
	}

	sched := &Scheduler{
		registry: registry,
	}
	gotPackages, gotErrs := sched.Poll(time.Now())
	if len(gotErrs) != 1 {
		t.Fatalf("incorrect number of errors received. expected %d got %d", 1, len(gotErrs))
	}
	if gotErrs[0] != packageErr {
		t.Fatalf("incorrect error received. expected %v got %v", packageErr, gotErrs[0])
	}
	feed := registry["foo-feed"].(mockFeed)
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
