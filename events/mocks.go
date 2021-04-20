package events

type MockSink struct {
	events []Event
}

func (s *MockSink) GetEvents() []Event {
	return s.events
}

func (s *MockSink) AddEvent(e Event) error {
	s.events = append(s.events, e)
	return nil
}

type MockEvent struct {
	Component string
	Type      string
	Message   string
}

func (e MockEvent) GetComponent() string {
	return e.Component
}

func (e MockEvent) GetType() string {
	return e.Type
}

func (e MockEvent) GetMessage() string {
	return e.Message
}
