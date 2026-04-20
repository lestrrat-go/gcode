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
