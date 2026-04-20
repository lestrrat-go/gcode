package gcode_test

import (
	"fmt"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestSimpleMacroExpand(t *testing.T) {
	t.Parallel()
	lines := []gcode.Line{
		{HasCommand: true, Command: gcode.Command{Name: "G28"}},
		{HasCommand: true, Command: gcode.Command{
			Name: "G1",
			Args: []gcode.Argument{
				{Key: "X", Raw: "10", Number: 10, IsNumeric: true},
				{Key: "Y", Raw: "20", Number: 20, IsNumeric: true},
			},
		}},
	}

	m := gcode.NewSimpleMacro("home-and-move", lines)
	require.Equal(t, "home-and-move", m.Name())

	expanded, err := m.Expand(nil)
	require.NoError(t, err)
	require.Len(t, expanded, 2)
	require.Equal(t, "G28", expanded[0].Command.Name)
	require.Equal(t, "G1", expanded[1].Command.Name)
	require.Len(t, expanded[1].Command.Args, 2)
}

func TestSimpleMacroExpandIsDeepCopy(t *testing.T) {
	t.Parallel()
	lines := []gcode.Line{
		{HasCommand: true, Command: gcode.Command{
			Name: "G1",
			Args: []gcode.Argument{{Key: "X", Raw: "10", Number: 10, IsNumeric: true}},
		}},
	}
	m := gcode.NewSimpleMacro("move", lines)

	first, err := m.Expand(nil)
	require.NoError(t, err)
	first[0].Command.Args[0].Number = 999

	second, err := m.Expand(nil)
	require.NoError(t, err)
	require.InDelta(t, 10.0, second[0].Command.Args[0].Number, 1e-9)
}

func TestMacroRegistryLookupExpand(t *testing.T) {
	t.Parallel()
	reg := gcode.NewMacroRegistry()
	reg.Register(gcode.NewSimpleMacro("preheat", []gcode.Line{
		{HasCommand: true, Command: gcode.Command{
			Name: "M104",
			Args: []gcode.Argument{{Key: "S", Raw: "200", Number: 200, IsNumeric: true}},
		}},
	}))

	m, ok := reg.Lookup("preheat")
	require.True(t, ok)
	require.Equal(t, "preheat", m.Name())

	expanded, err := reg.Expand("preheat", nil)
	require.NoError(t, err)
	require.Len(t, expanded, 1)
	require.Equal(t, "M104", expanded[0].Command.Name)
}

func TestMacroRegistryUnknown(t *testing.T) {
	t.Parallel()
	reg := gcode.NewMacroRegistry()
	_, err := reg.Expand("nope", nil)
	require.Error(t, err)
}

type customMacro struct{ name string }

func (m *customMacro) Name() string { return m.name }
func (m *customMacro) Expand(args map[string]float64) ([]gcode.Line, error) {
	x, ok := args["X"]
	if !ok {
		return nil, fmt.Errorf("missing required arg X")
	}
	return []gcode.Line{
		{HasCommand: true, Command: gcode.Command{
			Name: "G1",
			Args: []gcode.Argument{{Key: "X", Raw: fmt.Sprintf("%g", x), Number: x, IsNumeric: true}},
		}},
	}, nil
}

func TestCustomMacro(t *testing.T) {
	t.Parallel()
	reg := gcode.NewMacroRegistry()
	reg.Register(&customMacro{name: "move-x"})

	expanded, err := reg.Expand("move-x", map[string]float64{"X": 42})
	require.NoError(t, err)
	require.InDelta(t, 42.0, expanded[0].Command.Args[0].Number, 1e-9)

	_, err = reg.Expand("move-x", nil)
	require.Error(t, err)
}
