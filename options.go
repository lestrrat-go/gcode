package gcode

import "github.com/lestrrat-go/option/v2"

// ParseOption is an option that can be passed to parsing functions.
type ParseOption interface {
	option.Interface
	parseOption()
}

// GenerateOption is an option that can be passed to generation functions.
type GenerateOption interface {
	option.Interface
	generateOption()
}

// Option satisfies both ParseOption and GenerateOption.
type Option interface {
	ParseOption
	GenerateOption
}

type parseOption struct {
	option.Interface
}

func (*parseOption) parseOption() {}

type generateOption struct {
	option.Interface
}

func (*generateOption) generateOption() {}

type sharedOption struct {
	option.Interface
}

func (*sharedOption) parseOption()    {}
func (*sharedOption) generateOption() {}

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
