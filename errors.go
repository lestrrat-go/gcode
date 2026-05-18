package gcode

import (
	"errors"
	"fmt"
)

// ErrParse is the sentinel for all parse errors. Use errors.Is(err, ErrParse).
var ErrParse = errors.New("gcode: parse error")

// ParseErrorDetail is the exported interface callers use with errors.As
// to extract structured information from parse errors.
//
// Usage:
//
//	var detail gcode.ParseErrorDetail
//	if errors.As(err, &detail) {
//	    fmt.Printf("line %d, col %d: near %q\n", detail.Line(), detail.Column(), detail.Text())
//	}
type ParseErrorDetail interface {
	error
	// Line returns the 1-based source line number where the error occurred.
	Line() int
	// Column returns the 1-based byte column where the error occurred.
	Column() int
	// Text returns a short excerpt of the offending text.
	Text() string
}

// parseError is the unexported implementation of ParseErrorDetail.
type parseError struct {
	line   int
	column int
	text   string
	err    error
}

// Error returns a formatted error message including location and context.
func (e *parseError) Error() string {
	return fmt.Sprintf("gcode: parse error at line %d col %d: %s (near %q)", e.line, e.column, e.err, e.text)
}

// Unwrap returns the underlying error.
func (e *parseError) Unwrap() error { return e.err }

// Is reports whether target matches ErrParse.
func (e *parseError) Is(target error) bool {
	return target == ErrParse
}

// Line returns the 1-based source line number.
func (e *parseError) Line() int { return e.line }

// Column returns the 1-based byte column.
func (e *parseError) Column() int { return e.column }

// Text returns the offending text excerpt.
func (e *parseError) Text() string { return e.text }

// makeParseError is the private constructor for parse errors.
func makeParseError(line, col int, text string, err error) *parseError {
	return &parseError{
		line:   line,
		column: col,
		text:   text,
		err:    err,
	}
}
