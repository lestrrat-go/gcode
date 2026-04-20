// Package reprap provides a G-code dialect for RepRap firmware,
// extending [marlin] with RepRap-specific commands.
package reprap

import (
	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

var dialect = build()

// Dialect returns the shared RepRap dialect singleton. The dialect
// inherits all Marlin commands and adds RepRap-specific ones.
// Mutating it affects every caller; use [gcode.Dialect.Extend] for a
// private child.
func Dialect() *gcode.Dialect { return dialect }

func build() *gcode.Dialect {
	return marlin.Dialect().Extend("reprap").
		Register(gcode.NewCommand("G10").Describe("retract / set tool offset").Optional("R", "S", "X", "Y", "Z")).
		Register(gcode.NewCommand("G11").Describe("unretract")).
		Register(gcode.NewCommand("M116").Describe("wait for temperatures")).
		Register(gcode.NewCommand("M557").Describe("define print bed mesh").Optional("X", "Y", "R", "S")).
		Register(gcode.NewCommand("M558").Describe("set probe type").Optional("P", "R", "S"))
}
