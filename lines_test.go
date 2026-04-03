package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestBlankLine(t *testing.T) {
	t.Parallel()

	l := gcode.BlankLine()
	require.True(t, l.IsBlank())
}

func TestCommentLine(t *testing.T) {
	t.Parallel()

	l := gcode.CommentLine("test")
	require.True(t, l.HasComment)
	require.False(t, l.HasCommand)
	require.Equal(t, gcode.CommentSemicolon, l.Comment.Form)
	require.Equal(t, "test", l.Comment.Text)
}

func TestParenCommentLine(t *testing.T) {
	t.Parallel()

	l := gcode.ParenCommentLine("test")
	require.True(t, l.HasComment)
	require.False(t, l.HasCommand)
	require.Equal(t, gcode.CommentParenthesis, l.Comment.Form)
	require.Equal(t, "test", l.Comment.Text)
}
