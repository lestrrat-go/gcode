package gcode

import "github.com/lestrrat-go/option/v2"

// WithStrict returns a ParseOption that enables strict parsing mode.
func WithStrict() ParseOption {
	return &parseOption{option.New[bool](identStrict{}, true)}
}
