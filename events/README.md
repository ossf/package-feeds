# Event Handling

package-feeds supports publishing specific application 'events' to be processed.

## Configuration

```
events:
  sink: "stdout"
  filter:
    enabled_event_types: ["LOSSY_FEED"]
    disabled_event_types: []
    enabled_components: ["Feeds"]
```

## Events

**N.B** Currently only events for potential loss during package polling are available.

Types:
- "LOSSY_FEED" - Potential loss was detected in a feed

Components:
- "Feeds" - Events which occur within feed logic

Sinks:
- "stdout" - Logs events to stdout