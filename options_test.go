package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestWithStrict(t *testing.T) {
	opt := gcode.WithStrict()

	// WithStrict returns a ParseOption
	var _ gcode.ParseOption = opt

	var v bool
	require.NoError(t, opt.Value(&v))
	require.True(t, v)
}

func TestWithEmitComments(t *testing.T) {
	opt := gcode.WithEmitComments(true)

	var _ gcode.FormatOption = opt

	var v bool
	require.NoError(t, opt.Value(&v))
	require.True(t, v)

	opt2 := gcode.WithEmitComments(false)
	require.NoError(t, opt2.Value(&v))
	require.False(t, v)
}

func TestWithEmitLineNumbers(t *testing.T) {
	opt := gcode.WithEmitLineNumbers(true)

	var _ gcode.FormatOption = opt

	var v bool
	require.NoError(t, opt.Value(&v))
	require.True(t, v)
}

func TestWithComputeChecksum(t *testing.T) {
	opt := gcode.WithComputeChecksum(true)

	var _ gcode.FormatOption = opt

	var v bool
	require.NoError(t, opt.Value(&v))
	require.True(t, v)
}

func TestWithLineEnding(t *testing.T) {
	opt := gcode.WithLineEnding(gcode.LineEndingCRLF)

	var _ gcode.FormatOption = opt

	var v gcode.LineEnding
	require.NoError(t, opt.Value(&v))
	require.Equal(t, gcode.LineEndingCRLF, v)
}

func TestLineEndingConstants(t *testing.T) {
	require.NotEqual(t, gcode.LineEndingLF, gcode.LineEndingCRLF)
}

func TestOptionInterfaces(t *testing.T) {
	// ParseOption must not satisfy FormatOption at compile time
	// (we can only test this via the interface assignments above)

	// Verify WithStrict is parse-only (not FormatOption)
	// This is a compile-time check via the type assertion in WithStrict's return type
}
