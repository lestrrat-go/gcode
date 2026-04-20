package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestCommentFormString(t *testing.T) {
	t.Parallel()
	require.Equal(t, "semicolon", gcode.CommentSemicolon.String())
	require.Equal(t, "parenthesis", gcode.CommentParenthesis.String())
	require.Equal(t, "CommentForm(99)", gcode.CommentForm(99).String())
}
