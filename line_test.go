package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestLineIsBlank(t *testing.T) {
	t.Parallel()
	require.True(t, gcode.Line{}.IsBlank())
	require.False(t, gcode.Line{LineNumber: 1}.IsBlank())
	require.False(t, gcode.Line{HasCommand: true, Command: gcode.Command{Name: "G28"}}.IsBlank())
	require.False(t, gcode.Line{HasComment: true}.IsBlank())
	require.False(t, gcode.Line{HasChecksum: true}.IsBlank())
}

func TestLineClone(t *testing.T) {
	t.Parallel()
	original := gcode.Line{
		LineNumber: 5,
		HasCommand: true,
		Command: gcode.Command{
			Name: "G1",
			Args: []gcode.Argument{
				{Key: "X", Raw: "10", Number: 10, IsNumeric: true},
				{Key: "Y", Raw: "20", Number: 20, IsNumeric: true},
			},
		},
		HasComment: true,
		Comment:    gcode.Comment{Text: " hello", Form: gcode.CommentSemicolon},
	}
	cloned := original.Clone()

	cloned.Command.Args[0].Number = 999
	cloned.Command.Args[0].Raw = "999"

	require.InDelta(t, 10.0, original.Command.Args[0].Number, 1e-9)
	require.Equal(t, "10", original.Command.Args[0].Raw)
}

func TestArgumentIsFlag(t *testing.T) {
	t.Parallel()
	require.True(t, gcode.Argument{Key: "X"}.IsFlag())
	require.False(t, gcode.Argument{Key: "X", Raw: "10"}.IsFlag())
}

func TestNewLineFluent(t *testing.T) {
	t.Parallel()
	l := gcode.NewLine("G1").
		LineNo(7).
		Arg("X", "10.5").
		ArgF("Y", 20).
		Flag("Z").
		WithComment(" trailing").
		WithChecksum(42)

	require.True(t, l.HasCommand)
	require.Equal(t, "G1", l.Command.Name)
	require.Equal(t, 7, l.LineNumber)
	require.Len(t, l.Command.Args, 3)

	require.Equal(t, "X", l.Command.Args[0].Key)
	require.Equal(t, "10.5", l.Command.Args[0].Raw)
	require.True(t, l.Command.Args[0].IsNumeric)
	require.InDelta(t, 10.5, l.Command.Args[0].Number, 1e-9)

	require.Equal(t, "Y", l.Command.Args[1].Key)
	require.Equal(t, "20", l.Command.Args[1].Raw)
	require.InDelta(t, 20.0, l.Command.Args[1].Number, 1e-9)

	require.Equal(t, "Z", l.Command.Args[2].Key)
	require.True(t, l.Command.Args[2].IsFlag())

	require.True(t, l.HasComment)
	require.Equal(t, " trailing", l.Comment.Text)
	require.Equal(t, gcode.CommentSemicolon, l.Comment.Form)

	require.True(t, l.HasChecksum)
	require.Equal(t, byte(42), l.Checksum)
}

func TestNewLineFluentIsCopyOnModify(t *testing.T) {
	t.Parallel()
	base := gcode.NewLine("G1").Arg("X", "10")
	a := base.Arg("Y", "20")
	b := base.Arg("Y", "30")

	require.Len(t, base.Command.Args, 1)
	require.Len(t, a.Command.Args, 2)
	require.Len(t, b.Command.Args, 2)
	require.Equal(t, "20", a.Command.Args[1].Raw)
	require.Equal(t, "30", b.Command.Args[1].Raw)
}

func TestNewComment(t *testing.T) {
	t.Parallel()
	l := gcode.NewComment(" header")
	require.False(t, l.HasCommand)
	require.True(t, l.HasComment)
	require.Equal(t, gcode.CommentSemicolon, l.Comment.Form)
	require.Equal(t, " header", l.Comment.Text)

	p := gcode.NewParenComment("setup")
	require.Equal(t, gcode.CommentParenthesis, p.Comment.Form)
	require.Equal(t, "setup", p.Comment.Text)
}

func TestLineFluentNoOpWithoutCommand(t *testing.T) {
	t.Parallel()
	l := gcode.NewComment(" hi").Arg("X", "10").Flag("Y")
	require.False(t, l.HasCommand)
	require.Empty(t, l.Command.Args)
}

func TestLineArgFP(t *testing.T) {
	t.Parallel()

	// 0.1 + 0.2 is the canonical float64-noise case, but Go's
	// constant-folding evaluates literal `0.1 + 0.2` exactly to 0.3
	// at compile time. Forcing the addition through variables makes
	// the compiler emit a runtime float64 add, which produces the
	// expected bit pattern (0x3FD3333333333334) whose shortest
	// representation is "0.30000000000000004".
	a, b := 0.1, 0.2
	noisy := a + b

	t.Run("ArgF leaks shortest-representation noise", func(t *testing.T) {
		l := gcode.NewLine("G1").ArgF("E", noisy)
		require.Equal(t, "0.30000000000000004", l.Command.Args[0].Raw)
	})

	t.Run("ArgFP formats to fixed decimals", func(t *testing.T) {
		l := gcode.NewLine("G1").
			ArgFP("X", 3, 10.1).
			ArgFP("Y", 3, 20.5).
			ArgFP("E", 5, noisy).
			ArgFP("F", 0, 1500)

		require.Equal(t, "10.100", l.Command.Args[0].Raw)
		require.Equal(t, "20.500", l.Command.Args[1].Raw)
		require.Equal(t, "0.30000", l.Command.Args[2].Raw)
		require.Equal(t, "1500", l.Command.Args[3].Raw)

		// Number keeps full float64 precision regardless of Raw format.
		require.InDelta(t, noisy, l.Command.Args[2].Number, 1e-15)
		require.True(t, l.Command.Args[2].IsNumeric)
	})

	t.Run("ArgFP forces decimal places even for whole values", func(t *testing.T) {
		l := gcode.NewLine("G1").ArgFP("X", 3, 10).ArgFP("F", 0, 1500.7)
		require.Equal(t, "10.000", l.Command.Args[0].Raw)
		require.Equal(t, "1501", l.Command.Args[1].Raw)
	})

	t.Run("negative precision clamps to 0", func(t *testing.T) {
		l := gcode.NewLine("G1").ArgFP("E", -5, 1.789)
		require.Equal(t, "2", l.Command.Args[0].Raw)
	})

	t.Run("no-op without a command", func(t *testing.T) {
		l := gcode.NewComment(" hi").ArgFP("X", 3, 1.5)
		require.False(t, l.HasCommand)
		require.Empty(t, l.Command.Args)
	})
}
