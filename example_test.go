package gcode_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

func Example_parse() {
	prog, err := gcode.ParseString("G28\nG1 X10 Y20 F3000\n")
	if err != nil {
		panic(err)
	}
	fmt.Println(prog.Len())
	line := prog.Line(0)
	fmt.Printf("%c%d\n", line.Command.Letter, line.Command.Number)
	// Output:
	// 2
	// G28
}

func Example_generate() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.Line{
				HasCommand: true,
				Command:    gcode.Command{Letter: 'G', Number: 28},
			},
			gcode.Line{
				HasCommand: true,
				Command: gcode.Command{
					Letter: 'G',
					Number: 1,
					Params: []gcode.Parameter{
						{Letter: 'X', Value: 10},
						{Letter: 'Y', Value: 20},
					},
				},
			},
		).
		Build()

	var sb strings.Builder
	if err := gcode.Format(&sb, prog); err != nil {
		panic(err)
	}
	fmt.Print(sb.String())
	// Output:
	// G28
	// G1 X10 Y20
}

func Example_macro() {
	macro := gcode.NewSimpleMacro("home-all", []gcode.Line{
		{
			HasCommand: true,
			Command:    gcode.Command{Letter: 'G', Number: 28},
		},
	})

	reg := gcode.NewMacroRegistry()
	reg.Register(macro)

	lines, err := reg.Expand("home-all", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%c%d\n", lines[0].Command.Letter, lines[0].Command.Number)
	// Output:
	// G28
}
