// Package marlin provides a G-code dialect for Marlin firmware.
package marlin

import (
	"strconv"

	"github.com/lestrrat-go/gcode"
)

var dialect = build()

// Dialect returns the shared Marlin dialect singleton. Mutating it
// (e.g. via [gcode.Dialect.Register]) affects every caller. To extend
// with custom commands without disturbing other consumers, call
// [gcode.Dialect.Extend] to obtain a private child first.
func Dialect() *gcode.Dialect { return dialect }

func build() *gcode.Dialect {
	d := gcode.NewDialect("marlin")

	xyze := []gcode.ParamDef{{Key: "X"}, {Key: "Y"}, {Key: "Z"}, {Key: "E"}, {Key: "F"}}

	d.Register(gcode.CommandDef{Name: "G0", Description: "rapid move", Params: xyze})
	d.Register(gcode.CommandDef{Name: "G1", Description: "linear move", Params: xyze})
	d.Register(gcode.CommandDef{Name: "G4", Description: "dwell", Params: []gcode.ParamDef{{Key: "P"}, {Key: "S"}}})
	d.Register(gcode.CommandDef{Name: "G28", Description: "auto home", Params: []gcode.ParamDef{{Key: "X"}, {Key: "Y"}, {Key: "Z"}}})
	d.Register(gcode.CommandDef{Name: "G29", Description: "bed levelling"})
	d.Register(gcode.CommandDef{Name: "G92", Description: "set position", Params: []gcode.ParamDef{{Key: "X"}, {Key: "Y"}, {Key: "Z"}, {Key: "E"}}})
	d.Register(gcode.CommandDef{Name: "G92.1", Description: "reset position offset", Params: []gcode.ParamDef{}})

	d.Register(gcode.CommandDef{Name: "M82", Description: "absolute E", Params: []gcode.ParamDef{}})
	d.Register(gcode.CommandDef{Name: "M83", Description: "relative E", Params: []gcode.ParamDef{}})
	d.Register(gcode.CommandDef{Name: "M104", Description: "set hotend temp", Params: []gcode.ParamDef{{Key: "S", Required: true}, {Key: "T"}}})
	d.Register(gcode.CommandDef{Name: "M105", Description: "report temp", Params: []gcode.ParamDef{}})
	d.Register(gcode.CommandDef{Name: "M106", Description: "fan on", Params: []gcode.ParamDef{{Key: "S"}, {Key: "P"}}})
	d.Register(gcode.CommandDef{Name: "M107", Description: "fan off", Params: []gcode.ParamDef{{Key: "P"}}})
	d.Register(gcode.CommandDef{Name: "M109", Description: "wait hotend", Params: []gcode.ParamDef{{Key: "S", Required: true}, {Key: "R"}, {Key: "T"}}})
	d.Register(gcode.CommandDef{Name: "M140", Description: "set bed temp", Params: []gcode.ParamDef{{Key: "S", Required: true}}})
	d.Register(gcode.CommandDef{Name: "M190", Description: "wait bed", Params: []gcode.ParamDef{{Key: "S", Required: true}, {Key: "R"}}})
	d.Register(gcode.CommandDef{Name: "M204", Description: "set accel", Params: []gcode.ParamDef{{Key: "P"}, {Key: "R"}, {Key: "T"}}})
	d.Register(gcode.CommandDef{Name: "M220", Description: "feed rate %", Params: []gcode.ParamDef{{Key: "S"}}})
	d.Register(gcode.CommandDef{Name: "M221", Description: "flow rate %", Params: []gcode.ParamDef{{Key: "S"}}})

	for i := range 6 {
		d.Register(gcode.CommandDef{
			Name:        "T" + strconv.Itoa(i),
			Description: "tool select",
			Params:      []gcode.ParamDef{},
		})
	}

	return d
}
