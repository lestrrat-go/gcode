// Package reprap provides a G-code dialect for RepRap firmware,
// extending [marlin] with RepRap-specific commands.
package reprap

import (
	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

// Dialect returns a fresh RepRap dialect instance. Each call returns
// an independent copy that can be modified without affecting other
// instances. The dialect inherits all Marlin commands and adds
// RepRap-specific ones.
func Dialect() *gcode.Dialect {
	d := marlin.Dialect().Extend("reprap")

	d.Register(gcode.CommandDef{
		Name:        "G10",
		Description: "retract / set tool offset",
		Params: []gcode.ParamDef{
			{Key: "R"}, {Key: "S"},
			{Key: "X"}, {Key: "Y"}, {Key: "Z"},
		},
	})
	d.Register(gcode.CommandDef{Name: "G11", Description: "unretract", Params: []gcode.ParamDef{}})
	d.Register(gcode.CommandDef{Name: "M116", Description: "wait for temperatures", Params: []gcode.ParamDef{}})
	d.Register(gcode.CommandDef{
		Name:        "M557",
		Description: "define print bed mesh",
		Params: []gcode.ParamDef{
			{Key: "X"}, {Key: "Y"}, {Key: "R"}, {Key: "S"},
		},
	})
	d.Register(gcode.CommandDef{
		Name:        "M558",
		Description: "set probe type",
		Params: []gcode.ParamDef{
			{Key: "P"}, {Key: "R"}, {Key: "S"},
		},
	})

	return d
}
