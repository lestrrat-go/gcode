package examples_test

import (
	"fmt"

	"github.com/lestrrat-go/gcode"
)

func ExampleParseString_inspectAllFields() {
	input := `
; preheat sequence
(tool change)
G28 X Y
G92.1
G1 X10.5 Y20 F3000 ;move to start
`

	prog, err := gcode.ParseString(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("lines: %d\n", prog.Len())

	// Also verify Lines() returns the same data
	allLines := prog.Lines()
	fmt.Printf("Lines() len: %d\n", len(allLines))

	for i := range prog.Len() {
		line := prog.Line(i)

		fmt.Printf("\n--- line %d ---\n", i)
		fmt.Printf("blank: %v\n", line.IsBlank())
		fmt.Printf("raw: %q\n", line.Raw)

		if line.HasCommand {
			cmd := line.Command
			fmt.Printf("command: %c%d\n", cmd.Letter, cmd.Number)
			if cmd.HasSubcode {
				fmt.Printf("subcode: %d\n", cmd.Subcode)
			}
			for _, p := range cmd.Params {
				fmt.Printf("  param %c = %g\n", p.Letter, p.Value)
			}
		} else {
			fmt.Println("command: none")
		}

		if line.HasComment {
			fmt.Printf("comment: %q (form=%s)\n", line.Comment.Text, line.Comment.Form.String())
		}

		if line.HasChecksum {
			fmt.Printf("checksum: %d\n", line.Checksum)
		}

		fmt.Printf("lineNumber: %d\n", line.LineNumber)
	}

	// Output:
	// lines: 6
	// Lines() len: 6
	//
	// --- line 0 ---
	// blank: true
	// raw: ""
	// command: none
	// lineNumber: 0
	//
	// --- line 1 ---
	// blank: false
	// raw: "; preheat sequence"
	// command: none
	// comment: " preheat sequence" (form=semicolon)
	// lineNumber: 0
	//
	// --- line 2 ---
	// blank: false
	// raw: "(tool change)"
	// command: none
	// comment: "tool change" (form=parenthesis)
	// lineNumber: 0
	//
	// --- line 3 ---
	// blank: false
	// raw: "G28 X Y"
	// command: G28
	//   param X = 0
	//   param Y = 0
	// lineNumber: 0
	//
	// --- line 4 ---
	// blank: false
	// raw: "G92.1"
	// command: G92
	// subcode: 1
	// lineNumber: 0
	//
	// --- line 5 ---
	// blank: false
	// raw: "G1 X10.5 Y20 F3000 ;move to start"
	// command: G1
	//   param X = 10.5
	//   param Y = 20
	//   param F = 3000
	// comment: "move to start" (form=semicolon)
	// lineNumber: 0
}
