package scheduler

import (
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
)

func TestBuildSchedules(t *testing.T) {
	t.Parallel()

	scheduledFeeds := map[string]feeds.ScheduledFeed{
		"Foo": mockFeed{
			packages: []*feeds.Package{
				{Name: "Foo"},
			},
			options: feeds.FeedOptions{PollRate: "30s"},
		},
		"Bar": mockFeed{
			packages: []*feeds.Package{
				{Name: "Bar"},
			},
			options: feeds.FeedOptions{PollRate: "30s"},
		},
		"Baz": mockFeed{
			packages: []*feeds.Package{
				{Name: "Baz"},
			},
			options: feeds.FeedOptions{PollRate: "20s"},
		},
		"Qux": mockFeed{
			packages: []*feeds.Package{
				{Name: "Baz"},
			},
		},
	}
	cutoff := time.Minute
	pub := mockPublisher{}
	schedules, err := buildSchedules(scheduledFeeds, pub, cutoff)
	if err != nil {
		t.Fatalf("Failed to build schedules: %v", err)
	}

	defaultFg, ok := schedules[""]
	if !ok {
		t.Fatalf("Schedules did not contain a FeedGroup under the default schedule")
	}
	twentySecFg, ok := schedules["@every 20s"]
	if !ok {
		t.Fatalf("Schedules did not contain a FeedGroup under the 20s schedule")
	}
	thirtySecFg, ok := schedules["@every 30s"]
	if !ok {
		t.Fatalf("Schedules did not contain a FeedGroup under the 30s schedule")
	}

	if len(defaultFg.feeds) != 1 {
		t.Fatalf("Default schedule contained %v feeds when %v was expected.", len(defaultFg.feeds), 1)
	}
	if len(twentySecFg.feeds) != 1 {
		t.Fatalf("20s schedule contained %v feeds when %v was expected.", len(twentySecFg.feeds), 1)
	}
	if len(thirtySecFg.feeds) != 2 {
		t.Fatalf("30s schedule contained %v feeds when %v was expected.", len(thirtySecFg.feeds), 2)
	}
}
