package events

import (
	"github.com/sirupsen/logrus"
)

const (
	LoggingEventSinkType = "stdout"
)

type LoggingEventSink struct {
	logger *logrus.Logger
}

// Creates an event sink which logs events using a provided logrus logger,
// fields "component" and "event_type" are applied to the logger and
// warnings are logged for each event.
func NewLoggingEventSink(logger *logrus.Logger) *LoggingEventSink {
	return &LoggingEventSink{
		logger: logger,
	}
}

func (sink LoggingEventSink) AddEvent(e Event) error {
	sink.logger.WithFields(logrus.Fields{
		"event_type": e.GetType(),
		"component":  e.GetComponent(),
	}).Warn(e.GetMessage())
	return nil
}
