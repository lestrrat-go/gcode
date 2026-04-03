package gcode

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineScanner(t *testing.T) {
	input := "G0 X10\nG1 Y20\n\n; comment\n"
	s := newLineScanner(strings.NewReader(input))

	require.True(t, s.scan())
	require.Equal(t, "G0 X10", s.text())
	require.Equal(t, 1, s.lineNumber())

	require.True(t, s.scan())
	require.Equal(t, "G1 Y20", s.text())
	require.Equal(t, 2, s.lineNumber())

	require.True(t, s.scan())
	require.Equal(t, "", s.text())
	require.Equal(t, 3, s.lineNumber())

	require.True(t, s.scan())
	require.Equal(t, "; comment", s.text())
	require.Equal(t, 4, s.lineNumber())

	require.False(t, s.scan())
	require.NoError(t, s.err())
}

func TestParseLine_BlankLine(t *testing.T) {
	line, err := parseLine("", 1)
	require.NoError(t, err)
	require.True(t, line.IsBlank())
	require.False(t, line.HasCommand)
	require.False(t, line.HasComment)
}

func TestParseLine_BlankLineWhitespace(t *testing.T) {
	line, err := parseLine("   \t  ", 1)
	require.NoError(t, err)
	require.True(t, line.IsBlank())
}

func TestParseLine_SemicolonComment(t *testing.T) {
	line, err := parseLine("; this is a comment", 1)
	require.NoError(t, err)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, " this is a comment", line.Comment.Text)
	require.Equal(t, CommentSemicolon, line.Comment.Form)
}

func TestParseLine_SemicolonCommentLeadingWhitespace(t *testing.T) {
	line, err := parseLine("   ; indented comment", 2)
	require.NoError(t, err)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, " indented comment", line.Comment.Text)
	require.Equal(t, CommentSemicolon, line.Comment.Form)
}

func TestParseLine_ParenthesisComment(t *testing.T) {
	line, err := parseLine("(this is a comment)", 1)
	require.NoError(t, err)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, "this is a comment", line.Comment.Text)
	require.Equal(t, CommentParenthesis, line.Comment.Form)
}

func TestParseLine_SimpleG0(t *testing.T) {
	line, err := parseLine("G0 X10 Y20 Z0.5", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 0, line.Command.Number)
	require.False(t, line.Command.HasSubcode)
	require.Len(t, line.Command.Params, 3)
	require.Equal(t, Parameter{Letter: 'X', Value: 10}, line.Command.Params[0])
	require.Equal(t, Parameter{Letter: 'Y', Value: 20}, line.Command.Params[1])
	require.Equal(t, Parameter{Letter: 'Z', Value: 0.5}, line.Command.Params[2])
	require.Equal(t, "G0 X10 Y20 Z0.5", line.Raw)
}

func TestParseLine_G1FloatParams(t *testing.T) {
	line, err := parseLine("G1 X10.5 Y-3.2 E0.04 F1200", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 4)
	require.Equal(t, Parameter{Letter: 'X', Value: 10.5}, line.Command.Params[0])
	require.Equal(t, Parameter{Letter: 'Y', Value: -3.2}, line.Command.Params[1])
	require.Equal(t, Parameter{Letter: 'E', Value: 0.04}, line.Command.Params[2])
	require.Equal(t, Parameter{Letter: 'F', Value: 1200}, line.Command.Params[3])
}

func TestParseLine_Subcode(t *testing.T) {
	line, err := parseLine("G92.1", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 92, line.Command.Number)
	require.True(t, line.Command.HasSubcode)
	require.Equal(t, 1, line.Command.Subcode)
}

func TestParseLine_MCode(t *testing.T) {
	line, err := parseLine("M104 S200", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('M'), line.Command.Letter)
	require.Equal(t, 104, line.Command.Number)
	require.Len(t, line.Command.Params, 1)
	require.Equal(t, Parameter{Letter: 'S', Value: 200}, line.Command.Params[0])
}

func TestParseLine_TCode(t *testing.T) {
	line, err := parseLine("T0", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('T'), line.Command.Letter)
	require.Equal(t, 0, line.Command.Number)
	require.Len(t, line.Command.Params, 0)
}

func TestParseLine_LineNumber(t *testing.T) {
	line, err := parseLine("N100 G1 X10 Y20", 1)
	require.NoError(t, err)
	require.Equal(t, 100, line.LineNumber)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, Parameter{Letter: 'X', Value: 10}, line.Command.Params[0])
	require.Equal(t, Parameter{Letter: 'Y', Value: 20}, line.Command.Params[1])
}

func TestParseLine_Checksum(t *testing.T) {
	raw := "N100 G1 X10"
	var cs byte
	for i := range len(raw) {
		cs ^= raw[i]
	}
	line, err := parseLine("N100 G1 X10*"+strconv.Itoa(int(cs)), 1)
	require.NoError(t, err)
	require.True(t, line.HasChecksum)
	require.Equal(t, cs, line.Checksum)
	require.Equal(t, 100, line.LineNumber)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
}

func TestParseLine_InlineSemicolonComment(t *testing.T) {
	line, err := parseLine("G28 X Y ; home X and Y", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 28, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
	require.True(t, line.HasComment)
	require.Equal(t, " home X and Y", line.Comment.Text)
	require.Equal(t, CommentSemicolon, line.Comment.Form)
}

func TestParseLine_InlineParenthesisComment(t *testing.T) {
	line, err := parseLine("G28 X Y (home axes)", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 28, line.Command.Number)
	require.True(t, line.HasComment)
	require.Equal(t, "home axes", line.Comment.Text)
	require.Equal(t, CommentParenthesis, line.Comment.Form)
}

func TestParseLine_CaseInsensitive(t *testing.T) {
	line, err := parseLine("g1 x10 y20", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
}

func TestParseLine_NegativeParams(t *testing.T) {
	line, err := parseLine("G1 X-5 Y-10.5", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, Parameter{Letter: 'X', Value: -5}, line.Command.Params[0])
	require.Equal(t, Parameter{Letter: 'Y', Value: -10.5}, line.Command.Params[1])
}

func TestParseLine_MalformedInput(t *testing.T) {
	_, err := parseLine("G X10", 5)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrParse)

	var detail ParseErrorDetail
	require.ErrorAs(t, err, &detail)
	require.Equal(t, 5, detail.Line())
}

func TestParseLine_CommandNoNumber(t *testing.T) {
	_, err := parseLine("G ;oops", 3)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrParse)
}

func TestParseLine_WhitespaceVariations(t *testing.T) {
	line, err := parseLine("  G1   X10   Y20  ", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, Parameter{Letter: 'X', Value: 10}, line.Command.Params[0])
	require.Equal(t, Parameter{Letter: 'Y', Value: 20}, line.Command.Params[1])
}

func TestParseLine_BothComments_SemicolonWins(t *testing.T) {
	line, err := parseLine("G28 (paren comment) ; semi comment", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, CommentSemicolon, line.Comment.Form)
	require.Equal(t, " semi comment", line.Comment.Text)
}

func TestParseLine_FlagStyleParams(t *testing.T) {
	line, err := parseLine("G28 X Y", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, float64(0), line.Command.Params[0].Value)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
	require.Equal(t, float64(0), line.Command.Params[1].Value)
}
