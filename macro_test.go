package gcode_test

import (
	"fmt"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

func TestSimpleMacroExpand(t *testing.T) {
	lines := []gcode.Line{
		{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 28,
			},
		},
		{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 1,
				Params: []gcode.Parameter{
					{Letter: 'X', Value: 10},
					{Letter: 'Y', Value: 20},
				},
			},
		},
	}

	m := gcode.NewSimpleMacro("home-and-move", lines)
	require.Equal(t, "home-and-move", m.Name())

	expanded, err := m.Expand(nil)
	require.NoError(t, err)
	require.Len(t, expanded, 2)

	require.Equal(t, byte('G'), expanded[0].Command.Letter)
	require.Equal(t, 28, expanded[0].Command.Number)

	require.Equal(t, byte('G'), expanded[1].Command.Letter)
	require.Equal(t, 1, expanded[1].Command.Number)
	require.Len(t, expanded[1].Command.Params, 2)
	require.Equal(t, byte('X'), expanded[1].Command.Params[0].Letter)
	require.InDelta(t, 10.0, expanded[1].Command.Params[0].Value, 0.001)
	require.Equal(t, byte('Y'), expanded[1].Command.Params[1].Letter)
	require.InDelta(t, 20.0, expanded[1].Command.Params[1].Value, 0.001)
}

func TestSimpleMacroIgnoresArgs(t *testing.T) {
	lines := []gcode.Line{
		{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 28,
			},
		},
	}

	m := gcode.NewSimpleMacro("home", lines)

	args := map[string]float64{"X": 100, "Y": 200}
	expanded, err := m.Expand(args)
	require.NoError(t, err)
	require.Len(t, expanded, 1)
	require.Equal(t, 28, expanded[0].Command.Number)
}

func TestSimpleMacroExpandDeepCopy(t *testing.T) {
	lines := []gcode.Line{
		{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 1,
				Params: []gcode.Parameter{
					{Letter: 'X', Value: 10},
				},
			},
		},
	}

	m := gcode.NewSimpleMacro("move", lines)

	expanded, err := m.Expand(nil)
	require.NoError(t, err)

	// Mutate the expanded result.
	expanded[0].Command.Params[0].Value = 999

	// Expand again and verify the original is unchanged.
	expanded2, err := m.Expand(nil)
	require.NoError(t, err)
	require.InDelta(t, 10.0, expanded2[0].Command.Params[0].Value, 0.001)
}

func TestMacroRegistryExpandUnknown(t *testing.T) {
	reg := gcode.NewMacroRegistry()

	_, err := reg.Expand("nonexistent", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent")
}

func TestMacroRegistryRegisterAndExpand(t *testing.T) {
	lines := []gcode.Line{
		{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'M',
				Number: 104,
				Params: []gcode.Parameter{
					{Letter: 'S', Value: 200},
				},
			},
		},
	}

	reg := gcode.NewMacroRegistry()
	reg.Register(gcode.NewSimpleMacro("preheat", lines))

	m, ok := reg.Lookup("preheat")
	require.True(t, ok)
	require.Equal(t, "preheat", m.Name())

	expanded, err := reg.Expand("preheat", nil)
	require.NoError(t, err)
	require.Len(t, expanded, 1)
	require.Equal(t, byte('M'), expanded[0].Command.Letter)
	require.Equal(t, 104, expanded[0].Command.Number)
}

func TestMacroRegistryLookupMissing(t *testing.T) {
	reg := gcode.NewMacroRegistry()

	_, ok := reg.Lookup("missing")
	require.False(t, ok)
}

// customMacro is a test-only Macro implementation that builds lines
// dynamically from the args map.
type customMacro struct {
	name string
}

func (m *customMacro) Name() string { return m.name }

func (m *customMacro) Expand(args map[string]float64) ([]gcode.Line, error) {
	xVal, ok := args["X"]
	if !ok {
		return nil, fmt.Errorf("missing required arg X")
	}
	return []gcode.Line{
		{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'G',
				Number: 1,
				Params: []gcode.Parameter{
					{Letter: 'X', Value: xVal},
				},
			},
		},
	}, nil
}

func TestMacroRegistryCustomMacro(t *testing.T) {
	reg := gcode.NewMacroRegistry()
	reg.Register(&customMacro{name: "move-x"})

	expanded, err := reg.Expand("move-x", map[string]float64{"X": 42})
	require.NoError(t, err)
	require.Len(t, expanded, 1)
	require.InDelta(t, 42.0, expanded[0].Command.Params[0].Value, 0.001)
}

func TestMacroRegistryCustomMacroError(t *testing.T) {
	reg := gcode.NewMacroRegistry()
	reg.Register(&customMacro{name: "move-x"})

	_, err := reg.Expand("move-x", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required arg X")
}
