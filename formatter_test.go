package gcode_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestFormatLine_BlankLine(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	err := g.FormatLine(&buf, gcode.Line{})
	require.NoError(t, err)
	require.Equal(t, "", buf.String())
}

func TestFormatLine_CommentOnlySemicolon(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasComment: true,
		Comment: gcode.Comment{
			Text: "comment text",
			Form: gcode.CommentSemicolon,
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, ";comment text", buf.String())
}

func TestFormatLine_CommentOnlyParenthesis(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasComment: true,
		Comment: gcode.Comment{
			Text: "comment text",
			Form: gcode.CommentParenthesis,
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "(comment text)", buf.String())
}

func TestFormatLine_EmitCommentsFalse(t *testing.T) {
	g := gcode.NewFormatter(gcode.WithEmitComments(false))
	var buf bytes.Buffer
	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 28,
		},
		HasComment: true,
		Comment: gcode.Comment{
			Text: "home",
			Form: gcode.CommentSemicolon,
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G28", buf.String())
}

func TestFormatLine_SimpleCommand(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 0,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 10},
				{Letter: 'Y', Value: 20},
			},
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G0 X10 Y20", buf.String())
}

func TestFormatLine_CommandWithSubcode(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter:     'G',
			Number:     92,
			Subcode:    1,
			HasSubcode: true,
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G92.1", buf.String())
}

func TestFormatLine_CommandWithFloatParams(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 1,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 10.5},
				{Letter: 'Y', Value: -3.2},
				{Letter: 'E', Value: 0.04},
				{Letter: 'F', Value: 1200},
			},
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G1 X10.5 Y-3.2 E0.04 F1200", buf.String())
}

func TestFormatLine_LineNumbersEnabled(t *testing.T) {
	g := gcode.NewFormatter(gcode.WithEmitLineNumbers(true))
	var buf bytes.Buffer
	line := gcode.Line{
		LineNumber: 100,
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 1,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 10},
				{Letter: 'Y', Value: 20},
			},
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "N100 G1 X10 Y20", buf.String())
}

func TestFormatLine_LineNumbersDisabled(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		LineNumber: 100,
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 1,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 10},
				{Letter: 'Y', Value: 20},
			},
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G1 X10 Y20", buf.String())
}

func TestFormatLine_Checksum(t *testing.T) {
	g := gcode.NewFormatter(gcode.WithEmitLineNumbers(true), gcode.WithComputeChecksum(true))
	var buf bytes.Buffer
	line := gcode.Line{
		LineNumber: 100,
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 1,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 10},
				{Letter: 'Y', Value: 20},
			},
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)

	// Compute expected checksum: XOR all bytes of "N100 G1 X10 Y20"
	content := "N100 G1 X10 Y20"
	var cs byte
	for i := range len(content) {
		cs ^= content[i]
	}
	expected := fmt.Sprintf("%s*%d", content, cs)
	require.Equal(t, expected, buf.String())
}

func TestFormatLine_CRLFLineEnding(t *testing.T) {
	g := gcode.NewFormatter(gcode.WithLineEnding(gcode.LineEndingCRLF))

	pb := gcode.NewProgramBuilder()
	pb.Append(
		gcode.Line{
			HasCommand: true,
			Command:    gcode.Command{Letter: 'G', Number: 0},
		},
		gcode.Line{
			HasCommand: true,
			Command:    gcode.Command{Letter: 'G', Number: 1},
		},
	)
	prog := pb.Build()

	var buf bytes.Buffer
	err := g.Format(&buf, prog)
	require.NoError(t, err)
	require.Equal(t, "G0\r\nG1\r\n", buf.String())
}

func TestFormatLine_LFLineEnding(t *testing.T) {
	g := gcode.NewFormatter()

	pb := gcode.NewProgramBuilder()
	pb.Append(
		gcode.Line{
			HasCommand: true,
			Command:    gcode.Command{Letter: 'G', Number: 0},
		},
		gcode.Line{
			HasCommand: true,
			Command:    gcode.Command{Letter: 'G', Number: 1},
		},
	)
	prog := pb.Build()

	var buf bytes.Buffer
	err := g.Format(&buf, prog)
	require.NoError(t, err)
	require.Equal(t, "G0\nG1\n", buf.String())
}

func TestFormat_String(t *testing.T) {
	pb := gcode.NewProgramBuilder()
	pb.Append(gcode.G(0).X(10).Build())
	prog := pb.Build()

	var sb strings.Builder
	err := gcode.Format(&sb, prog)
	require.NoError(t, err)
	require.Equal(t, "G0 X10\n", sb.String())
}

func TestFormat_MultipleLines(t *testing.T) {
	pb := gcode.NewProgramBuilder()
	pb.Append(
		gcode.Line{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 28,
			},
		},
		gcode.Line{}, // blank line
		gcode.Line{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 1,
				Params: []gcode.Parameter{
					{Letter: 'X', Value: 10},
					{Letter: 'Y', Value: 20},
				},
			},
		},
	)
	prog := pb.Build()

	var buf bytes.Buffer
	g := gcode.NewFormatter()
	err := g.Format(&buf, prog)
	require.NoError(t, err)
	require.Equal(t, "G28\n\nG1 X10 Y20\n", buf.String())
}

func TestFormatLine_CommandWithTrailingSemicolonComment(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 28,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 0},
				{Letter: 'Y', Value: 0},
			},
		},
		HasComment: true,
		Comment: gcode.Comment{
			Text: "home",
			Form: gcode.CommentSemicolon,
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G28 X0 Y0 ;home", buf.String())
}

func TestFormatLine_CommandWithTrailingParenComment(t *testing.T) {
	g := gcode.NewFormatter()
	var buf bytes.Buffer
	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 28,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 0},
				{Letter: 'Y', Value: 0},
			},
		},
		HasComment: true,
		Comment: gcode.Comment{
			Text: "home",
			Form: gcode.CommentParenthesis,
		},
	}
	err := g.FormatLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G28 X0 Y0 (home)", buf.String())
}
