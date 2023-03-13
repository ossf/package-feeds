package events

const (
	// Event Types.
	LossyFeedEventType = "LOSSY_FEED"

	// Components.
	FeedsComponentType = "Feeds"
)

type Sink interface {
	AddEvent(e Event) error
}

type Event interface {
	GetComponent() string
	GetType() string
	GetMessage() string
}

type Filter struct {
	EnabledEventTypes  []string `yaml:"enabled_event_types"`
	DisabledEventTypes []string `yaml:"disabled_event_types"`

	EnabledComponents []string `yaml:"enabled_components"`
}

type Handler struct {
	eventSink   Sink
	eventFilter Filter
}

func NewHandler(sink Sink, filter Filter) *Handler {
	return &Handler{
		eventSink:   sink,
		eventFilter: filter,
	}
}

func NewNullHandler() *Handler {
	return &Handler{}
}

// Creates a filter for use with an event handler, nil can be provided for non values.
func NewFilter(enabledEventTypes, disabledEventTypes, enabledComponents []string) *Filter {
	if enabledEventTypes == nil {
		enabledEventTypes = []string{}
	}
	if disabledEventTypes == nil {
		disabledEventTypes = []string{}
	}
	if enabledComponents == nil {
		enabledComponents = []string{}
	}
	return &Filter{
		EnabledEventTypes:  enabledEventTypes,
		DisabledEventTypes: disabledEventTypes,
		EnabledComponents:  enabledComponents,
	}
}

// Dispatches an event to the configured sink if it passes the configured filter.
func (h *Handler) DispatchEvent(e Event) error {
	if h.eventSink == nil {
		return nil
	}
	filter := h.eventFilter

	if filter.ShouldDispatch(e) {
		return h.eventSink.AddEvent(e)
	}
	return nil
}

func (h *Handler) GetSink() Sink {
	return h.eventSink
}

func (h *Handler) GetFilter() Filter {
	return h.eventFilter
}

// ShouldDispatch checks whether an event should be dispatched under the
// configured filter options.
// Options are applied as follows:
//   - disabled event types are always disabled.
//   - enabled event types are enabled
//   - enabled components are enabled except for disabled event types.
func (f Filter) ShouldDispatch(e Event) bool {
	dispatch := false
	eComponent := e.GetComponent()
	eType := e.GetType()

	// Enable components.
	if stringExistsInSlice(eComponent, f.EnabledComponents) {
		dispatch = true
	}
	// Handle specific event types.
	if stringExistsInSlice(eType, f.EnabledEventTypes) {
		dispatch = true
	} else if stringExistsInSlice(eType, f.DisabledEventTypes) {
		dispatch = false
	}
	return dispatch
}

// Checks for existence of a string within a slice of strings.
func stringExistsInSlice(s string, slice []string) bool {
	for _, sliceStr := range slice {
		if s == sliceStr {
			return true
		}
	}
	return false
}
