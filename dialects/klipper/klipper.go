// Package klipper provides a G-code dialect for Klipper firmware.
//
// Klipper accepts standard G/M/T codes (inherited here from the
// Marlin dialect) plus a large family of "extended" commands invoked
// by bare identifier and named arguments — for example:
//
//	SET_FAN_SPEED FAN=cooling SPEED=0.5
//	EXCLUDE_OBJECT_DEFINE NAME=part_0 CENTER=120,120 POLYGON=[[...]]
//
// [Dialect] returns the always-available subset: classic Marlin G/M
// codes plus the Klipper extended commands that are part of Klipper
// core and registered unconditionally (pressure advance, velocity
// limits, gcode-state save/restore, print stats).
//
// Many useful Klipper commands are conditional on configuration or
// supplied by third-party plugins. They live behind opt-in helpers:
//
//   - [WithBedMesh]       — requires [bed_mesh] in printer.cfg
//   - [WithExcludeObject] — requires [exclude_object] in printer.cfg
//   - [WithFanGeneric]    — requires at least one [fan_generic]
//   - [WithTimelapse]     — requires the moonraker-timelapse plugin
//
// Each helper clones the input dialect (the package singleton would
// otherwise mutate globally) and returns a new dialect with the
// feature commands registered. Compose them by chaining:
//
//	d := klipper.WithExcludeObject(klipper.WithBedMesh(klipper.Dialect()))
package klipper

import (
	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

var dialect = build()

// Dialect returns the shared Klipper dialect singleton, covering only
// the Klipper extended commands that are unconditionally available in
// a stock Klipper install. Use the With* helpers to layer on
// configuration-dependent or plugin-supplied commands. Mutating the
// singleton affects every caller; use [gcode.Dialect.Extend] (or a
// helper) for a private child.
func Dialect() *gcode.Dialect { return dialect }

func build() *gcode.Dialect {
	d := marlin.Dialect().Extend("klipper")

	d.Register(gcode.CommandDef{Name: "SET_PRESSURE_ADVANCE", Description: "set extruder pressure advance", Params: []gcode.ParamDef{{Key: "ADVANCE"}, {Key: "SMOOTH_TIME"}, {Key: "EXTRUDER"}}})
	d.Register(gcode.CommandDef{Name: "SET_VELOCITY_LIMIT", Description: "set kinematic limits", Params: []gcode.ParamDef{{Key: "VELOCITY"}, {Key: "ACCEL"}, {Key: "ACCEL_TO_DECEL"}, {Key: "SQUARE_CORNER_VELOCITY"}}})
	d.Register(gcode.CommandDef{Name: "SAVE_GCODE_STATE", Description: "snapshot tool state", Params: []gcode.ParamDef{{Key: "NAME"}}})
	d.Register(gcode.CommandDef{Name: "RESTORE_GCODE_STATE", Description: "restore tool state", Params: []gcode.ParamDef{{Key: "NAME"}, {Key: "MOVE"}, {Key: "MOVE_SPEED"}}})
	d.Register(gcode.CommandDef{Name: "SET_PRINT_STATS_INFO", Description: "report layer progress", Params: []gcode.ParamDef{{Key: "TOTAL_LAYER"}, {Key: "CURRENT_LAYER"}}})

	return d
}

// WithBedMesh returns a new dialect equal to d plus the Klipper
// [bed_mesh] commands (BED_MESH_CALIBRATE, BED_MESH_PROFILE,
// BED_MESH_CLEAR). d is not modified.
func WithBedMesh(d *gcode.Dialect) *gcode.Dialect {
	out := d.Extend(d.Name())
	out.Register(gcode.CommandDef{Name: "BED_MESH_CALIBRATE", Description: "probe and store bed mesh"})
	out.Register(gcode.CommandDef{Name: "BED_MESH_PROFILE", Description: "load/save/remove a bed mesh profile", Params: []gcode.ParamDef{{Key: "LOAD"}, {Key: "SAVE"}, {Key: "REMOVE"}}})
	out.Register(gcode.CommandDef{Name: "BED_MESH_CLEAR", Description: "clear active bed mesh"})
	return out
}

// WithExcludeObject returns a new dialect equal to d plus the Klipper
// [exclude_object] commands (EXCLUDE_OBJECT_DEFINE, EXCLUDE_OBJECT_START,
// EXCLUDE_OBJECT_END, EXCLUDE_OBJECT). d is not modified.
func WithExcludeObject(d *gcode.Dialect) *gcode.Dialect {
	out := d.Extend(d.Name())
	out.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT_DEFINE", Description: "declare a printable object", Params: []gcode.ParamDef{{Key: "NAME", Required: true}, {Key: "CENTER"}, {Key: "POLYGON"}}})
	out.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT_START", Description: "begin printing an object", Params: []gcode.ParamDef{{Key: "NAME", Required: true}}})
	out.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT_END", Description: "end printing an object", Params: []gcode.ParamDef{{Key: "NAME"}}})
	out.Register(gcode.CommandDef{Name: "EXCLUDE_OBJECT", Description: "toggle object exclusion", Params: []gcode.ParamDef{{Key: "NAME"}, {Key: "CURRENT"}, {Key: "RESET"}}})
	return out
}

// WithFanGeneric returns a new dialect equal to d plus the SET_FAN_SPEED
// command supplied by Klipper's [fan_generic] config sections.
// d is not modified.
func WithFanGeneric(d *gcode.Dialect) *gcode.Dialect {
	out := d.Extend(d.Name())
	out.Register(gcode.CommandDef{Name: "SET_FAN_SPEED", Description: "set named fan speed", Params: []gcode.ParamDef{{Key: "FAN", Required: true}, {Key: "SPEED"}}})
	return out
}

// WithTimelapse returns a new dialect equal to d plus the
// TIMELAPSE_TAKE_FRAME command supplied by the moonraker-timelapse
// plugin. d is not modified.
func WithTimelapse(d *gcode.Dialect) *gcode.Dialect {
	out := d.Extend(d.Name())
	out.Register(gcode.CommandDef{Name: "TIMELAPSE_TAKE_FRAME", Description: "trigger timelapse frame"})
	return out
}
