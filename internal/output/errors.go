package output

import "fmt"

// Exit codes for the CLI.
const (
	ExitOK         = 0
	ExitGeneral    = 1
	ExitValidation = 2
	ExitAuth       = 3
	ExitNetwork    = 4
)

// ExitError is a structured error that carries an exit code and optional hint.
type ExitError struct {
	Code int
	Err  error
	Hint string // actionable remediation hint
}

func (e *ExitError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\n  Hint: %s", e.Err.Error(), e.Hint)
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error { return e.Err }

// Errorf creates a general ExitError (exit code 1).
func Errorf(format string, a ...interface{}) *ExitError {
	return &ExitError{Code: ExitGeneral, Err: fmt.Errorf(format, a...)}
}

// ErrValidation creates a validation ExitError (exit code 2).
func ErrValidation(format string, a ...interface{}) *ExitError {
	return &ExitError{Code: ExitValidation, Err: fmt.Errorf(format, a...)}
}

// ErrAuth creates an authentication ExitError (exit code 3).
func ErrAuth(format string, a ...interface{}) *ExitError {
	return &ExitError{Code: ExitAuth, Err: fmt.Errorf(format, a...)}
}

// ErrNetwork creates a network ExitError (exit code 4).
func ErrNetwork(format string, a ...interface{}) *ExitError {
	return &ExitError{Code: ExitNetwork, Err: fmt.Errorf(format, a...)}
}

// ErrWithHint creates an ExitError with an actionable hint.
func ErrWithHint(code int, msg, hint string) *ExitError {
	return &ExitError{Code: code, Err: fmt.Errorf("%s", msg), Hint: hint}
}

// ExitCodeFrom extracts the exit code from an error.
// Returns ExitGeneral (1) for non-ExitError errors, ExitOK (0) for nil.
func ExitCodeFrom(err error) int {
	if err == nil {
		return ExitOK
	}
	if e, ok := err.(*ExitError); ok {
		return e.Code
	}
	return ExitGeneral
}
