// Package gcode provides a parser and formatter for G-code, the control
// language used by CNC machines and 3D printers.
//
// Use [Parse] or [ParseString] to parse G-code into an immutable [Program],
// and [Format] or [NewFormatter] to serialize it back. Programs are built
// manually with [ProgramBuilder]. A [MacroRegistry] lets you define named
// command sequences that can be expanded on demand.
//
// Dialect-aware parsing is available via the [Dialect] interface and the
// built-in Marlin and RepRap dialects in sub-packages.
package gcode

import (
	"io"
)

// Parse reads G-code from src and returns an immutable Program.
func Parse(src io.Reader, options ...ParseOption) (*Program, error) {
	p := NewParser(options...)
	return p.Parse(src)
}

// ParseString parses G-code from a string and returns an immutable Program.
func ParseString(src string, options ...ParseOption) (*Program, error) {
	p := NewParser(options...)
	return p.ParseString(src)
}

// ParseBytes parses G-code from a byte slice and returns an immutable Program.
func ParseBytes(src []byte, options ...ParseOption) (*Program, error) {
	p := NewParser(options...)
	return p.ParseBytes(src)
}

// Format writes the G-code representation of prog to w using the given options.
func Format(w io.Writer, prog *Program, options ...FormatOption) error {
	f := NewFormatter(options...)
	return f.Format(w, prog)
}


