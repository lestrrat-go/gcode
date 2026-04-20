package gcode_test

import (
	"strings"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/klipper"
	"github.com/lestrrat-go/gcode/dialects/marlin"
	"github.com/lestrrat-go/gcode/dialects/reprap"
	"github.com/stretchr/testify/require"
)

func TestNewDialect(t *testing.T) {
	t.Parallel()
	d := gcode.NewDialect("test")
	require.Equal(t, "test", d.Name())
}

func TestDialectRegisterAndLookup(t *testing.T) {
	t.Parallel()
	d := gcode.NewDialect("test")
	d.Register(gcode.CommandDef{
		Name:        "G0",
		Description: "rapid",
		Params:      []gcode.ParamDef{{Key: "X"}, {Key: "Y"}},
	})

	def, ok := d.LookupCommand("G0")
	require.True(t, ok)
	require.Equal(t, "rapid", def.Description)
	require.Len(t, def.Params, 2)

	_, ok = d.LookupCommand("G99")
	require.False(t, ok)
}

func TestDialectExtendIndependence(t *testing.T) {
	t.Parallel()
	parent := gcode.NewDialect("parent")
	parent.Register(gcode.CommandDef{Name: "G0"})

	child := parent.Extend("child")
	require.Equal(t, "child", child.Name())

	_, ok := child.LookupCommand("G0")
	require.True(t, ok)

	child.Register(gcode.CommandDef{Name: "G1"})
	_, ok = parent.LookupCommand("G1")
	require.False(t, ok)
}

func TestDialectCommands(t *testing.T) {
	t.Parallel()
	d := gcode.NewDialect("test")
	d.Register(gcode.CommandDef{Name: "G0"})
	d.Register(gcode.CommandDef{Name: "G1"})
	d.Register(gcode.CommandDef{Name: "M104"})

	require.Len(t, d.Commands(), 3)
}

func TestMarlinDialect(t *testing.T) {
	t.Parallel()
	d := marlin.Dialect()
	require.Equal(t, "marlin", d.Name())

	for _, name := range []string{"G0", "G1", "G28", "G92.1", "M104", "T0"} {
		_, ok := d.LookupCommand(name)
		require.True(t, ok, "expected dialect to define %s", name)
	}
}

func TestRepRapDialect(t *testing.T) {
	t.Parallel()
	d := reprap.Dialect()
	require.Equal(t, "reprap", d.Name())

	// Inherited.
	_, ok := d.LookupCommand("G0")
	require.True(t, ok)
	// RepRap-specific.
	_, ok = d.LookupCommand("G10")
	require.True(t, ok)
	_, ok = d.LookupCommand("M557")
	require.True(t, ok)
}

func TestKlipperDialectCore(t *testing.T) {
	t.Parallel()
	d := klipper.Dialect()
	require.Equal(t, "klipper", d.Name())

	// Always-available core extended commands.
	for _, name := range []string{
		"G1", "M104", // inherited Marlin
		"SET_PRESSURE_ADVANCE",
		"SET_VELOCITY_LIMIT",
		"SAVE_GCODE_STATE",
		"RESTORE_GCODE_STATE",
		"SET_PRINT_STATS_INFO",
	} {
		_, ok := d.LookupCommand(name)
		require.True(t, ok, "expected klipper core dialect to define %s", name)
	}

	// Optional features must NOT be in the core singleton.
	for _, name := range []string{
		"BED_MESH_CALIBRATE",
		"EXCLUDE_OBJECT_DEFINE",
		"SET_FAN_SPEED",
		"TIMELAPSE_TAKE_FRAME",
	} {
		_, ok := d.LookupCommand(name)
		require.False(t, ok, "expected %s to be opt-in via With...", name)
	}
}

func TestKlipperWithHelpersClone(t *testing.T) {
	t.Parallel()

	base := klipper.Dialect()
	withMesh := klipper.WithBedMesh(base)
	withMeshAndExclude := klipper.WithExcludeObject(withMesh)

	// Helpers do not mutate the inputs.
	_, ok := base.LookupCommand("BED_MESH_CALIBRATE")
	require.False(t, ok, "WithBedMesh must not mutate input")
	_, ok = withMesh.LookupCommand("EXCLUDE_OBJECT_DEFINE")
	require.False(t, ok, "WithExcludeObject must not mutate input")

	// Returned dialects accumulate features.
	_, ok = withMesh.LookupCommand("BED_MESH_CALIBRATE")
	require.True(t, ok)
	_, ok = withMeshAndExclude.LookupCommand("BED_MESH_CALIBRATE")
	require.True(t, ok)
	_, ok = withMeshAndExclude.LookupCommand("EXCLUDE_OBJECT_DEFINE")
	require.True(t, ok)
}

func TestKlipperParsesExtendedWithOptIn(t *testing.T) {
	t.Parallel()
	src := `EXCLUDE_OBJECT_DEFINE NAME=part_0 CENTER=135.5,136 POLYGON=[[1,2],[3,4]]
EXCLUDE_OBJECT_START NAME=part_0
G1 X10 Y20
EXCLUDE_OBJECT_END NAME=part_0
`
	d := klipper.WithExcludeObject(klipper.Dialect())
	r := gcode.NewReader(
		strings.NewReader(src),
		gcode.WithDialect(d),
		gcode.WithStrict(),
	)
	count := 0
	for line, err := range r.All() {
		require.NoError(t, err)
		require.True(t, line.HasCommand)
		count++
	}
	require.Equal(t, 4, count)
}

func TestWithDialect(t *testing.T) {
	t.Parallel()
	d := gcode.NewDialect("test")
	require.NotNil(t, gcode.WithDialect(d))
}

func TestDialectRegistry(t *testing.T) {
	t.Parallel()
	r := gcode.NewDialectRegistry()
	d := gcode.NewDialect("test")
	r.Register(d)

	found, ok := r.Lookup("test")
	require.True(t, ok)
	require.Equal(t, "test", found.Name())

	_, ok = r.Lookup("nonexistent")
	require.False(t, ok)
}
