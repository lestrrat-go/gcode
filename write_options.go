package gcode

import "github.com/lestrrat-go/option/v2"

// WithEmitComments controls whether the [Writer] emits comments.
// Default: true.
func WithEmitComments(v bool) WriteOption {
	return &writeOption{option.New[bool](identEmitComments{}, v)}
}

// WithEmitLineNumbers controls whether the [Writer] emits the
// per-line N field. Default: false.
func WithEmitLineNumbers(v bool) WriteOption {
	return &writeOption{option.New[bool](identEmitLineNumbers{}, v)}
}

// WithComputeChecksum controls whether the [Writer] computes and
// appends a checksum to each emitted line. Default: false.
func WithComputeChecksum(v bool) WriteOption {
	return &writeOption{option.New[bool](identComputeChecksum{}, v)}
}

// WithLineEnding selects the line ending style used by the [Writer].
// Default: [LineEndingLF].
func WithLineEnding(le LineEnding) WriteOption {
	return &writeOption{option.New[LineEnding](identLineEnding{}, le)}
}
