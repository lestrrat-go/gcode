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
