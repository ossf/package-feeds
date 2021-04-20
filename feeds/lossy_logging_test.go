package feeds

import (
	"testing"
	"time"

	"github.com/ossf/package-feeds/events"
)

func TestProcessPackagesNoOverlap(t *testing.T) {
	t.Parallel()
	feedName := "foo-feed"

	mockSink := &events.MockSink{}
	allowLossyFeedEventsFilter := events.NewFilter([]string{events.LossyFeedEventType}, nil, nil)
	eventHandler := events.NewHandler(mockSink, *allowLossyFeedEventsFilter)
	lossyFeedAlerter := NewLossyFeedAlerter(eventHandler)

	baseTime := time.Date(2021, 4, 20, 14, 30, 0, 0, time.UTC)
	pkgs1 := []*Package{
		NewPackage(baseTime.Add(-time.Minute*2), "foopkg", "1.0", feedName),
		NewPackage(baseTime.Add(-time.Minute*3), "barpkg", "1.0", feedName),
	}
	// Populate previous packages
	lossyFeedAlerter.ProcessPackages(feedName, pkgs1)

	pkgs2 := []*Package{
		NewPackage(baseTime, "bazpkg", "1.0", feedName),
		NewPackage(baseTime.Add(-time.Minute*1), "quxpkg", "2.0", feedName),
	}
	// Trigger no overlap
	lossyFeedAlerter.ProcessPackages(feedName, pkgs2)

	evs := mockSink.GetEvents()

	if len(evs) != 1 {
		t.Fatalf("ProcessPackages failed to detect a lack of overlap")
	}

	if evs[0].GetType() != events.LossyFeedEventType {
		t.Errorf("ProcessPackages did not produce a lossy feed event")
	}
}

func TestProcessPackagesWithOverlap(t *testing.T) {
	t.Parallel()
	feedName := "foo-feed"

	mockSink := &events.MockSink{}
	allowLossyFeedEventsFilter := events.NewFilter([]string{events.LossyFeedEventType}, nil, nil)
	eventHandler := events.NewHandler(mockSink, *allowLossyFeedEventsFilter)
	lossyFeedAlerter := NewLossyFeedAlerter(eventHandler)

	baseTime := time.Date(2021, 4, 20, 14, 30, 0, 0, time.UTC)
	pkgs1 := []*Package{
		NewPackage(baseTime.Add(-time.Minute*2), "foopkg", "1.0", feedName),
		NewPackage(baseTime.Add(-time.Minute*3), "barpkg", "1.0", feedName),
	}
	// Populate previous packages
	lossyFeedAlerter.ProcessPackages(feedName, pkgs1)

	pkgs2 := []*Package{
		NewPackage(baseTime, "bazpkg", "1.0", feedName),
		NewPackage(baseTime.Add(-time.Minute*1), "quxpkg", "2.0", feedName),
		NewPackage(baseTime.Add(-time.Minute*2), "foopkg", "1.0", feedName),
	}
	// Trigger overlap
	lossyFeedAlerter.ProcessPackages(feedName, pkgs2)

	evs := mockSink.GetEvents()

	if len(evs) != 0 {
		t.Fatalf("ProcessPackages failed to identify an overlap when one existed")
	}
}
