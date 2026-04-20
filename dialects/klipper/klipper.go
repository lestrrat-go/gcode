// Package klipper provides a G-code dialect for Klipper firmware.
//
// Klipper accepts standard G/M/T codes (inherited here from the
// Marlin dialect) plus a large family of "extended" commands invoked
// by bare identifier and named arguments — for example:
//
//	SET_FAN_SPEED FAN=cooling SPEED=0.5
//	EXCLUDE_OBJECT_DEFINE NAME=part_0 CENTER=120,120 POLYGON=[[...]]
//
// Slicer output for Klipper printers (notably OrcaSlicer) routinely
// mixes both forms in the same file. The [gcode.Reader] handles the
// tokenisation; this dialect adds the well-known extended command
// names so they pass strict-mode validation.
package klipper

import (
	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

var dialect = build()

// Dialect returns the shared Klipper dialect singleton. The dialect
// inherits all Marlin commands and adds Klipper-specific extended
// commands. Mutating it affects every caller; use
// [gcode.Dialect.Extend] for a private child.
func Dialect() *gcode.Dialect { return dialect }

func build() *gcode.Dialect {
	d := marlin.Dialect().Extend("klipper")

	// Object exclusion — emitted per-print by OrcaSlicer and PrusaSlicer.
	d.Register(gcode.CommandDef{
		Name:        "EXCLUDE_OBJECT_DEFINE",
		Description: "declare a printable object",
		Params: []gcode.ParamDef{
			{Key: "NAME", Required: true},
			{Key: "CENTER"},
			{Key: "POLYGON"},
		},
	})
	d.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT_START", Description: "begin printing an object", Params: []gcode.ParamDef{{Key: "NAME", Required: true}}})
	d.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT_END", Description: "end printing an object", Params: []gcode.ParamDef{{Key: "NAME"}}})
	d.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT", Description: "toggle object exclusion", Params: []gcode.ParamDef{{Key: "NAME"}, {Key: "CURRENT"}, {Key: "RESET"}}})

	// Print state and progress.
	d.Register(gcode.CommandDef{
		Name:        "SET_PRINT_STATS_INFO",
		Description: "report layer progress",
		Params: []gcode.ParamDef{
			{Key: "TOTAL_LAYER"},
			{Key: "CURRENT_LAYER"},
		},
	})

	// Fan, pressure advance, velocity limits.
	d.Register(gcode.CommandDef{
		Name:        "SET_FAN_SPEED",
		Description: "set named fan speed",
		Params: []gcode.ParamDef{
			{Key: "FAN", Required: true},
			{Key: "SPEED"},
		},
	})
	d.Register(gcode.CommandDef{
		Name:        "SET_PRESSURE_ADVANCE",
		Description: "set extruder pressure advance",
		Params: []gcode.ParamDef{
			{Key: "ADVANCE"},
			{Key: "SMOOTH_TIME"},
			{Key: "EXTRUDER"},
		},
	})
	d.Register(gcode.CommandDef{
		Name:        "SET_VELOCITY_LIMIT",
		Description: "set kinematic limits",
		Params: []gcode.ParamDef{
			{Key: "VELOCITY"},
			{Key: "ACCEL"},
			{Key: "ACCEL_TO_DECEL"},
			{Key: "SQUARE_CORNER_VELOCITY"},
		},
	})

	// Bed mesh.
	d.Register(gcode.CommandDef{Name: "BED_MESH_CALIBRATE", Description: "probe and store bed mesh"})
	d.Register(gcode.CommandDef{
		Name:        "BED_MESH_PROFILE",
		Description: "load/save/remove a bed mesh profile",
		Params: []gcode.ParamDef{
			{Key: "LOAD"}, {Key: "SAVE"}, {Key: "REMOVE"},
		},
	})
	d.Register(gcode.CommandDef{Name: "BED_MESH_CLEAR", Description: "clear active bed mesh"})

	// State save/restore.
	d.Register(gcode.CommandDef{Name: "SAVE_GCODE_STATE", Description: "snapshot tool state", Params: []gcode.ParamDef{{Key: "NAME"}}})
	d.Register(gcode.CommandDef{Name: "RESTORE_GCODE_STATE", Description: "restore tool state", Params: []gcode.ParamDef{{Key: "NAME"}, {Key: "MOVE"}, {Key: "MOVE_SPEED"}}})

	// Timelapse plugin (very common in Klipper installs).
	d.Register(gcode.CommandDef{Name: "TIMELAPSE_TAKE_FRAME", Description: "trigger timelapse frame"})

	return d
}
