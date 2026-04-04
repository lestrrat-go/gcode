package gcode_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func compareLine(t *testing.T, idx int, a, b gcode.Line) {
	t.Helper()

	require.Equal(t, a.LineNumber, b.LineNumber, "line %d: LineNumber mismatch", idx)
	require.Equal(t, a.HasCommand, b.HasCommand, "line %d: HasCommand mismatch", idx)
	if a.HasCommand {
		require.Equal(t, a.Command.Letter, b.Command.Letter, "line %d: Command.Letter mismatch", idx)
		require.Equal(t, a.Command.Number, b.Command.Number, "line %d: Command.Number mismatch", idx)
		require.Equal(t, a.Command.HasSubcode, b.Command.HasSubcode, "line %d: Command.HasSubcode mismatch", idx)
		if a.Command.HasSubcode {
			require.Equal(t, a.Command.Subcode, b.Command.Subcode, "line %d: Command.Subcode mismatch", idx)
		}
		require.Equal(t, len(a.Command.Params), len(b.Command.Params), "line %d: Params length mismatch", idx)
		for j := range a.Command.Params {
			require.Equal(t, a.Command.Params[j].Letter, b.Command.Params[j].Letter, "line %d param %d: Letter mismatch", idx, j)
			require.InDelta(t, a.Command.Params[j].Value, b.Command.Params[j].Value, 1e-9, "line %d param %d: Value mismatch", idx, j)
		}
	}
	require.Equal(t, a.HasComment, b.HasComment, "line %d: HasComment mismatch", idx)
	if a.HasComment {
		require.Equal(t, a.Comment.Text, b.Comment.Text, "line %d: Comment.Text mismatch", idx)
		require.Equal(t, a.Comment.Form, b.Comment.Form, "line %d: Comment.Form mismatch", idx)
	}
	require.Equal(t, a.HasChecksum, b.HasChecksum, "line %d: HasChecksum mismatch", idx)
	if a.HasChecksum {
		require.Equal(t, a.Checksum, b.Checksum, "line %d: Checksum mismatch", idx)
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		fmtOpts []gcode.FormatOption
	}{
		{
			name: "marlin_start",
			file: "marlin_start.gcode",
		},
		{
			name: "line_numbers",
			file: "line_numbers.gcode",
			fmtOpts: []gcode.FormatOption{
				gcode.WithEmitLineNumbers(true),
				gcode.WithComputeChecksum(true),
			},
		},
		{
			name: "mixed_comments",
			file: "mixed_comments.gcode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.file))
			require.NoError(t, err)

			// First parse
			prog1, err := gcode.ParseBytes(data)
			require.NoError(t, err)

			// Format
			var sb strings.Builder
			require.NoError(t, gcode.Format(&sb, prog1, tt.fmtOpts...))
			output := sb.String()
			require.NotEmpty(t, output)

			// Second parse
			prog2, err := gcode.ParseString(output)
			require.NoError(t, err)

			// Compare field-by-field
			lines1 := prog1.Lines()
			lines2 := prog2.Lines()
			require.Equal(t, len(lines1), len(lines2), "line count mismatch")

			for i := range lines1 {
				compareLine(t, i, lines1[i], lines2[i])
			}
		})
	}
}
