package gcode

import "github.com/lestrrat-go/option/v3"

// WithStrict enables strict parsing mode on the [Reader]. When a
// [Dialect] is also attached, unknown commands cause [Reader.Read]
// to return a parse error instead of being accepted.
func WithStrict() ReadOption {
	return &readOption{option.New[bool](identStrict{}, true)}
}

// WithMaxLineSize sets the maximum byte length of a single source
// line accepted by the [Reader]. Lines longer than this size cause
// [Reader.Read] to return an error. The default is 16 MiB, which
// accommodates very long Klipper extended-command lines.
func WithMaxLineSize(n int) ReadOption {
	return &readOption{option.New[int](identMaxLineSize{}, n)}
}
