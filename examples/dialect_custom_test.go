package examples_test

import (
	"fmt"
	"sort"

	"github.com/lestrrat-go/gcode"
)

func ExampleNewDialect() {
	d := gcode.NewDialect("my-cnc")
	fmt.Println(d.Name())

	// Register commands with full definitions
	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      0,
		Description: "rapid move",
		Params: []gcode.ParamDef{
			{Letter: 'X', Required: false, Description: "X position"},
			{Letter: 'Y', Required: false, Description: "Y position"},
			{Letter: 'Z', Required: false, Description: "Z position"},
			{Letter: 'F', Required: false, Description: "feed rate"},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      1,
		Description: "linear move",
		Params: []gcode.ParamDef{
			{Letter: 'X', Required: false, Description: "X position"},
			{Letter: 'Y', Required: false, Description: "Y position"},
			{Letter: 'Z', Required: false, Description: "Z position"},
			{Letter: 'F', Required: true, Description: "feed rate"},
		},
	})

	d.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      92,
		Subcode:     1,
		HasSubcode:  true,
		Description: "reset offsets",
		Params:      []gcode.ParamDef{},
	})

	// Lookup existing command
	def, ok := d.LookupCommand('G', 0, 0)
	if ok {
		fmt.Printf("G0: %s\n", def.Description)
		fmt.Printf("G0 params: %d\n", len(def.Params))
		fmt.Printf("G0 param[0]: %c required=%v desc=%q\n",
			def.Params[0].Letter, def.Params[0].Required, def.Params[0].Description)
	}

	// Lookup subcode command
	def, ok = d.LookupCommand('G', 92, 1)
	if ok {
		fmt.Printf("G92.1: %s (subcode=%d, hasSubcode=%v)\n",
			def.Description, def.Subcode, def.HasSubcode)
	}

	// Lookup missing command
	_, ok = d.LookupCommand('M', 999, 0)
	fmt.Printf("M999 found: %v\n", ok)

	// Commands() returns all definitions (sorted for determinism)
	cmds := d.Commands()
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].Letter != cmds[j].Letter {
			return cmds[i].Letter < cmds[j].Letter
		}
		if cmds[i].Number != cmds[j].Number {
			return cmds[i].Number < cmds[j].Number
		}
		return cmds[i].Subcode < cmds[j].Subcode
	})
	fmt.Printf("total commands: %d\n", len(cmds))
	for _, c := range cmds {
		if c.HasSubcode {
			fmt.Printf("  %c%d.%d\n", c.Letter, c.Number, c.Subcode)
		} else {
			fmt.Printf("  %c%d\n", c.Letter, c.Number)
		}
	}
	// Output:
	// my-cnc
	// G0: rapid move
	// G0 params: 4
	// G0 param[0]: X required=false desc="X position"
	// G92.1: reset offsets (subcode=1, hasSubcode=true)
	// M999 found: false
	// total commands: 3
	//   G0
	//   G1
	//   G92.1
}

func ExampleDialect_Extend() {
	base := gcode.NewDialect("base")
	base.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      0,
		Description: "rapid move",
	})
	base.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      1,
		Description: "linear move",
	})

	// Extend creates an independent child dialect
	child := base.Extend("extended")
	child.Register(gcode.CommandDef{
		Letter:      'G',
		Number:      2,
		Description: "arc CW",
	})

	fmt.Printf("base name: %s\n", base.Name())
	fmt.Printf("child name: %s\n", child.Name())
	fmt.Printf("base commands: %d\n", len(base.Commands()))
	fmt.Printf("child commands: %d\n", len(child.Commands()))

	// Child has parent commands
	_, ok := child.LookupCommand('G', 0, 0)
	fmt.Printf("child has G0: %v\n", ok)

	// Parent does NOT have child-only commands
	_, ok = base.LookupCommand('G', 2, 0)
	fmt.Printf("base has G2: %v\n", ok)
	// Output:
	// base name: base
	// child name: extended
	// base commands: 2
	// child commands: 3
	// child has G0: true
	// base has G2: false
}

func ExampleDialectRegistry() {
	reg := gcode.NewDialectRegistry()

	d1 := gcode.NewDialect("grbl")
	d1.Register(gcode.CommandDef{Letter: 'G', Number: 0, Description: "rapid"})

	d2 := gcode.NewDialect("smoothie")
	d2.Register(gcode.CommandDef{Letter: 'G', Number: 0, Description: "rapid"})

	reg.Register(d1)
	reg.Register(d2)

	found, ok := reg.Lookup("grbl")
	fmt.Printf("grbl found: %v, name: %s\n", ok, found.Name())

	_, ok = reg.Lookup("unknown")
	fmt.Printf("unknown found: %v\n", ok)
	// Output:
	// grbl found: true, name: grbl
	// unknown found: false
}
