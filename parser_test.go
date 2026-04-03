package gcode_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestParser_BlankLinesPreserved(t *testing.T) {
	input := "G0 X10\n\n\nG1 Y20\n"
	prog, err := gcode.ParseString(input)
	require.NoError(t, err)
	require.Equal(t, 4, prog.Len())
	require.True(t, prog.Line(0).HasCommand)
	require.True(t, prog.Line(1).IsBlank())
	require.True(t, prog.Line(2).IsBlank())
	require.True(t, prog.Line(3).HasCommand)
}

func TestParser_SemicolonCommentOnlyLine(t *testing.T) {
	prog, err := gcode.ParseString("; this is a comment\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, gcode.CommentSemicolon, line.Comment.Form)
	require.Equal(t, " this is a comment", line.Comment.Text)
}

func TestParser_ParenthesisCommentOnlyLine(t *testing.T) {
	prog, err := gcode.ParseString("(this is a comment)\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, gcode.CommentParenthesis, line.Comment.Form)
	require.Equal(t, "this is a comment", line.Comment.Text)
}

func TestParser_SimpleG0G1WithFloatParams(t *testing.T) {
	prog, err := gcode.ParseString("G0 X10.5 Y-3.2\nG1 E0.04 F1200\n")
	require.NoError(t, err)
	require.Equal(t, 2, prog.Len())

	l0 := prog.Line(0)
	require.True(t, l0.HasCommand)
	require.Equal(t, byte('G'), l0.Command.Letter)
	require.Equal(t, 0, l0.Command.Number)
	require.Equal(t, 2, len(l0.Command.Params))
	require.Equal(t, byte('X'), l0.Command.Params[0].Letter)
	require.InDelta(t, 10.5, l0.Command.Params[0].Value, 1e-9)
	require.Equal(t, byte('Y'), l0.Command.Params[1].Letter)
	require.InDelta(t, -3.2, l0.Command.Params[1].Value, 1e-9)

	l1 := prog.Line(1)
	require.True(t, l1.HasCommand)
	require.Equal(t, byte('G'), l1.Command.Letter)
	require.Equal(t, 1, l1.Command.Number)
	require.Equal(t, 2, len(l1.Command.Params))
	require.InDelta(t, 0.04, l1.Command.Params[0].Value, 1e-9)
	require.InDelta(t, 1200.0, l1.Command.Params[1].Value, 1e-9)
}

func TestParser_MCodeCommand(t *testing.T) {
	prog, err := gcode.ParseString("M104 S200\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('M'), line.Command.Letter)
	require.Equal(t, 104, line.Command.Number)
	require.Equal(t, 1, len(line.Command.Params))
	require.Equal(t, byte('S'), line.Command.Params[0].Letter)
	require.InDelta(t, 200.0, line.Command.Params[0].Value, 1e-9)
}

func TestParser_TCode(t *testing.T) {
	prog, err := gcode.ParseString("T0\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('T'), line.Command.Letter)
	require.Equal(t, 0, line.Command.Number)
}

func TestParser_LineNumberPrefix(t *testing.T) {
	prog, err := gcode.ParseString("N100 G1 X10\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.Equal(t, 100, line.LineNumber)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
}

func TestParser_ChecksumSuffix(t *testing.T) {
	prog, err := gcode.ParseString("N100 G1 X10*42\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasChecksum)
	require.Equal(t, byte(42), line.Checksum)
}

func TestParser_InlineTrailingSemicolonComment(t *testing.T) {
	prog, err := gcode.ParseString("G28 X0 Y0 ;home axes\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, gcode.CommentSemicolon, line.Comment.Form)
	require.Equal(t, "home axes", line.Comment.Text)
}

func TestParser_InlineTrailingParenthesisComment(t *testing.T) {
	prog, err := gcode.ParseString("G28 X0 Y0 (home axes)\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, gcode.CommentParenthesis, line.Comment.Form)
	require.Equal(t, "home axes", line.Comment.Text)
}

func TestParser_MalformedInput(t *testing.T) {
	_, err := gcode.ParseString("G\n")
	require.Error(t, err)
	require.True(t, errors.Is(err, gcode.ErrParse))
}

func TestParser_StrictModeUnknownCommand(t *testing.T) {
	d := gcode.NewDialect("test")
	d.Register(gcode.CommandDef{Letter: 'G', Number: 0})

	_, err := gcode.ParseString("M999\n", gcode.WithStrict(), gcode.WithDialect(d))
	require.Error(t, err)
	require.True(t, errors.Is(err, gcode.ErrParse))
}

func TestParser_StrictModeKnownCommand(t *testing.T) {
	d := gcode.NewDialect("test")
	d.Register(gcode.CommandDef{Letter: 'G', Number: 0})

	prog, err := gcode.ParseString("G0 X10\n", gcode.WithStrict(), gcode.WithDialect(d))
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
}

func TestParser_NonStrictUnknownCommand(t *testing.T) {
	prog, err := gcode.ParseString("M999\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('M'), line.Command.Letter)
	require.Equal(t, 999, line.Command.Number)
}

func TestParser_MultiLineProgram(t *testing.T) {
	input := strings.Join([]string{
		"G28",
		"G0 X0 Y0",
		"; move to start",
		"G1 X100 Y200 F1500",
		"M104 S200",
	}, "\n") + "\n"

	prog, err := gcode.ParseString(input)
	require.NoError(t, err)
	require.Equal(t, 5, prog.Len())

	require.Equal(t, byte('G'), prog.Line(0).Command.Letter)
	require.Equal(t, 28, prog.Line(0).Command.Number)

	require.Equal(t, byte('G'), prog.Line(1).Command.Letter)
	require.Equal(t, 0, prog.Line(1).Command.Number)

	require.True(t, prog.Line(2).HasComment)
	require.False(t, prog.Line(2).HasCommand)

	require.Equal(t, byte('G'), prog.Line(3).Command.Letter)
	require.Equal(t, 1, prog.Line(3).Command.Number)

	require.Equal(t, byte('M'), prog.Line(4).Command.Letter)
	require.Equal(t, 104, prog.Line(4).Command.Number)
}

func TestParseString(t *testing.T) {
	prog, err := gcode.ParseString("G0 X10\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	require.Equal(t, byte('G'), prog.Line(0).Command.Letter)
}

func TestParseBytes(t *testing.T) {
	prog, err := gcode.ParseBytes([]byte("G0 X10\n"))
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	require.Equal(t, byte('G'), prog.Line(0).Command.Letter)
}

func TestParser_CommandWithSubcode(t *testing.T) {
	prog, err := gcode.ParseString("G92.1\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 92, line.Command.Number)
	require.True(t, line.Command.HasSubcode)
	require.Equal(t, 1, line.Command.Subcode)
}

func TestParser_CaseInsensitivity(t *testing.T) {
	prog, err := gcode.ParseString("g0 x10 y20\n")
	require.NoError(t, err)
	require.Equal(t, 1, prog.Len())
	line := prog.Line(0)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 0, line.Command.Number)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
}
