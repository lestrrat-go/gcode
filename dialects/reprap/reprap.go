// Package reprap provides a G-code dialect for RepRap firmware.
// It extends the Marlin dialect with RepRap-specific commands.
package reprap

import (
	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

// Dialect returns a fresh RepRap dialect instance. Each call returns an
// independent copy that can be modified without affecting other instances.
// The dialect inherits all Marlin commands and adds RepRap-specific ones.
func Dialect() *gcode.Dialect {
	d := marlin.Dialect().Extend("reprap")

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      10,
		Description: "retract/set tool offset",
		Params: []gcode.ParamDef{
			{Letter: 'R'}, {Letter: 'S'},
			{Letter: 'X'}, {Letter: 'Y'}, {Letter: 'Z'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      11,
		Description: "unretract",
		Params:      []gcode.ParamDef{},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      116,
		Description: "wait for temperatures",
		Params:      []gcode.ParamDef{},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      557,
		Description: "define print bed mesh",
		Params: []gcode.ParamDef{
			{Letter: 'X'}, {Letter: 'Y'},
			{Letter: 'R'}, {Letter: 'S'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      558,
		Description: "set probe type",
		Params: []gcode.ParamDef{
			{Letter: 'P'}, {Letter: 'R'}, {Letter: 'S'},
		},
	})

	return d
}
