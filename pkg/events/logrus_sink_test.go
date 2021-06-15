package events

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestLogrusSink(t *testing.T) {
	t.Parallel()

	log, hook := test.NewNullLogger()

	sink := NewLoggingEventSink(log)

	event := LossyFeedEvent{
		Feed: "Foo",
	}

	err := sink.AddEvent(event)
	if err != nil {
		t.Error(err)
	}

	logEntry := hook.LastEntry()
	if logEntry == nil {
		t.Fatal("Log entry was not added to the configured logger")
	}

	if logEntry.Data["event_type"] != event.GetType() {
		t.Errorf(
			"Log entry had incorrect event_type field '%v' when '%v' was expected",
			logEntry.Data["event_type"],
			event.GetType(),
		)
	}

	if logEntry.Data["component"] != event.GetComponent() {
		t.Errorf(
			"Log entry had incorrect component field '%v' when '%v' was expected",
			logEntry.Data["component"],
			event.GetComponent(),
		)
	}
}
