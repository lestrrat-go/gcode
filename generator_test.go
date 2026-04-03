package gcode_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestGenerateLine_BlankLine(t *testing.T) {
	g := gcode.NewGenerator()
	var buf bytes.Buffer
	err := g.GenerateLine(&buf, gcode.Line{})
	require.NoError(t, err)
	require.Equal(t, "", buf.String())
}

func TestGenerateLine_CommentOnlySemicolon(t *testing.T) {
	g := gcode.NewGenerator()
	var buf bytes.Buffer
	line := gcode.Line{
		HasComment: true,
		Comment: gcode.Comment{
			Text: "comment text",
			Form: gcode.CommentSemicolon,
		},
	}
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, ";comment text", buf.String())
}

func TestGenerateLine_CommentOnlyParenthesis(t *testing.T) {
	g := gcode.NewGenerator()
	var buf bytes.Buffer
	line := gcode.Line{
		HasComment: true,
		Comment: gcode.Comment{
			Text: "comment text",
			Form: gcode.CommentParenthesis,
		},
	}
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "(comment text)", buf.String())
}

func TestGenerateLine_EmitCommentsFalse(t *testing.T) {
	g := gcode.NewGenerator(gcode.WithEmitComments(false))
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G28", buf.String())
}

func TestGenerateLine_SimpleCommand(t *testing.T) {
	g := gcode.NewGenerator()
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G0 X10 Y20", buf.String())
}

func TestGenerateLine_CommandWithSubcode(t *testing.T) {
	g := gcode.NewGenerator()
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G92.1", buf.String())
}

func TestGenerateLine_CommandWithFloatParams(t *testing.T) {
	g := gcode.NewGenerator()
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G1 X10.5 Y-3.2 E0.04 F1200", buf.String())
}

func TestGenerateLine_LineNumbersEnabled(t *testing.T) {
	g := gcode.NewGenerator(gcode.WithEmitLineNumbers(true))
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "N100 G1 X10 Y20", buf.String())
}

func TestGenerateLine_LineNumbersDisabled(t *testing.T) {
	g := gcode.NewGenerator()
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G1 X10 Y20", buf.String())
}

func TestGenerateLine_Checksum(t *testing.T) {
	g := gcode.NewGenerator(gcode.WithEmitLineNumbers(true), gcode.WithComputeChecksum(true))
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
	err := g.GenerateLine(&buf, line)
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

func TestGenerateLine_CRLFLineEnding(t *testing.T) {
	g := gcode.NewGenerator(gcode.WithLineEnding(gcode.LineEndingCRLF))

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
	err := g.Generate(&buf, prog)
	require.NoError(t, err)
	require.Equal(t, "G0\r\nG1\r\n", buf.String())
}

func TestGenerateLine_LFLineEnding(t *testing.T) {
	g := gcode.NewGenerator()

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
	err := g.Generate(&buf, prog)
	require.NoError(t, err)
	require.Equal(t, "G0\nG1\n", buf.String())
}

func TestGenerateString(t *testing.T) {
	pb := gcode.NewProgramBuilder()
	pb.Append(gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 0,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 10},
			},
		},
	})
	prog := pb.Build()

	s, err := gcode.GenerateString(prog)
	require.NoError(t, err)
	require.Equal(t, "G0 X10\n", s)
}

func TestGenerate_MultipleLines(t *testing.T) {
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
	g := gcode.NewGenerator()
	err := g.Generate(&buf, prog)
	require.NoError(t, err)
	require.Equal(t, "G28\n\nG1 X10 Y20\n", buf.String())
}

func TestGenerateLine_CommandWithTrailingSemicolonComment(t *testing.T) {
	g := gcode.NewGenerator()
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G28 X0 Y0 ;home", buf.String())
}

func TestGenerateLine_CommandWithTrailingParenComment(t *testing.T) {
	g := gcode.NewGenerator()
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
	err := g.GenerateLine(&buf, line)
	require.NoError(t, err)
	require.Equal(t, "G28 X0 Y0 (home)", buf.String())
}
