package gcode_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/klipper"
	"github.com/lestrrat-go/gcode/dialects/marlin"
	"github.com/stretchr/testify/require"
)

func readAll(t *testing.T, src string, opts ...gcode.ReadOption) []gcode.Line {
	t.Helper()
	r := gcode.NewReader(strings.NewReader(src), opts...)
	var out []gcode.Line
	var line gcode.Line
	for {
		err := r.Read(&line)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		out = append(out, line.Clone())
	}
	return out
}

func TestReaderClassic(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "G28\nG1 X10 Y20 F1500\nM104 S200\n")
	require.Len(t, lines, 3)

	require.True(t, lines[0].HasCommand)
	require.Equal(t, "G28", lines[0].Command.Name)
	require.Empty(t, lines[0].Command.Args)

	require.Equal(t, "G1", lines[1].Command.Name)
	require.Len(t, lines[1].Command.Args, 3)
	require.Equal(t, "X", lines[1].Command.Args[0].Key)
	require.True(t, lines[1].Command.Args[0].IsNumeric)
	require.InDelta(t, 10.0, lines[1].Command.Args[0].Number, 1e-9)

	require.Equal(t, "M104", lines[2].Command.Name)
	require.InDelta(t, 200.0, lines[2].Command.Args[0].Number, 1e-9)
}

func TestReaderSubcode(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "G92.1\n")
	require.Equal(t, "G92.1", lines[0].Command.Name)
}

func TestReaderClassicLowercase(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "g1 x10 y20\n")
	require.Equal(t, "G1", lines[0].Command.Name)
	require.Equal(t, "X", lines[0].Command.Args[0].Key)
	require.Equal(t, "Y", lines[0].Command.Args[1].Key)
}

func TestReaderClassicFlag(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "G28 X Y\n")
	require.Len(t, lines[0].Command.Args, 2)
	require.True(t, lines[0].Command.Args[0].IsFlag())
	require.True(t, lines[0].Command.Args[1].IsFlag())
}

func TestReaderNumberForms(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "G1 X-10.5 Y+20 Z.25 E0\n")
	args := lines[0].Command.Args
	require.InDelta(t, -10.5, args[0].Number, 1e-9)
	require.InDelta(t, 20.0, args[1].Number, 1e-9)
	require.InDelta(t, 0.25, args[2].Number, 1e-9)
	require.InDelta(t, 0.0, args[3].Number, 1e-9)
	require.True(t, args[3].IsNumeric)
}

func TestReaderLineNumberAndChecksum(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "N20 G1 X10*42\n")
	require.Equal(t, 20, lines[0].LineNumber)
	require.True(t, lines[0].HasChecksum)
	require.Equal(t, byte(42), lines[0].Checksum)
}

func TestReaderSemicolonComment(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "; just a comment\n")
	require.True(t, lines[0].HasComment)
	require.False(t, lines[0].HasCommand)
	require.Equal(t, gcode.CommentSemicolon, lines[0].Comment.Form)
	require.Equal(t, " just a comment", lines[0].Comment.Text)
}

func TestReaderParenthesisComment(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "(a paren comment)\n")
	require.True(t, lines[0].HasComment)
	require.Equal(t, gcode.CommentParenthesis, lines[0].Comment.Form)
	require.Equal(t, "a paren comment", lines[0].Comment.Text)
}

func TestReaderTrailingComments(t *testing.T) {
	t.Parallel()

	t.Run("semicolon", func(t *testing.T) {
		lines := readAll(t, "G28 ; home\n")
		require.True(t, lines[0].HasCommand)
		require.True(t, lines[0].HasComment)
		require.Equal(t, " home", lines[0].Comment.Text)
	})

	t.Run("parenthesis", func(t *testing.T) {
		lines := readAll(t, "G28 (home)\n")
		require.True(t, lines[0].HasComment)
		require.Equal(t, gcode.CommentParenthesis, lines[0].Comment.Form)
		require.Equal(t, "home", lines[0].Comment.Text)
	})
}

func TestReaderBlankLines(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "\n\nG28\n\n")
	require.Len(t, lines, 4)
	require.True(t, lines[0].IsBlank())
	require.True(t, lines[1].IsBlank())
	require.False(t, lines[2].IsBlank())
	require.True(t, lines[3].IsBlank())
}

