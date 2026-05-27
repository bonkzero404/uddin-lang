package interpreter

import (
	"fmt"
	"time"
)

// eventEmitFunc emits an event to the event stream
// Usage: event_emit("user_login", {"user": "john", "timestamp": "2023-12-25T10:30:00Z"})
func eventEmitFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		return Value(fmt.Errorf("event_emit requires at least 1 argument (eventType, [data])"))
	}

	eventType, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("event_emit: first argument must be an event type string"))
	}

	interp.eventMutex.Lock()
	defer interp.eventMutex.Unlock()

	// Create event structure
	event := map[string]any{
		"type":      eventType,
		"timestamp": time.Now().Format(time.RFC3339),
		"id":        fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// Add event data if provided
	if len(args) > 1 {
		event["data"] = args[1]
	}

	// Add to event store
	interp.eventStore = append(interp.eventStore, event)

	// Keep only recent events (last 1000)
	if len(interp.eventStore) > 1000 {
		interp.eventStore = interp.eventStore[len(interp.eventStore)-1000:]
	}

	return Value(true)
}

// eventDefinePatternFunc defines an event pattern for detection
// Usage: event_define_pattern("login_sequence", ["user_login", "page_view", "purchase"])
func eventDefinePatternFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(fmt.Errorf("event_define_pattern requires 2 arguments (patternName, eventSequence)"))
	}

	patternName, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("event_define_pattern: first argument must be a pattern name string"))
	}

	// Extract event sequence
	var sequence []string
	switch v := args[1].(type) {
	case *[]Value:
		for _, item := range *v {
			if s, ok := item.(string); ok {
				sequence = append(sequence, s)
			} else {
				return Value(fmt.Errorf("event_define_pattern: sequence must contain only strings"))
			}
		}
	default:
		return Value(fmt.Errorf("event_define_pattern: second argument must be an array of event types"))
	}

	interp.eventMutex.Lock()
	defer interp.eventMutex.Unlock()

	// Store pattern
	pattern := map[string]any{
		"sequence": sequence,
		"created":  time.Now().Format(time.RFC3339),
	}

	// Add optional time window if provided
	if len(args) > 2 {
		if windowStr, ok := args[2].(string); ok {
			if duration, err := time.ParseDuration(windowStr); err == nil {
				pattern["window"] = duration.String()
			}
		}
	}

	interp.eventPatterns[patternName] = pattern
	return Value(true)
}

// eventGetWindowFunc gets events within a time window
// Usage: event_get_window("5m") or event_get_window("1h", "user_login")
func eventGetWindowFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		return Value(fmt.Errorf("event_get_window requires at least 1 argument (timeWindow, [eventType])"))
	}

	windowStr, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("event_get_window: first argument must be a duration string"))
	}

	duration, err := time.ParseDuration(windowStr)
	if err != nil {
		return Value(fmt.Errorf("event_get_window: invalid duration format: %v", err))
	}

	interp.eventMutex.RLock()
	defer interp.eventMutex.RUnlock()

	cutoffTime := time.Now().Add(-duration)
	var filteredEvents []map[string]any

	for _, event := range interp.eventStore {
		// Parse event timestamp
		if timestampStr, ok := event["timestamp"].(string); ok {
			if eventTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				if eventTime.After(cutoffTime) {
					// Filter by event type if specified
					if len(args) > 1 {
						filterType, ok := args[1].(string)
						if !ok {
							continue
						}
						if eventType, ok := event["type"].(string); ok && eventType == filterType {
							filteredEvents = append(filteredEvents, event)
						}
					} else {
						filteredEvents = append(filteredEvents, event)
					}
				}
			}
		}
	}

	// Convert to Value format
	result := make([]Value, len(filteredEvents))
	for i, event := range filteredEvents {
		eventMap := make(map[string]Value)
		for k, v := range event {
			eventMap[k] = Value(v)
		}
		result[i] = Value(&eventMap)
	}

	return Value(&result)
}

// eventClearFunc clears events from the event store
// Usage: event_clear() or event_clear("user_login")
func eventClearFunc(interp *interpreter, pos Position, args []Value) Value {
	interp.eventMutex.Lock()
	defer interp.eventMutex.Unlock()

	if len(args) == 0 {
		// Clear all events
		count := len(interp.eventStore)
		interp.eventStore = make([]map[string]any, 0)
		return Value(count) // int, not int64
	}

	// Clear events of specific type
	eventType, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "event_clear() requires first argument to be an event type string"))
	}

	var filteredEvents []map[string]any
	count := 0

	for _, event := range interp.eventStore {
		if et, ok := event["type"].(string); ok && et == eventType {
			count++
		} else {
			filteredEvents = append(filteredEvents, event)
		}
	}

	interp.eventStore = filteredEvents
	return Value(count) // int, not int64
}

// eventCountFunc counts events in the store
// Usage: event_count() or event_count("user_login")
func eventCountFunc(interp *interpreter, pos Position, args []Value) Value {
	interp.eventMutex.RLock()
	defer interp.eventMutex.RUnlock()

	if len(args) == 0 {
		// Return total count
		return Value(len(interp.eventStore)) // int, not int64
	}

	// Count events of specific type
	eventType, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "event_count() requires first argument to be an event type string"))
	}

	count := 0
	for _, event := range interp.eventStore {
		if et, ok := event["type"].(string); ok && et == eventType {
			count++
		}
	}

	return Value(count) // int, not int64
}
