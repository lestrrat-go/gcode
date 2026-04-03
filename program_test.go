package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestLineIsBlank(t *testing.T) {
	t.Parallel()

	t.Run("blank line", func(t *testing.T) {
		var l gcode.Line
		require.True(t, l.IsBlank())
	})

	t.Run("has command", func(t *testing.T) {
		l := gcode.Line{HasCommand: true}
		require.False(t, l.IsBlank())
	})

	t.Run("has comment", func(t *testing.T) {
		l := gcode.Line{HasComment: true}
		require.False(t, l.IsBlank())
	})

	t.Run("has line number", func(t *testing.T) {
		l := gcode.Line{LineNumber: 10}
		require.False(t, l.IsBlank())
	})
}

func TestProgramBuilder(t *testing.T) {
	t.Parallel()

	t.Run("empty program", func(t *testing.T) {
		p := gcode.NewProgramBuilder().Build()
		require.Equal(t, 0, p.Len())
		require.Empty(t, p.Lines())
	})

	t.Run("append and build", func(t *testing.T) {
		line1 := gcode.Line{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 28,
				Params: []gcode.Parameter{
					{Letter: 'X', Value: 0},
					{Letter: 'Y', Value: 0},
				},
			},
		}
		line2 := gcode.Line{
			HasComment: true,
			Comment:    gcode.Comment{Text: "home axes", Form: gcode.CommentSemicolon},
		}

		p := gcode.NewProgramBuilder().
			Append(line1).
			Append(line2).
			Build()

		require.Equal(t, 2, p.Len())
		require.Equal(t, line1.Command.Letter, p.Line(0).Command.Letter)
		require.Equal(t, line1.Command.Number, p.Line(0).Command.Number)
		require.Equal(t, "home axes", p.Line(1).Comment.Text)
	})

	t.Run("params are copied", func(t *testing.T) {
		params := []gcode.Parameter{
			{Letter: 'X', Value: 10},
		}
		line := gcode.Line{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 1,
				Params: params,
			},
		}

		p := gcode.NewProgramBuilder().Append(line).Build()

		// Mutate original params slice — should not affect program.
		params[0].Value = 999

		require.Equal(t, float64(10), p.Line(0).Command.Params[0].Value)
	})

	t.Run("Lines returns a copy", func(t *testing.T) {
		line := gcode.Line{HasComment: true, Comment: gcode.Comment{Text: "test"}}
		p := gcode.NewProgramBuilder().Append(line).Build()

		lines := p.Lines()
		lines[0].Comment.Text = "mutated"

		require.Equal(t, "test", p.Line(0).Comment.Text)
	})
}
