package examples_test

import (
	"errors"
	"fmt"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
	"github.com/lestrrat-go/gcode/dialects/reprap"
)

func Example_marlinDialect() {
	d := marlin.Dialect()
	fmt.Printf("name: %s\n", d.Name())

	// Marlin knows G28 (auto home)
	def, ok := d.LookupCommand('G', 28, 0)
	if ok {
		fmt.Printf("G28: %s\n", def.Description)
	}

	// Marlin knows M104 (set hotend temp) with required S param
	def, ok = d.LookupCommand('M', 104, 0)
	if ok {
		fmt.Printf("M104: %s\n", def.Description)
		for _, p := range def.Params {
			fmt.Printf("  param %c required=%v\n", p.Letter, p.Required)
		}
	}

	// Marlin knows G92.1 (subcode)
	def, ok = d.LookupCommand('G', 92, 1)
	if ok {
		fmt.Printf("G92.1: %s (hasSubcode=%v)\n", def.Description, def.HasSubcode)
	}
	// Output:
	// name: marlin
	// G28: auto home
	// M104: set hotend temp
	//   param S required=true
	//   param T required=false
	// G92.1: reset position offset (hasSubcode=true)
}

func Example_repRapDialect() {
	d := reprap.Dialect()
	fmt.Printf("name: %s\n", d.Name())

	// RepRap inherits Marlin commands
	_, ok := d.LookupCommand('G', 28, 0)
	fmt.Printf("has G28 (from marlin): %v\n", ok)

	// RepRap adds its own commands
	def, ok := d.LookupCommand('G', 10, 0)
	if ok {
		fmt.Printf("G10: %s\n", def.Description)
	}

	def, ok = d.LookupCommand('M', 557, 0)
	if ok {
		fmt.Printf("M557: %s\n", def.Description)
	}

	// G999 is not in RepRap either
	_, ok = d.LookupCommand('G', 999, 0)
	fmt.Printf("has G999: %v\n", ok)
	// Output:
	// name: reprap
	// has G28 (from marlin): true
	// G10: retract/set tool offset
	// M557: define print bed mesh
	// has G999: false
}

func ExampleWithStrict_dialect() {
	d := marlin.Dialect()

	// Valid Marlin program
	valid := "G28\nM104 S200\nG1 X10 Y20 F3000\n"
	_, err := gcode.ParseString(valid, gcode.WithStrict(), gcode.WithDialect(d))
	fmt.Printf("valid program error: %v\n", err)

	// Program with unknown command for Marlin
	invalid := "G28\nG999\n"
	_, err = gcode.ParseString(invalid, gcode.WithStrict(), gcode.WithDialect(d))
	fmt.Printf("invalid is parse error: %v\n", errors.Is(err, gcode.ErrParse))
	// Output:
	// valid program error: <nil>
	// invalid is parse error: true
}
