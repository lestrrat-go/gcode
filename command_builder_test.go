package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestCommandBuilder(t *testing.T) {
	t.Parallel()

	t.Run("G basic", func(t *testing.T) {
		l := gcode.G(0).Line()
		require.True(t, l.HasCommand)
		require.Equal(t, byte('G'), l.Command.Letter)
		require.Equal(t, 0, l.Command.Number)
		require.False(t, l.HasComment)
	})

	t.Run("M with param", func(t *testing.T) {
		l := gcode.M(104).S(200).Line()
		require.True(t, l.HasCommand)
		require.Equal(t, byte('M'), l.Command.Letter)
		require.Equal(t, 104, l.Command.Number)
		require.Len(t, l.Command.Params, 1)
		require.Equal(t, byte('S'), l.Command.Params[0].Letter)
		require.Equal(t, float64(200), l.Command.Params[0].Value)
	})

	t.Run("T tool select", func(t *testing.T) {
		l := gcode.T(0).Line()
		require.True(t, l.HasCommand)
		require.Equal(t, byte('T'), l.Command.Letter)
		require.Equal(t, 0, l.Command.Number)
	})

	t.Run("Cmd generic", func(t *testing.T) {
		l := gcode.Cmd('G', 1).X(10).Line()
		require.True(t, l.HasCommand)
		require.Equal(t, byte('G'), l.Command.Letter)
		require.Equal(t, 1, l.Command.Number)
		require.Len(t, l.Command.Params, 1)
		require.Equal(t, byte('X'), l.Command.Params[0].Letter)
		require.Equal(t, float64(10), l.Command.Params[0].Value)
	})

	t.Run("GSub subcode", func(t *testing.T) {
		l := gcode.GSub(92, 1).Line()
		require.True(t, l.HasCommand)
		require.Equal(t, byte('G'), l.Command.Letter)
		require.Equal(t, 92, l.Command.Number)
		require.True(t, l.Command.HasSubcode)
		require.Equal(t, 1, l.Command.Subcode)
	})

	t.Run("all common params", func(t *testing.T) {
		l := gcode.G(1).X(10).Y(20).Z(0.5).E(0.04).F(1200).Line()
		require.True(t, l.HasCommand)
		require.Len(t, l.Command.Params, 5)

		expected := []struct {
			letter byte
			value  float64
		}{
			{'X', 10},
			{'Y', 20},
			{'Z', 0.5},
			{'E', 0.04},
			{'F', 1200},
		}
		for i, exp := range expected {
			require.Equal(t, exp.letter, l.Command.Params[i].Letter)
			require.Equal(t, exp.value, l.Command.Params[i].Value)
		}
	})

	t.Run("S and P params", func(t *testing.T) {
		l := gcode.G(4).S(1000).P(500).Line()
		require.Len(t, l.Command.Params, 2)
		require.Equal(t, byte('S'), l.Command.Params[0].Letter)
		require.Equal(t, float64(1000), l.Command.Params[0].Value)
		require.Equal(t, byte('P'), l.Command.Params[1].Letter)
		require.Equal(t, float64(500), l.Command.Params[1].Value)
	})

	t.Run("R param", func(t *testing.T) {
		l := gcode.G(1).R(5).Line()
		require.Len(t, l.Command.Params, 1)
		require.Equal(t, byte('R'), l.Command.Params[0].Letter)
		require.Equal(t, float64(5), l.Command.Params[0].Value)
	})

	t.Run("generic Param for arc", func(t *testing.T) {
		l := gcode.G(1).Param('I', 3.5).Param('J', 4.2).Line()
		require.Len(t, l.Command.Params, 2)
		require.Equal(t, byte('I'), l.Command.Params[0].Letter)
		require.Equal(t, 3.5, l.Command.Params[0].Value)
		require.Equal(t, byte('J'), l.Command.Params[1].Letter)
		require.Equal(t, 4.2, l.Command.Params[1].Value)
	})

	t.Run("semicolon trailing comment", func(t *testing.T) {
		l := gcode.G(28).Comment("home all").Line()
		require.True(t, l.HasCommand)
		require.True(t, l.HasComment)
		require.Equal(t, gcode.CommentSemicolon, l.Comment.Form)
		require.Equal(t, "home all", l.Comment.Text)
	})

	t.Run("paren trailing comment", func(t *testing.T) {
		l := gcode.G(1).X(10).ParenComment("move").Line()
		require.True(t, l.HasCommand)
		require.True(t, l.HasComment)
		require.Equal(t, gcode.CommentParenthesis, l.Comment.Form)
		require.Equal(t, "move", l.Comment.Text)
	})
}
