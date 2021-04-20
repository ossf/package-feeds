package events

import (
	"testing"
)

func TestHandlerDispatchEventNoFilterConfigured(t *testing.T) {
	t.Parallel()

	sink := &MockSink{}
	filter := NewFilter(nil, nil, nil)

	handler := NewHandler(sink, *filter)

	event := &LossyFeedEvent{
		Feed: "Foo",
	}

	err := handler.DispatchEvent(event)
	if err != nil {
		t.Fatal(err)
	}

	if len(sink.events) != 0 {
		t.Error("LossyFeedEvent was dispatched despite not being enabled")
	}
}

func TestHandlerDispatchEventFilterAllowLossyFeed(t *testing.T) {
	t.Parallel()

	sink := &MockSink{}
	filter := NewFilter([]string{LossyFeedEventType}, nil, nil)

	handler := NewHandler(sink, *filter)

	event := &LossyFeedEvent{
		Feed: "Foo",
	}

	err := handler.DispatchEvent(event)
	if err != nil {
		t.Fatal(err)
	}

	if len(sink.events) != 1 {
		t.Error("LossyFeedEvent was not dispatched despite being configured to allow dispatch")
	}
}

func TestHandlerDispatchEventFilterAllowFeedComponent(t *testing.T) {
	t.Parallel()

	sink := &MockSink{}
	filter := NewFilter(nil, nil, []string{FeedsComponentType})

	handler := NewHandler(sink, *filter)

	event := &LossyFeedEvent{
		Feed: "Foo",
	}

	err := handler.DispatchEvent(event)
	if err != nil {
		t.Fatal(err)
	}

	if len(sink.events) != 1 {
		t.Error("LossyFeedEvent was not dispatched despite feeds component being allowed")
	}
}

func TestHandlerDispatchEventFilterDisableLossyFeed(t *testing.T) {
	t.Parallel()

	sink := &MockSink{}
	filter := NewFilter(nil, []string{LossyFeedEventType}, []string{FeedsComponentType})

	handler := NewHandler(sink, *filter)

	event := &LossyFeedEvent{
		Feed: "Foo",
	}

	err := handler.DispatchEvent(event)
	if err != nil {
		t.Fatal(err)
	}

	if len(sink.events) != 0 {
		t.Error("LossyFeedEvent was dispatched despite being configured to disable dispatch")
	}
}
