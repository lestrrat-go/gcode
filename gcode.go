package gcode

import (
	"bytes"
	"io"
	"strings"
)

// Parse reads G-code from src and returns an immutable Program.
func Parse(src io.Reader, options ...ParseOption) (*Program, error) {
	p := NewParser(options...)
	return p.Parse(src)
}

// ParseString parses G-code from a string and returns an immutable Program.
func ParseString(src string, options ...ParseOption) (*Program, error) {
	p := NewParser(options...)
	return p.ParseString(src)
}

// ParseBytes parses G-code from a byte slice and returns an immutable Program.
func ParseBytes(src []byte, options ...ParseOption) (*Program, error) {
	p := NewParser(options...)
	return p.ParseBytes(src)
}

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
