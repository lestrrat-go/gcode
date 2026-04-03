// Package marlin provides a G-code dialect for Marlin firmware.
package marlin

import "github.com/lestrrat-go/gcode"

// Dialect returns a fresh Marlin dialect instance. Each call returns an
// independent copy that can be modified without affecting other instances.
func Dialect() *gcode.Dialect {
	d := gcode.NewDialect("marlin")

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      0,
		Description: "rapid move",
		Params: []gcode.ParamDef{
			{Letter: 'X'}, {Letter: 'Y'}, {Letter: 'Z'},
			{Letter: 'E'}, {Letter: 'F'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      1,
		Description: "linear move",
		Params: []gcode.ParamDef{
			{Letter: 'X'}, {Letter: 'Y'}, {Letter: 'Z'},
			{Letter: 'E'}, {Letter: 'F'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      4,
		Description: "dwell",
		Params: []gcode.ParamDef{
			{Letter: 'P'}, {Letter: 'S'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      28,
		Description: "auto home",
		Params: []gcode.ParamDef{
			{Letter: 'X'}, {Letter: 'Y'}, {Letter: 'Z'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      29,
		Description: "bed levelling",
		Params:      nil, // unconstrained
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      92,
		Description: "set position",
		Params: []gcode.ParamDef{
			{Letter: 'X'}, {Letter: 'Y'}, {Letter: 'Z'}, {Letter: 'E'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      92,
		Subcode:     1,
		HasSubcode:  true,
		Description: "reset position offset",
		Params:      []gcode.ParamDef{},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      82,
		Description: "absolute E",
		Params:      []gcode.ParamDef{},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      83,
		Description: "relative E",
		Params:      []gcode.ParamDef{},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      104,
		Description: "set hotend temp",
		Params: []gcode.ParamDef{
			{Letter: 'S', Required: true},
			{Letter: 'T'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      105,
		Description: "report temp",
		Params:      []gcode.ParamDef{},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      106,
		Description: "fan on",
		Params: []gcode.ParamDef{
			{Letter: 'S'}, {Letter: 'P'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      107,
		Description: "fan off",
		Params: []gcode.ParamDef{
			{Letter: 'P'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      109,
		Description: "wait hotend",
		Params: []gcode.ParamDef{
			{Letter: 'S', Required: true},
			{Letter: 'R'},
			{Letter: 'T'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      140,
		Description: "set bed temp",
		Params: []gcode.ParamDef{
			{Letter: 'S', Required: true},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      190,
		Description: "wait bed",
		Params: []gcode.ParamDef{
			{Letter: 'S', Required: true},
			{Letter: 'R'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      204,
		Description: "set accel",
		Params: []gcode.ParamDef{
			{Letter: 'P'}, {Letter: 'R'}, {Letter: 'T'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      220,
		Description: "feed rate %",
		Params: []gcode.ParamDef{
			{Letter: 'S'},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'M',
		Number:      221,
		Description: "flow rate %",
		Params: []gcode.ParamDef{
			{Letter: 'S'},
		},
	})

	// T0-T5 tool select.
	for i := range 6 {
		d.Register(gcode.CommandDef{
			Letter:      'T',
			Number:      i,
			Description: "tool select",
			Params:      []gcode.ParamDef{},
		})
	}

	return d
}
