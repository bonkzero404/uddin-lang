package stdlib_cdc

import (
	"fmt"
	"time"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

type cdcModule struct{}

func (m *cdcModule) Name() string { return "cdc" }

func (m *cdcModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"emit":           eventEmitFunc,
		"define_pattern": eventDefinePatternFunc,
		"get_window":     eventGetWindowFunc,
		"clear":          eventClearFunc,
		"count":          eventCountFunc,
	}
}

func init() {
	interpreter.RegisterModule(&cdcModule{})
}

// eventEmitFunc emits an event to the CDC event store.
func eventEmitFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 {
		return interpreter.Value(fmt.Errorf("cdc.emit() requires at least 1 argument (eventType, [data])"))
	}
	eventType, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("cdc.emit(): first argument must be an event type string"))
	}

	store := ctx.CDCStore()
	store.Lock()
	defer store.Unlock()

	event := map[string]any{
		"type":      eventType,
		"timestamp": time.Now().Format(time.RFC3339),
		"id":        fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	if len(args) > 1 {
		event["data"] = args[1]
	}

	store.AppendEvent(event)

	// Keep only the last 1000 events
	events := store.Events()
	if len(events) > 1000 {
		store.SetEvents(events[len(events)-1000:])
	}

	return interpreter.Value(true)
}

// eventDefinePatternFunc defines an event pattern for detection.
func eventDefinePatternFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(fmt.Errorf("cdc.define_pattern() requires 2 arguments (patternName, eventSequence)"))
	}
	patternName, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("cdc.define_pattern(): first argument must be a pattern name string"))
	}

	var sequence []string
	switch v := args[1].(type) {
	case *[]interpreter.Value:
		for _, item := range *v {
			if s, ok := item.(string); ok {
				sequence = append(sequence, s)
			} else {
				return interpreter.Value(fmt.Errorf("cdc.define_pattern(): sequence must contain only strings"))
			}
		}
	default:
		return interpreter.Value(fmt.Errorf("cdc.define_pattern(): second argument must be an array of event types, got %T", v))
	}

	store := ctx.CDCStore()
	store.Lock()
	defer store.Unlock()

	pattern := map[string]any{
		"sequence": sequence,
		"created":  time.Now().Format(time.RFC3339),
	}
	if len(args) > 2 {
		if windowStr, ok := args[2].(string); ok {
			if duration, err := time.ParseDuration(windowStr); err == nil {
				pattern["window"] = duration.String()
			}
		}
	}

	store.SetPattern(patternName, pattern)
	return interpreter.Value(true)
}

// eventGetWindowFunc gets events within a time window.
func eventGetWindowFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 {
		return interpreter.Value(fmt.Errorf("cdc.get_window() requires at least 1 argument (timeWindow, [eventType])"))
	}
	windowStr, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("cdc.get_window(): first argument must be a duration string"))
	}
	duration, err := time.ParseDuration(windowStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("cdc.get_window(): invalid duration format: %v", err))
	}

	store := ctx.CDCStore()
	store.RLock()
	defer store.RUnlock()

	cutoffTime := time.Now().Add(-duration)
	var filteredEvents []map[string]any

	for _, event := range store.Events() {
		if timestampStr, ok := event["timestamp"].(string); ok {
			if eventTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				if eventTime.After(cutoffTime) {
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

	result := make([]interpreter.Value, len(filteredEvents))
	for i, event := range filteredEvents {
		eventMap := make(map[string]interpreter.Value)
		for k, v := range event {
			eventMap[k] = interpreter.Value(v)
		}
		result[i] = interpreter.Value(&eventMap)
	}
	return interpreter.Value(&result)
}

// eventClearFunc clears events from the CDC store.
func eventClearFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	store := ctx.CDCStore()
	store.Lock()
	defer store.Unlock()

	if len(args) == 0 {
		count := len(store.Events())
		store.SetEvents(make([]map[string]any, 0))
		return interpreter.Value(count)
	}

	eventType, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "cdc.clear(): first argument must be an event type string"))
	}

	var filtered []map[string]any
	count := 0
	for _, event := range store.Events() {
		if et, ok := event["type"].(string); ok && et == eventType {
			count++
		} else {
			filtered = append(filtered, event)
		}
	}
	store.SetEvents(filtered)
	return interpreter.Value(count)
}

// eventCountFunc counts events in the CDC store.
func eventCountFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	store := ctx.CDCStore()
	store.RLock()
	defer store.RUnlock()

	if len(args) == 0 {
		return interpreter.Value(len(store.Events()))
	}

	eventType, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "cdc.count(): first argument must be an event type string"))
	}

	count := 0
	for _, event := range store.Events() {
		if et, ok := event["type"].(string); ok && et == eventType {
			count++
		}
	}
	return interpreter.Value(count)
}
