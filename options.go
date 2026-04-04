package gcode

import "github.com/lestrrat-go/option/v2"

// ParseOption is an option that can be passed to parsing functions.
type ParseOption interface {
	option.Interface
	parseOption()
}

// FormatOption is an option that can be passed to generation functions.
type FormatOption interface {
	option.Interface
	formatOption()
}

// Option satisfies both ParseOption and FormatOption.
type Option interface {
	ParseOption
	FormatOption
}

type parseOption struct {
	option.Interface
}

func (*parseOption) parseOption() {}

type formatOption struct {
	option.Interface
}

func (*formatOption) formatOption() {}

type sharedOption struct {
	option.Interface
}

func (*sharedOption) parseOption()    {}
func (*sharedOption) formatOption() {}

// Ident types for option identification.
type identDialect struct{}
type identStrict struct{}
type identEmitComments struct{}
type identEmitLineNumbers struct{}
type identComputeChecksum struct{}
type identLineEnding struct{}

// WithDialect returns an option that associates a dialect with a parse or
// generate operation.
func WithDialect(d *Dialect) Option {
	return &sharedOption{option.New(identDialect{}, d)}
}

// LineEnding represents the line ending style used during G-code generation.
type LineEnding int

const (
	// LineEndingLF uses a single line feed character.
	LineEndingLF LineEnding = iota
	// LineEndingCRLF uses a carriage return followed by a line feed.
	LineEndingCRLF
)
