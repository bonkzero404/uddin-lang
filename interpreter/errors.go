package interpreter

import (
	"fmt"
)

// ErrorInterpreter is the interface for all error types in the interpreter.
// It extends the standard error interface and adds a Position method to get
// the location where the error occurred in the source code.
type ErrorInterpreter interface {
	error
	Position() Position
}

// TypeError is returned for invalid types and wrong number of arguments.
// This error occurs when operations are performed on incompatible types
// or when functions are called with incorrect argument types or counts.
type TypeError struct {
	Message string
	pos     Position
}

// Error returns the formatted error message including position information.
func (e TypeError) Error() string {
	return fmt.Sprintf("type error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

// Position returns the position (line and column) where the error occurred in the source.
func (e TypeError) Position() Position {
	return e.pos
}

// typeError creates a new TypeError with the given position and formatted message.
// This is a helper function used internally to create type errors.
func typeError(pos Position, format string, args ...any) error {
	return TypeError{fmt.Sprintf(format, args...), pos}
}

// ValueError is returned for invalid values (out of bounds index, etc).
// This error occurs when a value is valid in type but invalid in context,
// such as accessing an array with an index that is out of bounds.
type ValueError struct {
	Message string
	pos     Position
}

// Error returns the formatted error message including position information.
func (e ValueError) Error() string {
	return fmt.Sprintf("value error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

// Position returns the position (line and column) where the error occurred in the source.
func (e ValueError) Position() Position {
	return e.pos
}

// valueError creates a new ValueError with the given position and formatted message.
// This is a helper function used internally to create value errors.
func valueError(pos Position, format string, args ...any) error {
	return ValueError{fmt.Sprintf(format, args...), pos}
}

// NameError is returned when a variable or function name is not found in the current scope.
// This error occurs when trying to access a variable that hasn't been defined.
type NameError struct {
	Message string
	pos     Position
}

// Error returns the formatted error message including position information.
func (e NameError) Error() string {
	return fmt.Sprintf("name error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

// Position returns the position (line and column) where the error occurred in the source.
func (e NameError) Position() Position {
	return e.pos
}

// nameError creates a new NameError with the given position and formatted message.
// This is a helper function used internally to create name errors.
func nameError(pos Position, format string, args ...any) error {
	return NameError{fmt.Sprintf(format, args...), pos}
}

// RuntimeError is returned for other or internal runtime errors that don't fit into
// the other error categories. This is a general-purpose error type for the interpreter.
type RuntimeError struct {
	Message string
	pos     Position
}

// Error returns the formatted error message including position information.
func (e RuntimeError) Error() string {
	return fmt.Sprintf("runtime error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

// Position returns the position (line and column) where the error occurred in the source.
func (e RuntimeError) Position() Position {
	return e.pos
}

// runtimeError creates a new RuntimeError with the given position and formatted message.
// This is a helper function used internally to create runtime errors.
func runtimeError(pos Position, format string, args ...any) error {
	return RuntimeError{fmt.Sprintf(format, args...), pos}
}

// BreakException is used to implement break control flow in loops.
// This is not an actual error but uses the exception mechanism to unwind the stack.
type BreakException struct {
	pos Position
}

// Error returns the formatted error message including position information.
func (e BreakException) Error() string {
	return fmt.Sprintf("break at %d:%d", e.pos.Line, e.pos.Column)
}

// Position returns the position (line and column) where the break occurred.
func (e BreakException) Position() Position {
	return e.pos
}

// ContinueException is used to implement continue control flow in loops.
// This is not an actual error but uses the exception mechanism to unwind the stack.
type ContinueException struct {
	pos Position
}

// Error returns the formatted error message including position information.
func (e ContinueException) Error() string {
	return fmt.Sprintf("continue at %d:%d", e.pos.Line, e.pos.Column)
}

// Position returns the position (line and column) where the continue occurred.
func (e ContinueException) Position() Position {
	return e.pos
}
