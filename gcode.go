package gcode

import (
	"bytes"
	"io"
	"strings"
)

// Generate writes the G-code representation of prog to w using the given options.
func Generate(w io.Writer, prog *Program, options ...GenerateOption) error {
	g := NewGenerator(options...)
	return g.Generate(w, prog)
}

// GenerateString returns the G-code representation of prog as a string.
func GenerateString(prog *Program, options ...GenerateOption) (string, error) {
	var sb strings.Builder
	if err := Generate(&sb, prog, options...); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// GenerateBytes returns the G-code representation of prog as a byte slice.
func GenerateBytes(prog *Program, options ...GenerateOption) ([]byte, error) {
	var buf bytes.Buffer
	if err := Generate(&buf, prog, options...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
