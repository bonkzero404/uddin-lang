package interpreter

import (
	"bytes"
	"testing"
)

// makeIsolatedInterp creates a fresh interpreter instance for isolation testing.
func makeIsolatedInterp() *interpreter {
	var buf bytes.Buffer
	config := &Config{
		Stdout:     &buf,
		IsUnitTest: true,
	}
	return newInterpreter(config)
}

// Two engines must not share factDatabase or eventStore.
// Each call to newInterpreter() must produce fully independent state.
func TestEngineIsolation_FactDatabase(t *testing.T) {
	// Engine 1: assert a fact
	interp1 := makeIsolatedInterp()
	factAssertFunc(interp1, noPos, []Value{"customer", "alice", map[string]Value{"status": "premium"}})

	// Engine 2 (new interpreter): query the same fact — must NOT be found
	interp2 := makeIsolatedInterp()
	result := factQueryFunc(interp2, noPos, []Value{"customer", "alice"})

	// With per-interpreter factDatabase, this returns nil (fact not found)
	// With old global var, it returns a non-nil value (bug)
	if result != nil && result != false {
		t.Errorf("isolation broken: engine 2 sees engine 1's facts, got %v", result)
	}
}

func TestEngineIsolation_EventStore(t *testing.T) {
	// Engine 1: emit an event
	interp1 := makeIsolatedInterp()
	eventEmitFunc(interp1, noPos, []Value{"login", map[string]Value{"user": "alice"}})

	// Engine 2 (new interpreter): count events — must be 0
	interp2 := makeIsolatedInterp()
	result := eventCountFunc(interp2, noPos, []Value{})

	if result != 0 {
		t.Errorf("isolation broken: engine 2 sees %v events from engine 1", result)
	}
}
