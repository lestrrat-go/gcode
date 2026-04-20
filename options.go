package gcode

import "github.com/lestrrat-go/option/v3"

// ReadOption configures a [Reader].
type ReadOption interface {
	option.Interface
	readOption()
}

// WriteOption configures a [Writer].
type WriteOption interface {
	option.Interface
	writeOption()
}

// Option satisfies both ReadOption and WriteOption.
type Option interface {
	ReadOption
	WriteOption
}

type readOption struct{ option.Interface }

func (*readOption) readOption() {}

type writeOption struct{ option.Interface }

func (*writeOption) writeOption() {}

type sharedOption struct{ option.Interface }

func (*sharedOption) readOption()  {}
func (*sharedOption) writeOption() {}

type identDialect struct{}
type identStrict struct{}
type identEmitComments struct{}
type identEmitLineNumbers struct{}
type identComputeChecksum struct{}
type identLineEnding struct{}
type identMaxLineSize struct{}

// WithDialect attaches a [Dialect] to a Reader or Writer.
func WithDialect(d *Dialect) Option {
	return &sharedOption{option.New(identDialect{}, d)}
}

// LineEnding selects the line ending style emitted by a [Writer].
type LineEnding int

const (
	// LineEndingLF uses a single line feed character.
	LineEndingLF LineEnding = iota
	// LineEndingCRLF uses a carriage return followed by a line feed.
	LineEndingCRLF
)
