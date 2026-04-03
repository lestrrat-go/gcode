package scan

import (
	"strings"
	"testing"

	gcode "github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	input := "G0 X10\nG1 Y20\n\n; comment\n"
	s := NewScanner(strings.NewReader(input))

	require.True(t, s.Scan())
	require.Equal(t, "G0 X10", s.Text())
	require.Equal(t, 1, s.LineNum())

	require.True(t, s.Scan())
	require.Equal(t, "G1 Y20", s.Text())
	require.Equal(t, 2, s.LineNum())

	require.True(t, s.Scan())
	require.Equal(t, "", s.Text())
	require.Equal(t, 3, s.LineNum())

	require.True(t, s.Scan())
	require.Equal(t, "; comment", s.Text())
	require.Equal(t, 4, s.LineNum())

	require.False(t, s.Scan())
	require.NoError(t, s.Err())
}

func TestParseLine_BlankLine(t *testing.T) {
	line, err := ParseLine("", 1)
	require.NoError(t, err)
	require.True(t, line.IsBlank())
	require.False(t, line.HasCommand)
	require.False(t, line.HasComment)
}

func TestParseLine_BlankLineWhitespace(t *testing.T) {
	line, err := ParseLine("   \t  ", 1)
	require.NoError(t, err)
	require.True(t, line.IsBlank())
}

func TestParseLine_SemicolonComment(t *testing.T) {
	line, err := ParseLine("; this is a comment", 1)
	require.NoError(t, err)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, " this is a comment", line.Comment.Text)
	require.Equal(t, gcode.CommentSemicolon, line.Comment.Form)
}

func TestParseLine_SemicolonCommentLeadingWhitespace(t *testing.T) {
	line, err := ParseLine("   ; indented comment", 2)
	require.NoError(t, err)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, " indented comment", line.Comment.Text)
	require.Equal(t, gcode.CommentSemicolon, line.Comment.Form)
}

func TestParseLine_ParenthesisComment(t *testing.T) {
	line, err := ParseLine("(this is a comment)", 1)
	require.NoError(t, err)
	require.False(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, "this is a comment", line.Comment.Text)
	require.Equal(t, gcode.CommentParenthesis, line.Comment.Form)
}

func TestParseLine_SimpleG0(t *testing.T) {
	line, err := ParseLine("G0 X10 Y20 Z0.5", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 0, line.Command.Number)
	require.False(t, line.Command.HasSubcode)
	require.Len(t, line.Command.Params, 3)
	require.Equal(t, gcode.Parameter{Letter: 'X', Value: 10}, line.Command.Params[0])
	require.Equal(t, gcode.Parameter{Letter: 'Y', Value: 20}, line.Command.Params[1])
	require.Equal(t, gcode.Parameter{Letter: 'Z', Value: 0.5}, line.Command.Params[2])
	require.Equal(t, "G0 X10 Y20 Z0.5", line.Raw)
}

func TestParseLine_G1FloatParams(t *testing.T) {
	line, err := ParseLine("G1 X10.5 Y-3.2 E0.04 F1200", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 4)
	require.Equal(t, gcode.Parameter{Letter: 'X', Value: 10.5}, line.Command.Params[0])
	require.Equal(t, gcode.Parameter{Letter: 'Y', Value: -3.2}, line.Command.Params[1])
	require.Equal(t, gcode.Parameter{Letter: 'E', Value: 0.04}, line.Command.Params[2])
	require.Equal(t, gcode.Parameter{Letter: 'F', Value: 1200}, line.Command.Params[3])
}

func TestParseLine_Subcode(t *testing.T) {
	line, err := ParseLine("G92.1", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 92, line.Command.Number)
	require.True(t, line.Command.HasSubcode)
	require.Equal(t, 1, line.Command.Subcode)
}

func TestParseLine_MCode(t *testing.T) {
	line, err := ParseLine("M104 S200", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('M'), line.Command.Letter)
	require.Equal(t, 104, line.Command.Number)
	require.Len(t, line.Command.Params, 1)
	require.Equal(t, gcode.Parameter{Letter: 'S', Value: 200}, line.Command.Params[0])
}

func TestParseLine_TCode(t *testing.T) {
	line, err := ParseLine("T0", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('T'), line.Command.Letter)
	require.Equal(t, 0, line.Command.Number)
	require.Len(t, line.Command.Params, 0)
}

