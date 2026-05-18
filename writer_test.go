package gcode_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func writeAll(t *testing.T, lines []gcode.Line, opts ...gcode.WriteOption) string {
	t.Helper()
	var buf bytes.Buffer
	w := gcode.NewWriter(&buf, opts...)
	for _, l := range lines {
		require.NoError(t, w.Write(l))
	}
	require.NoError(t, w.Flush())
	return buf.String()
}

func TestWriterClassic(t *testing.T) {
	t.Parallel()
	out := writeAll(t, []gcode.Line{
		gcode.NewLine("G28"),
		gcode.NewLine("G1").Arg("X", "10").Arg("Y", "20"),
	})
	require.Equal(t, "G28\nG1 X10 Y20\n", out)
}

func TestWriterExtended(t *testing.T) {
	t.Parallel()
	out := writeAll(t, []gcode.Line{
		gcode.NewLine("SET_FAN_SPEED").Arg("FAN", "cooling").Arg("SPEED", "0.5"),
	})
	require.Equal(t, "SET_FAN_SPEED FAN=cooling SPEED=0.5\n", out)
}

func TestWriterExtendedFlag(t *testing.T) {
	t.Parallel()
	out := writeAll(t, []gcode.Line{
		gcode.NewLine("TIMELAPSE_TAKE_FRAME"),
	})
	require.Equal(t, "TIMELAPSE_TAKE_FRAME\n", out)
}

func TestWriterCommentOnly(t *testing.T) {
	t.Parallel()
	out := writeAll(t, []gcode.Line{
		gcode.NewComment(" hello"),
	})
	require.Equal(t, "; hello\n", out)
}

func TestWriterTrailingComment(t *testing.T) {
	t.Parallel()
	out := writeAll(t, []gcode.Line{
		gcode.NewLine("G28").WithComment(" home"),
	})
	require.Equal(t, "G28 ; home\n", out)
}

func TestWriterBlank(t *testing.T) {
	t.Parallel()
	out := writeAll(t, []gcode.Line{{}})
	require.Equal(t, "\n", out)
}

func TestWriterLineNumberAndChecksum(t *testing.T) {
	t.Parallel()
	out := writeAll(t,
		[]gcode.Line{gcode.NewLine("G1").LineNo(7).Arg("X", "10")},
		gcode.WithEmitLineNumbers(true),
		gcode.WithComputeChecksum(true),
	)
	// Verify shape; checksum value is whatever XOR of body produces.
	require.True(t, strings.HasPrefix(out, "N7 G1 X10*"))
	require.True(t, strings.HasSuffix(out, "\n"))
}

func TestWriterCRLF(t *testing.T) {
	t.Parallel()
	out := writeAll(t,
		[]gcode.Line{gcode.NewLine("G28")},
		gcode.WithLineEnding(gcode.LineEndingCRLF),
	)
	require.Equal(t, "G28\r\n", out)
}

func TestWriterEmitCommentsFalse(t *testing.T) {
	t.Parallel()
	out := writeAll(t,
		[]gcode.Line{gcode.NewLine("G28").WithComment(" home")},
		gcode.WithEmitComments(false),
	)
	require.Equal(t, "G28\n", out)
}

func TestWriterArgPrecision(t *testing.T) {
	t.Parallel()

	// ArgF stores the shortest float64 representation in Raw. With
	// per-key precision the Writer reformats from Number instead.
	line := gcode.NewLine("G1").
		ArgF("X", 10.1).
		ArgF("Y", 20.5).
		ArgF("Z", 0.20000000000000018). // float64 noise
		ArgF("E", 1.0336500000000002).
		ArgF("F", 1500)

	t.Run("default keeps Raw verbatim", func(t *testing.T) {
		out := writeAll(t, []gcode.Line{line})
		require.Equal(t,
			"G1 X10.1 Y20.5 Z0.20000000000000018 E1.0336500000000002 F1500\n",
			out)
	})

	t.Run("slicer-style precision", func(t *testing.T) {
		out := writeAll(t, []gcode.Line{line}, gcode.WithArgPrecision(map[string]int{
			"X": 3, "Y": 3, "Z": 3,
			"E": 5,
			"F": 0,
		}))
		require.Equal(t, "G1 X10.100 Y20.500 Z0.200 E1.03365 F1500\n", out)
	})

	t.Run("unconfigured keys pass through verbatim", func(t *testing.T) {
		out := writeAll(t, []gcode.Line{line}, gcode.WithArgPrecision(map[string]int{
			"E": 5,
		}))
		require.Equal(t,
			"G1 X10.1 Y20.5 Z0.20000000000000018 E1.03365 F1500\n",
			out)
	})

	t.Run("non-numeric Raw unaffected", func(t *testing.T) {
		l := gcode.NewLine("G1").Arg("X", "abc")
		out := writeAll(t, []gcode.Line{l}, gcode.WithArgPrecision(map[string]int{"X": 3}))
		require.Equal(t, "G1 Xabc\n", out)
	})

	t.Run("bare flag unaffected", func(t *testing.T) {
		l := gcode.NewLine("G28").Flag("X").Flag("Y")
		out := writeAll(t, []gcode.Line{l}, gcode.WithArgPrecision(map[string]int{"X": 3, "Y": 3}))
		require.Equal(t, "G28 X Y\n", out)
	})

	t.Run("extended args reformat too", func(t *testing.T) {
		l := gcode.NewLine("SET_FAN_SPEED").
			Arg("FAN", "cooling").
			ArgF("SPEED", 0.5000000000000001)
		out := writeAll(t, []gcode.Line{l}, gcode.WithArgPrecision(map[string]int{"SPEED": 2}))
		require.Equal(t, "SET_FAN_SPEED FAN=cooling SPEED=0.50\n", out)
	})

	t.Run("negative precision clamped to 0", func(t *testing.T) {
		out := writeAll(t,
			[]gcode.Line{gcode.NewLine("G1").ArgF("E", 1.789)},
			gcode.WithArgPrecision(map[string]int{"E": -5}),
		)
		require.Equal(t, "G1 E2\n", out)
	})
}
