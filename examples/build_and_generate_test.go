package examples_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

func ExampleProgramBuilder() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.Line{
				HasComment: true,
				Comment:    gcode.Comment{Text: " start print", Form: gcode.CommentSemicolon},
			},
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
						{Letter: 'X', Value: 50},
						{Letter: 'Y', Value: 50},
						{Letter: 'F', Value: 1500},
					},
				},
			},
		).
		Build()

	s, err := gcode.GenerateString(prog)
	if err != nil {
		panic(err)
	}
	fmt.Print(s)
	// Output:
	// ; start print
	// G28
	// G1 X50 Y50 F1500
}

func ExampleGenerateBytes() {
	prog := gcode.NewProgramBuilder().
		Append(gcode.Line{
			HasCommand: true,
			Command: gcode.Command{
				Letter: 'M',
				Number: 104,
				Params: []gcode.Parameter{{Letter: 'S', Value: 200}},
			},
		}).
		Build()

	b, err := gcode.GenerateBytes(prog)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", b)
	// Output:
	// M104 S200
}

func ExampleNewGenerator_emitComments() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.Line{
				HasComment: true,
				Comment:    gcode.Comment{Text: " this is a comment", Form: gcode.CommentSemicolon},
			},
			gcode.Line{
				HasCommand: true,
				Command:    gcode.Command{Letter: 'G', Number: 28},
				HasComment: true,
				Comment:    gcode.Comment{Text: "home all", Form: gcode.CommentSemicolon},
			},
		).
		Build()

	// With comments (default)
	s1, err := gcode.GenerateString(prog)
	if err != nil {
		panic(err)
	}
	fmt.Print(s1)

	fmt.Println("---")

	// Without comments
	s2, err := gcode.GenerateString(prog, gcode.WithEmitComments(false))
	if err != nil {
		panic(err)
	}
	fmt.Print(s2)
	// Output:
	// ; this is a comment
	// G28 ;home all
	// ---
	//
	// G28
}

func ExampleNewGenerator_lineNumbersAndChecksum() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.Line{
				LineNumber: 1,
				HasCommand: true,
				Command:    gcode.Command{Letter: 'G', Number: 28},
			},
			gcode.Line{
				LineNumber: 2,
				HasCommand: true,
				Command: gcode.Command{
					Letter: 'G',
					Number: 1,
					Params: []gcode.Parameter{
						{Letter: 'X', Value: 10},
					},
				},
			},
		).
		Build()

	s, err := gcode.GenerateString(prog,
		gcode.WithEmitLineNumbers(true),
		gcode.WithComputeChecksum(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Print(s)
	// Output:
	// N1 G28*18
	// N2 G1 X10*83
}

func ExampleGenerator_GenerateLine() {
	gen := gcode.NewGenerator()
	var sb strings.Builder

	line := gcode.Line{
		HasCommand: true,
		Command: gcode.Command{
			Letter: 'G',
			Number: 0,
			Params: []gcode.Parameter{
				{Letter: 'X', Value: 100},
				{Letter: 'Y', Value: 200},
			},
		},
	}

	if err := gen.GenerateLine(&sb, line); err != nil {
		panic(err)
	}
	fmt.Println(sb.String())
	// Output:
	// G0 X100 Y200
}

func ExampleWithLineEnding_crlf() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.Line{
				HasCommand: true,
				Command:    gcode.Command{Letter: 'G', Number: 28},
			},
			gcode.Line{
				HasCommand: true,
				Command:    gcode.Command{Letter: 'G', Number: 0, Params: []gcode.Parameter{{Letter: 'Z', Value: 5}}},
			},
		).
		Build()

	var buf bytes.Buffer
	gen := gcode.NewGenerator(gcode.WithLineEnding(gcode.LineEndingCRLF))
	if err := gen.Generate(&buf, prog); err != nil {
		panic(err)
	}

	// Show that each line ends with \r\n
	data := buf.Bytes()
	for _, b := range data {
		if b == '\r' {
			fmt.Print("\\r")
		} else if b == '\n' {
			fmt.Print("\\n")
		} else {
			fmt.Printf("%c", b)
		}
	}
	// Output:
	// G28\r\nG0 Z5\r\n
}
