package interpreter

import "fmt"

// Set implementation using map[string]Value for uniqueness
type Set struct {
	data map[string]Value
	keys []Value // Keep track of insertion order
}

// Stack implementation using slice
type Stack struct {
	data []Value
}

// Queue implementation using slice
type Queue struct {
	data []Value
}

// Helper function to convert Value to string key for Set
func valueToKey(v Value) string {
	switch val := v.(type) {
	case string:
		return "s:" + val
	case int:
		return fmt.Sprintf("i:%d", val)
	case float64:
		return fmt.Sprintf("f:%g", val)
	case bool:
		return fmt.Sprintf("b:%t", val)
	case nil:
		return "n:null"
	default:
		return fmt.Sprintf("o:%p", val)
	}
}

// setNewFunc implements the set_new() built-in function
// set_new() -> set
// Example: s = set_new()
func setNewFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_new", args, 0)
	return Value(&Set{
		data: make(map[string]Value),
		keys: make([]Value, 0),
	})
}

// setAddFunc implements the set_add() built-in function
// set_add(set, element) -> null
// Example: set_add(s, "hello")
func setAddFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "set_add", args, 2); err != nil {
		return Value(err)
	}
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_add() requires first argument to be a set"))
	}

	key := valueToKey(args[1])
	if _, exists := set.data[key]; !exists {
		set.data[key] = args[1]
		set.keys = append(set.keys, args[1])
	}
	return Value(nil)
}

// setRemoveFunc implements the set_remove() built-in function
// set_remove(set, element) -> bool
// Example: set_remove(s, "hello") -> true if removed, false if not found
func setRemoveFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "set_remove", args, 2); err != nil {
		return Value(err)
	}
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_remove() requires first argument to be a set"))
	}

	key := valueToKey(args[1])
	if _, exists := set.data[key]; exists {
		delete(set.data, key)
		// Remove from keys slice
		for i, v := range set.keys {
			if evalEqual(pos, v, args[1]).(bool) {
				set.keys = append(set.keys[:i], set.keys[i+1:]...)
				break
			}
		}
		return Value(true)
	}
	return Value(false)
}

// setHasFunc implements the set_has() built-in function
// set_has(set, element) -> bool
// Example: set_has(s, "hello") -> true if exists
func setHasFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "set_has", args, 2); err != nil {
		return Value(err)
	}
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_has() requires first argument to be a set"))
	}

	key := valueToKey(args[1])
	_, exists := set.data[key]
	return Value(exists)
}

// setSizeFunc implements the set_size() built-in function
// set_size(set) -> int
// Example: set_size(s) -> 3
func setSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "set_size", args, 1); err != nil {
		return Value(err)
	}
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_size() requires a set argument"))
	}

	return Value(len(set.data))
}

// setToArrayFunc implements the set_to_array() built-in function
// set_to_array(set) -> array
// Example: set_to_array(s) -> ["hello", "world"]
func setToArrayFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "set_to_array", args, 1); err != nil {
		return Value(err)
	}
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_to_array() requires a set argument"))
	}

	result := make([]Value, len(set.keys))
	copy(result, set.keys)
	return Value(&result)
}

// stackNewFunc implements the stack_new() built-in function
// stack_new() -> stack
// Example: s = stack_new()
func stackNewFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "stack_new", args, 0); err != nil {
		return Value(err)
	}
	return Value(&Stack{
		data: make([]Value, 0),
	})
}

// stackPushFunc implements the stack_push() built-in function
// stack_push(stack, element) -> null
// Example: stack_push(s, "hello")
func stackPushFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "stack_push", args, 2); err != nil {
		return Value(err)
	}
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_push() requires first argument to be a stack"))
	}

	stack.data = append(stack.data, args[1])
	return Value(nil)
}

// stackPopFunc implements the stack_pop() built-in function
// stack_pop(stack) -> any
// Example: stack_pop(s) -> "hello"
func stackPopFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "stack_pop", args, 1); err != nil {
		return Value(err)
	}
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_pop() requires a stack argument"))
	}

	if len(stack.data) == 0 {
		return Value(nil)
	}

	last := stack.data[len(stack.data)-1]
	stack.data = stack.data[:len(stack.data)-1]
	return last
}

// stackPeekFunc implements the stack_peek() built-in function
// stack_peek(stack) -> any
// Example: stack_peek(s) -> "hello" (without removing)
func stackPeekFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "stack_peek", args, 1); err != nil {
		return Value(err)
	}
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_peek() requires a stack argument"))
	}

	if len(stack.data) == 0 {
		return Value(nil)
	}

	return stack.data[len(stack.data)-1]
}

// stackSizeFunc implements the stack_size() built-in function
// stack_size(stack) -> int
// Example: stack_size(s) -> 3
func stackSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "stack_size", args, 1); err != nil {
		return Value(err)
	}
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_size() requires a stack argument"))
	}

	return Value(len(stack.data))
}

// stackEmptyFunc implements the stack_empty() built-in function
// stack_empty(stack) -> bool
// Example: stack_empty(s) -> true/false
func stackEmptyFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "stack_empty", args, 1); err != nil {
		return Value(err)
	}
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_empty() requires a stack argument"))
	}

	return Value(len(stack.data) == 0)
}

// queueNewFunc implements the queue_new() built-in function
// queue_new() -> queue
// Example: q = queue_new()
func queueNewFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "queue_new", args, 0); err != nil {
		return Value(err)
	}
	return Value(&Queue{
		data: make([]Value, 0),
	})
}

// queueEnqueueFunc implements the queue_enqueue() built-in function
// queue_enqueue(queue, element) -> null
// Example: queue_enqueue(q, "hello")
func queueEnqueueFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "queue_enqueue", args, 2); err != nil {
		return Value(err)
	}
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_enqueue() requires first argument to be a queue"))
	}

	queue.data = append(queue.data, args[1])
	return Value(nil)
}

// queueDequeueFunc implements the queue_dequeue() built-in function
// queue_dequeue(queue) -> any
// Example: queue_dequeue(q) -> "hello"
func queueDequeueFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "queue_dequeue", args, 1); err != nil {
		return Value(err)
	}
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_dequeue() requires a queue argument"))
	}

	if len(queue.data) == 0 {
		return Value(nil)
	}

	first := queue.data[0]
	queue.data = queue.data[1:]
	return first
}

// queueFrontFunc implements the queue_front() built-in function
// queue_front(queue) -> any
// Example: queue_front(q) -> "hello" (without removing)
func queueFrontFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "queue_front", args, 1); err != nil {
		return Value(err)
	}
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_front() requires a queue argument"))
	}

	if len(queue.data) == 0 {
		return Value(nil)
	}

	return queue.data[0]
}

// queueSizeFunc implements the queue_size() built-in function
// queue_size(queue) -> int
// Example: queue_size(q) -> 3
func queueSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "queue_size", args, 1); err != nil {
		return Value(err)
	}
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_size() requires a queue argument"))
	}

	return Value(len(queue.data))
}

// queueEmptyFunc implements the queue_empty() built-in function
// queue_empty(queue) -> bool
// Example: queue_empty(q) -> true/false
func queueEmptyFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "queue_empty", args, 1); err != nil {
		return Value(err)
	}
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_empty() requires a queue argument"))
	}

	return Value(len(queue.data) == 0)
}
