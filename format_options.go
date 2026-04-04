package gcode

import "github.com/lestrrat-go/option/v2"

// WithEmitComments returns a FormatOption that controls whether
// comments are included in the formatted output.
func WithEmitComments(v bool) FormatOption {
	return &formatOption{option.New[bool](identEmitComments{}, v)}
}

// WithEmitLineNumbers returns a FormatOption that controls whether
// line numbers (N-words) are included in the formatted output.
func WithEmitLineNumbers(v bool) FormatOption {
	return &formatOption{option.New[bool](identEmitLineNumbers{}, v)}
}

// WithComputeChecksum returns a FormatOption that controls whether
// checksums are computed and appended to each line of the formatted output.
func WithComputeChecksum(v bool) FormatOption {
	return &formatOption{option.New[bool](identComputeChecksum{}, v)}
}

// WithLineEnding returns a FormatOption that sets the line ending
// style used in the formatted output.
func WithLineEnding(le LineEnding) FormatOption {
	return &formatOption{option.New[LineEnding](identLineEnding{}, le)}
}