func TestParseLine_LineNumber(t *testing.T) {
	line, err := ParseLine("N100 G1 X10 Y20", 1)
	require.NoError(t, err)
	require.Equal(t, 100, line.LineNumber)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, gcode.Parameter{Letter: 'X', Value: 10}, line.Command.Params[0])
	require.Equal(t, gcode.Parameter{Letter: 'Y', Value: 20}, line.Command.Params[1])
}

func TestParseLine_Checksum(t *testing.T) {
	// Compute expected checksum for "N100 G1 X10":
	// XOR of bytes 'N','1','0','0',' ','G','1',' ','X','1','0'
	raw := "N100 G1 X10"
	var cs byte
	for i := range len(raw) {
		cs ^= raw[i]
	}
	line, err := ParseLine("N100 G1 X10*"+itoa(int(cs)), 1)
	require.NoError(t, err)
	require.True(t, line.HasChecksum)
	require.Equal(t, cs, line.Checksum)
	require.Equal(t, 100, line.LineNumber)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
}

func TestParseLine_InlineSemicolonComment(t *testing.T) {
	line, err := ParseLine("G28 X Y ; home X and Y", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 28, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
	require.True(t, line.HasComment)
	require.Equal(t, " home X and Y", line.Comment.Text)
	require.Equal(t, gcode.CommentSemicolon, line.Comment.Form)
}

func TestParseLine_InlineParenthesisComment(t *testing.T) {
	line, err := ParseLine("G28 X Y (home axes)", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 28, line.Command.Number)
	require.True(t, line.HasComment)
	require.Equal(t, "home axes", line.Comment.Text)
	require.Equal(t, gcode.CommentParenthesis, line.Comment.Form)
}

func TestParseLine_CaseInsensitive(t *testing.T) {
	line, err := ParseLine("g1 x10 y20", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
}

func TestParseLine_NegativeParams(t *testing.T) {
	line, err := ParseLine("G1 X-5 Y-10.5", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, gcode.Parameter{Letter: 'X', Value: -5}, line.Command.Params[0])
	require.Equal(t, gcode.Parameter{Letter: 'Y', Value: -10.5}, line.Command.Params[1])
}

func TestParseLine_MalformedInput(t *testing.T) {
	_, err := ParseLine("G X10", 5)
	require.Error(t, err)
	require.ErrorIs(t, err, gcode.ErrParse)

	var detail gcode.ParseErrorDetail
	require.ErrorAs(t, err, &detail)
	require.Equal(t, 5, detail.Line())
}

func TestParseLine_CommandNoNumber(t *testing.T) {
	// A command letter followed by a non-digit is malformed.
	_, err := ParseLine("G ;oops", 3)
	require.Error(t, err)
	require.ErrorIs(t, err, gcode.ErrParse)
}

func TestParseLine_WhitespaceVariations(t *testing.T) {
	line, err := ParseLine("  G1   X10   Y20  ", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Equal(t, byte('G'), line.Command.Letter)
	require.Equal(t, 1, line.Command.Number)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, gcode.Parameter{Letter: 'X', Value: 10}, line.Command.Params[0])
	require.Equal(t, gcode.Parameter{Letter: 'Y', Value: 20}, line.Command.Params[1])
}

func TestParseLine_BothComments_SemicolonWins(t *testing.T) {
	// If line has both (...) and ; comments, keep ; and discard (...)
	line, err := ParseLine("G28 (paren comment) ; semi comment", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.True(t, line.HasComment)
	require.Equal(t, gcode.CommentSemicolon, line.Comment.Form)
	require.Equal(t, " semi comment", line.Comment.Text)
}

func TestParseLine_ParamValueOnlyLetters(t *testing.T) {
	// G28 takes flag-style params (X, Y with no value = 0)
	line, err := ParseLine("G28 X Y", 1)
	require.NoError(t, err)
	require.True(t, line.HasCommand)
	require.Len(t, line.Command.Params, 2)
	require.Equal(t, byte('X'), line.Command.Params[0].Letter)
	require.Equal(t, float64(0), line.Command.Params[0].Value)
	require.Equal(t, byte('Y'), line.Command.Params[1].Letter)
	require.Equal(t, float64(0), line.Command.Params[1].Value)
}

// itoa is a simple int-to-string helper for test readability.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
