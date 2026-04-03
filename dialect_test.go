package gcode_test

import (
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
	"github.com/lestrrat-go/gcode/dialects/reprap"
	"github.com/stretchr/testify/require"
)

func TestNewDialect(t *testing.T) {
	d := gcode.NewDialect("test")
	require.Equal(t, "test", d.Name())
}

func TestDialectRegisterAndLookup(t *testing.T) {
	d := gcode.NewDialect("test")
	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      0,
		Description: "rapid move",
		Params: []gcode.ParamDef{
			{Letter: 'X'},
			{Letter: 'Y'},
		},
	})

	def, ok := d.LookupCommand('G', 0, 0)
	require.True(t, ok)
	require.Equal(t, byte('G'), def.Letter)
	require.Equal(t, 0, def.Number)
	require.Equal(t, "rapid move", def.Description)
	require.Len(t, def.Params, 2)
}

func TestDialectLookupUnknown(t *testing.T) {
	d := gcode.NewDialect("test")
	_, ok := d.LookupCommand('G', 99, 0)
	require.False(t, ok)
}

func TestDialectExtendIndependence(t *testing.T) {
	parent := gcode.NewDialect("parent")
	parent.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      0,
		Description: "rapid move",
	})

	child := parent.Extend("child")
	require.Equal(t, "child", child.Name())

	// Child inherits parent commands.
	_, ok := child.LookupCommand('G', 0, 0)
	require.True(t, ok)

	// Register on child does not affect parent.
	child.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      1,
		Description: "linear move",
	})
	_, ok = child.LookupCommand('G', 1, 0)
	require.True(t, ok)
	_, ok = parent.LookupCommand('G', 1, 0)
	require.False(t, ok)
}

func TestDialectCommands(t *testing.T) {
	d := gcode.NewDialect("test")
	d.Register(gcode.CommandDef{Letter: 'G', Number: 0, Description: "rapid"})
	d.Register(gcode.CommandDef{Letter: 'G', Number: 1, Description: "linear"})
	d.Register(gcode.CommandDef{Letter: 'M', Number: 104, Description: "hotend"})

	cmds := d.Commands()
	require.Len(t, cmds, 3)
}

func TestMarlinDialect(t *testing.T) {
	d := marlin.Dialect()
	require.NotNil(t, d)
	require.Equal(t, "marlin", d.Name())

	// Spot-check G0.
	def, ok := d.LookupCommand('G', 0, 0)
	require.True(t, ok)
	require.Equal(t, byte('G'), def.Letter)
	require.Equal(t, 0, def.Number)

	// Spot-check G28.
	def, ok = d.LookupCommand('G', 28, 0)
	require.True(t, ok)
	require.Equal(t, "auto home", def.Description)

	// Spot-check M104.
	def, ok = d.LookupCommand('M', 104, 0)
	require.True(t, ok)
	require.Equal(t, "set hotend temp", def.Description)

	// Spot-check G92.1 subcode.
	def, ok = d.LookupCommand('G', 92, 1)
	require.True(t, ok)
	require.True(t, def.HasSubcode)
	require.Equal(t, 1, def.Subcode)
}

func TestRepRapDialect(t *testing.T) {
	d := reprap.Dialect()
	require.NotNil(t, d)
	require.Equal(t, "reprap", d.Name())

	// Inherited Marlin commands.
	_, ok := d.LookupCommand('G', 0, 0)
	require.True(t, ok)
	_, ok = d.LookupCommand('M', 104, 0)
	require.True(t, ok)

	// RepRap-specific commands.
	def, ok := d.LookupCommand('G', 10, 0)
	require.True(t, ok)
	require.Equal(t, "retract/set tool offset", def.Description)

	_, ok = d.LookupCommand('M', 116, 0)
	require.True(t, ok)
}

func TestMarlinDialectIndependentInstances(t *testing.T) {
	d1 := marlin.Dialect()
	d2 := marlin.Dialect()

	// Mutating one should not affect the other.
	d1.Register(gcode.CommandDef{Letter: 'G', Number: 999, Description: "custom"})
	_, ok := d1.LookupCommand('G', 999, 0)
	require.True(t, ok)
	_, ok = d2.LookupCommand('G', 999, 0)
	require.False(t, ok)
}

func TestWithDialect(t *testing.T) {
	d := gcode.NewDialect("test")
	opt := gcode.WithDialect(d)
	require.NotNil(t, opt)
}

func TestDialectRegistry(t *testing.T) {
	r := gcode.NewDialectRegistry()
	d := gcode.NewDialect("test")
	r.Register(d)

	found, ok := r.Lookup("test")
	require.True(t, ok)
	require.Equal(t, "test", found.Name())

	_, ok = r.Lookup("nonexistent")
	require.False(t, ok)
}
