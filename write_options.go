package gcode

import "github.com/lestrrat-go/option/v3"

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

// WithArgPrecision sets per-argument-key output precision for numeric
// arguments. Each entry maps an argument key (e.g. "E", "X") to the
// number of decimal places — passed to [strconv.FormatFloat] with the
// 'f' verb — that the [Writer] uses when emitting that argument.
//
// When a key has an entry here, the Writer renders [Argument.Number]
// at the configured precision instead of writing [Argument.Raw]
// verbatim. Keys without an entry preserve Raw verbatim, so existing
// callers that depend on round-trip fidelity are unaffected.
// Non-numeric arguments (where Argument.IsNumeric is false) and bare
// flags are unaffected regardless of the precision map.
//
// Typical 3D-printing slicer convention:
//
//	w := gcode.NewWriter(out, gcode.WithArgPrecision(map[string]int{
//	    "E": 5, "X": 3, "Y": 3, "Z": 3, "F": 0,
//	}))
//
// Pass nil or an empty map to disable; precision values are clamped to
// the 0..32 range that strconv.FormatFloat accepts.
func WithArgPrecision(prec map[string]int) WriteOption {
	cp := make(map[string]int, len(prec))
	for k, v := range prec {
		if v < 0 {
			v = 0
		}
		if v > 32 {
			v = 32
		}
		cp[k] = v
	}
	return &writeOption{option.New[map[string]int](identArgPrecision{}, cp)}
}
