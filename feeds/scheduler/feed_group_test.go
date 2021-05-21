package scheduler

import (
	"errors"
	"testing"
	"time"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/publisher"
)

var (
	errPackage    = errors.New("error fetching packages")
	errPublishing = errors.New("error publishing packages")
)

func TestFeedGroupPoll(t *testing.T) {
	t.Parallel()

	mockFeeds := []feeds.ScheduledFeed{
		mockFeed{
			packages: []*feeds.Package{
				{Name: "Foo"},
				{Name: "Bar"},
			},
		},
		mockFeed{
			packages: []*feeds.Package{
				{Name: "Baz"},
				{Name: "Qux"},
			},
		},
	}
	mockPub := mockPublisher{}
	var pub publisher.Publisher = mockPub

	feedGroup := NewFeedGroup(mockFeeds, pub, time.Minute)
	startLastPollValue := feedGroup.lastPoll

	pkgs, err := feedGroup.poll()
	if err != nil {
		t.Fatalf("Unexpected error arose during polling: %v", err)
	}
	if len(pkgs) != 4 {
		t.Fatalf("poll() returned %v packages when 4 were expected", len(pkgs))
	}
	if startLastPollValue.Equal(feedGroup.lastPoll) {
		t.Fatalf("Feed Group did not update last poll as expected")
	}
}

func TestFeedGroupPollWithErr(t *testing.T) {
	t.Parallel()

	mockFeeds := []feeds.ScheduledFeed{
		mockFeed{
			errs: []error{errPackage},
		},
		mockFeed{
			packages: []*feeds.Package{
				{Name: "Baz"},
				{Name: "Qux"},
			},
		},
	}

	mockPub := mockPublisher{}
	var pub publisher.Publisher = mockPub

	feedGroup := NewFeedGroup(mockFeeds, pub, time.Minute)
	startLastPollValue := feedGroup.lastPoll

	pkgs, err := feedGroup.poll()
	if err == nil {
		t.Fatalf("Expected error during polling")
	}
	if len(pkgs) != 2 {
		t.Fatalf("Expected 2 packages alongside errors but found %v", len(pkgs))
	}
	if startLastPollValue.Equal(feedGroup.lastPoll) {
		t.Fatalf("Feed Group did not update last poll as expected")
	}
}

func TestFeedGroupPublish(t *testing.T) {
	t.Parallel()

	pkgs := []*feeds.Package{
		{Name: "Baz"},
		{Name: "Qux"},
	}
	mockFeeds := []feeds.ScheduledFeed{}

	pubMessages := []string{}
	mockPub := mockPublisher{sendCallback: func(msg string) error {
		pubMessages = append(pubMessages, msg)
		return nil
	}}
	var pub publisher.Publisher = mockPub

	feedGroup := NewFeedGroup(mockFeeds, pub, time.Minute)
	numPublished, err := feedGroup.publishPackages(pkgs)
	if err != nil {
		t.Fatalf("Unexpected error whilst publishing packages: %v", err)
	}
	if numPublished != len(pkgs) {
		t.Fatalf("Expected %v packages to successfully publish but only %v were published", len(pkgs), numPublished)
	}
}

func TestFeedGroupPublishWithErr(t *testing.T) {
	t.Parallel()

	pkgs := []*feeds.Package{
		{Name: "Baz"},
		{Name: "Qux"},
	}
	mockFeeds := []feeds.ScheduledFeed{}

	mockPub := mockPublisher{sendCallback: func(msg string) error {
		return errPublishing
	}}
	var pub publisher.Publisher = mockPub

	feedGroup := NewFeedGroup(mockFeeds, pub, time.Minute)
	_, err := feedGroup.publishPackages(pkgs)
	if err == nil {
		t.Fatalf("publishPackages provided no error when publishing produced an error")
	}
}