func TestReaderExtendedCommand(t *testing.T) {
	t.Parallel()
	src := `EXCLUDE_OBJECT_DEFINE NAME=part_0 CENTER=120,120 POLYGON=[[1,2],[3,4]]
SET_FAN_SPEED FAN=cooling SPEED=0.5
TIMELAPSE_TAKE_FRAME
`
	lines := readAll(t, src)
	require.Len(t, lines, 3)

	require.Equal(t, "EXCLUDE_OBJECT_DEFINE", lines[0].Command.Name)
	require.Len(t, lines[0].Command.Args, 3)
	require.Equal(t, "NAME", lines[0].Command.Args[0].Key)
	require.Equal(t, "part_0", lines[0].Command.Args[0].Raw)
	require.Equal(t, "CENTER", lines[0].Command.Args[1].Key)
	require.Equal(t, "120,120", lines[0].Command.Args[1].Raw)
	require.Equal(t, "POLYGON", lines[0].Command.Args[2].Key)
	require.Equal(t, "[[1,2],[3,4]]", lines[0].Command.Args[2].Raw)

	require.Equal(t, "SET_FAN_SPEED", lines[1].Command.Name)
	require.Equal(t, "FAN", lines[1].Command.Args[0].Key)
	require.Equal(t, "cooling", lines[1].Command.Args[0].Raw)
	require.Equal(t, "SPEED", lines[1].Command.Args[1].Key)
	require.True(t, lines[1].Command.Args[1].IsNumeric)
	require.InDelta(t, 0.5, lines[1].Command.Args[1].Number, 1e-9)

	require.Equal(t, "TIMELAPSE_TAKE_FRAME", lines[2].Command.Name)
	require.Empty(t, lines[2].Command.Args)
}

func TestReaderExtendedPreservesArgKeyCase(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "BED_MESH_CALIBRATE mesh_min=1,2 ALGORITHM=bicubic\n")
	require.Equal(t, "BED_MESH_CALIBRATE", lines[0].Command.Name)
	require.Equal(t, "mesh_min", lines[0].Command.Args[0].Key)
	require.Equal(t, "ALGORITHM", lines[0].Command.Args[1].Key)
}

func TestReaderExtendedQuotedValue(t *testing.T) {
	t.Parallel()
	lines := readAll(t, `SET_DISPLAY_TEXT MSG="Hello World"`+"\n")
	require.Equal(t, "MSG", lines[0].Command.Args[0].Key)
	require.Equal(t, `"Hello World"`, lines[0].Command.Args[0].Raw)
}

func TestReaderClassicStringArgIsTolerant(t *testing.T) {
	t.Parallel()
	lines := readAll(t, "M117 Hello World\n")
	require.Equal(t, "M117", lines[0].Command.Name)
	require.Empty(t, lines[0].Command.Args)
	require.Equal(t, "M117 Hello World", lines[0].Raw)
}

func TestReaderStrictUnknownCommand(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(
		strings.NewReader("G999\n"),
		gcode.WithDialect(marlin.Dialect()),
		gcode.WithStrict(),
	)
	var line gcode.Line
	err := r.Read(&line)
	require.Error(t, err)
	require.True(t, errors.Is(err, gcode.ErrParse))
}

func TestReaderStrictKnownExtended(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(
		strings.NewReader("EXCLUDE_OBJECT_DEFINE NAME=foo\n"),
		gcode.WithDialect(klipper.Dialect()),
		gcode.WithStrict(),
	)
	var line gcode.Line
	require.NoError(t, r.Read(&line))
	require.Equal(t, "EXCLUDE_OBJECT_DEFINE", line.Command.Name)
}

func TestReaderClassicNoDigitsErrors(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(strings.NewReader("$\n"))
	var line gcode.Line
	err := r.Read(&line)
	require.Error(t, err)
	require.True(t, errors.Is(err, gcode.ErrParse))
}

func TestReaderUnclosedParenComment(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(strings.NewReader("(open\n"))
	var line gcode.Line
	err := r.Read(&line)
	require.Error(t, err)
}

func TestReaderAllIterator(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(strings.NewReader("G28\nG1 X10\n"))
	count := 0
	for line, err := range r.All() {
		require.NoError(t, err)
		require.True(t, line.HasCommand)
		count++
	}
	require.Equal(t, 2, count)
}

func TestReaderReuseInvalidation(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(strings.NewReader("G28\nG1 X10\n"))
	var line gcode.Line
	require.NoError(t, r.Read(&line))
	first := line.Command.Name // shares Reader buffer
	require.Equal(t, "G28", first)

	require.NoError(t, r.Read(&line))
	// After a second Read, line.Command.Name now points at the new line.
	require.Equal(t, "G1", line.Command.Name)
}

func TestReaderCloneRetains(t *testing.T) {
	t.Parallel()
	r := gcode.NewReader(strings.NewReader("G28 X10\nG1 Y20\n"))
	var line gcode.Line
	require.NoError(t, r.Read(&line))
	saved := line.Clone()
	require.NoError(t, r.Read(&line))
	require.Equal(t, "G28", saved.Command.Name)
	require.Equal(t, "X", saved.Command.Args[0].Key)
	require.InDelta(t, 10.0, saved.Command.Args[0].Number, 1e-9)
}

func TestCommandArg(t *testing.T) {
	t.Parallel()
	c := gcode.Command{
		Name: "G1",
		Args: []gcode.Argument{
			{Key: "X", Raw: "10", Number: 10, IsNumeric: true},
			{Key: "Y", Raw: "20", Number: 20, IsNumeric: true},
		},
	}
	a, ok := c.Arg("Y")
	require.True(t, ok)
	require.InDelta(t, 20.0, a.Number, 1e-9)

	_, ok = c.Arg("Z")
	require.False(t, ok)
}
