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
	return marlin.Dialect().Extend("klipper").
		Register(gcode.NewCommand("SET_PRESSURE_ADVANCE").Describe("set extruder pressure advance").Optional("ADVANCE", "SMOOTH_TIME", "EXTRUDER")).
		Register(gcode.NewCommand("SET_VELOCITY_LIMIT").Describe("set kinematic limits").Optional("VELOCITY", "ACCEL", "ACCEL_TO_DECEL", "SQUARE_CORNER_VELOCITY")).
		Register(gcode.NewCommand("SAVE_GCODE_STATE").Describe("snapshot tool state").Optional("NAME")).
		Register(gcode.NewCommand("RESTORE_GCODE_STATE").Describe("restore tool state").Optional("NAME", "MOVE", "MOVE_SPEED")).
		Register(gcode.NewCommand("SET_PRINT_STATS_INFO").Describe("report layer progress").Optional("TOTAL_LAYER", "CURRENT_LAYER"))
}

// WithBedMesh returns a new dialect equal to d plus the Klipper
// [bed_mesh] commands (BED_MESH_CALIBRATE, BED_MESH_PROFILE,
// BED_MESH_CLEAR). d is not modified.
func WithBedMesh(d *gcode.Dialect) *gcode.Dialect {
	return d.Extend(d.Name()).
		Register(gcode.NewCommand("BED_MESH_CALIBRATE").Describe("probe and store bed mesh")).
		Register(gcode.NewCommand("BED_MESH_PROFILE").Describe("load/save/remove a bed mesh profile").Optional("LOAD", "SAVE", "REMOVE")).
		Register(gcode.NewCommand("BED_MESH_CLEAR").Describe("clear active bed mesh"))
}

// WithExcludeObject returns a new dialect equal to d plus the Klipper
// [exclude_object] commands (EXCLUDE_OBJECT_DEFINE, EXCLUDE_OBJECT_START,
// EXCLUDE_OBJECT_END, EXCLUDE_OBJECT). d is not modified.
func WithExcludeObject(d *gcode.Dialect) *gcode.Dialect {
	return d.Extend(d.Name()).
		Register(gcode.NewCommand("EXCLUDE_OBJECT_DEFINE").Describe("declare a printable object").Required("NAME").Optional("CENTER", "POLYGON")).
		Register(gcode.NewCommand("EXCLUDE_OBJECT_START").Describe("begin printing an object").Required("NAME")).
		Register(gcode.NewCommand("EXCLUDE_OBJECT_END").Describe("end printing an object").Optional("NAME")).
		Register(gcode.NewCommand("EXCLUDE_OBJECT").Describe("toggle object exclusion").Optional("NAME", "CURRENT", "RESET"))
}

// WithFanGeneric returns a new dialect equal to d plus the SET_FAN_SPEED
// command supplied by Klipper's [fan_generic] config sections.
// d is not modified.
func WithFanGeneric(d *gcode.Dialect) *gcode.Dialect {
	return d.Extend(d.Name()).
		Register(gcode.NewCommand("SET_FAN_SPEED").Describe("set named fan speed").Required("FAN").Optional("SPEED"))
}

// WithTimelapse returns a new dialect equal to d plus the
// TIMELAPSE_TAKE_FRAME command supplied by the moonraker-timelapse
// plugin. d is not modified.
func WithTimelapse(d *gcode.Dialect) *gcode.Dialect {
	return d.Extend(d.Name()).
		Register(gcode.NewCommand("TIMELAPSE_TAKE_FRAME").Describe("trigger timelapse frame"))
}
