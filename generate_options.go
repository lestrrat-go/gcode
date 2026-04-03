package gcode

import "github.com/lestrrat-go/option/v2"

// WithEmitComments returns a GenerateOption that controls whether
// comments are included in the generated output.
func WithEmitComments(v bool) GenerateOption {
	return &generateOption{option.New[bool](identEmitComments{}, v)}
}

// WithEmitLineNumbers returns a GenerateOption that controls whether
// line numbers (N-words) are included in the generated output.
func WithEmitLineNumbers(v bool) GenerateOption {
	return &generateOption{option.New[bool](identEmitLineNumbers{}, v)}
}

// WithComputeChecksum returns a GenerateOption that controls whether
// checksums are computed and appended to each line of the generated output.
func WithComputeChecksum(v bool) GenerateOption {
	return &generateOption{option.New[bool](identComputeChecksum{}, v)}
}

// WithLineEnding returns a GenerateOption that sets the line ending
// style used in the generated output.
func WithLineEnding(le LineEnding) GenerateOption {
	return &generateOption{option.New[LineEnding](identLineEnding{}, le)}
}
