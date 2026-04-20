package gcode_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestParseError(t *testing.T) {
	t.Parallel()
	err := gcode.MakeParseErrorForTest(3, 5, "G99X", fmt.Errorf("unknown command"))

	t.Run("Error format", func(t *testing.T) {
		require.Equal(t, `gcode: parse error at line 3 col 5: unknown command (near "G99X")`, err.Error())
	})

	t.Run("Is ErrParse", func(t *testing.T) {
		require.True(t, errors.Is(err, gcode.ErrParse))
	})

	t.Run("Unwrap", func(t *testing.T) {
		inner := fmt.Errorf("bad value")
		err2 := gcode.MakeParseErrorForTest(1, 1, "X", inner)
		require.ErrorIs(t, err2, inner)
	})

	t.Run("As ParseErrorDetail", func(t *testing.T) {
		var detail gcode.ParseErrorDetail
		require.True(t, errors.As(err, &detail))
		require.Equal(t, 3, detail.Line())
		require.Equal(t, 5, detail.Column())
		require.Equal(t, "G99X", detail.Text())
	})
}
