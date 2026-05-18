package gcode

import "github.com/lestrrat-go/option/v3"

// WithStrict enables strict parsing mode on the [Reader] and attaches
// the given [Dialect] to validate against. Unknown commands — those not
// registered with d — cause [Reader.Read] to return a parse error.
//
// Strict mode is meaningless without a dialect, so the dialect is a
// required parameter rather than a separate option. To attach a dialect
// without enabling strict validation (for example, for inspection by
// downstream code), use [WithDialect].
func WithStrict(d *Dialect) ReadOption {
	return &readOption{option.New[*Dialect](identStrict{}, d)}
}

// WithMaxLineSize sets the maximum byte length of a single source
// line accepted by the [Reader]. Lines longer than this size cause
// [Reader.Read] to return an error. The default is 16 MiB, which
// accommodates very long Klipper extended-command lines.
func WithMaxLineSize(n int) ReadOption {
	return &readOption{option.New[int](identMaxLineSize{}, n)}
}
