// Enhanced error types for better error handling
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// EnhancedErrType represents the category of enhanced errors (internal)
type EnhancedErrType int

const (
	EnhancedErrTypeUnknown EnhancedErrType = iota
	EnhancedErrTypeConfig
	EnhancedErrTypeValidation
	EnhancedErrTypeNetwork
	EnhancedErrTypePermission
	EnhancedErrTypeNotFound
	EnhancedErrTypeTimeout
	EnhancedErrTypePanic
)

// EnhancedError provides structured error information
type EnhancedError struct {
	Type       EnhancedErrType
	Message    string
	Context    map[string]interface{}
	StackTrace []string
	Wrapped    error
	Component  string
	Operation  string
}

// NewEnhancedError creates a new enhanced error
func NewEnhancedError(errType EnhancedErrType, message string) *EnhancedError {
	return &EnhancedError{
		Type:       errType,
		Message:    message,
		Context:    make(map[string]interface{}),
		StackTrace: captureStackTrace(),
	}
}

// WithContext adds context to the error
func (e *EnhancedError) WithContext(key string, value interface{}) *EnhancedError {
	e.Context[key] = value
	return e
}

// WithComponent sets the component that generated the error
func (e *EnhancedError) WithComponent(component string) *EnhancedError {
	e.Component = component
	return e
}

// WithOperation sets the operation that failed
func (e *EnhancedError) WithOperation(operation string) *EnhancedError {
	e.Operation = operation
	return e
}

// Wrap wraps an existing error
func (e *EnhancedError) Wrap(err error) *EnhancedError {
	e.Wrapped = err
	return e
}

// Error implements the error interface
func (e *EnhancedError) Error() string {
	var parts []string

	if e.Component != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Component))
	}

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Operation))
	}

	parts = append(parts, e.Message)

	if e.Wrapped != nil {
		parts = append(parts, fmt.Sprintf(": %v", e.Wrapped))
	}

	return strings.Join(parts, " ")
}

// Unwrap returns the wrapped error
func (e *EnhancedError) Unwrap() error {
	return e.Wrapped
}

// captureStackTrace captures the current stack trace
func captureStackTrace() []string {
	var stack []string
	for i := 1; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stack
}

// IsEnhancedErrorType checks if an error is of a specific enhanced type
func IsEnhancedErrorType(err error, errType EnhancedErrType) bool {
	if enhanced, ok := err.(*EnhancedError); ok {
		return enhanced.Type == errType
	}
	return false
}

// GetEnhancedErrorType returns the enhanced error type if it's an enhanced error
func GetEnhancedErrorType(err error) EnhancedErrType {
	if enhanced, ok := err.(*EnhancedError); ok {
		return enhanced.Type
	}
	return EnhancedErrTypeUnknown
}
