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
	d := gcode.NewDialect("marlin").
		Register(gcode.NewCommand("G0").Describe("rapid move").Optional("X", "Y", "Z", "E", "F")).
		Register(gcode.NewCommand("G1").Describe("linear move").Optional("X", "Y", "Z", "E", "F")).
		Register(gcode.NewCommand("G4").Describe("dwell").Optional("P", "S")).
		Register(gcode.NewCommand("G28").Describe("auto home").Optional("X", "Y", "Z")).
		Register(gcode.NewCommand("G29").Describe("bed levelling")).
		Register(gcode.NewCommand("G92").Describe("set position").Optional("X", "Y", "Z", "E")).
		Register(gcode.NewCommand("G92.1").Describe("reset position offset")).
		Register(gcode.NewCommand("M82").Describe("absolute E")).
		Register(gcode.NewCommand("M83").Describe("relative E")).
		Register(gcode.NewCommand("M104").Describe("set hotend temp").Required("S").Optional("T")).
		Register(gcode.NewCommand("M105").Describe("report temp")).
		Register(gcode.NewCommand("M106").Describe("fan on").Optional("S", "P")).
		Register(gcode.NewCommand("M107").Describe("fan off").Optional("P")).
		Register(gcode.NewCommand("M109").Describe("wait hotend").Required("S").Optional("R", "T")).
		Register(gcode.NewCommand("M140").Describe("set bed temp").Required("S")).
		Register(gcode.NewCommand("M190").Describe("wait bed").Required("S").Optional("R")).
		Register(gcode.NewCommand("M204").Describe("set accel").Optional("P", "R", "T")).
		Register(gcode.NewCommand("M220").Describe("feed rate %").Optional("S")).
		Register(gcode.NewCommand("M221").Describe("flow rate %").Optional("S"))

	for i := range 6 {
		d.Register(gcode.NewCommand("T" + strconv.Itoa(i)).Describe("tool select"))
	}
	return d
}
