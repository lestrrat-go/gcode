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
			gcode.CommentLine(" start print"),
			gcode.G(28).Build(),
			gcode.G(1).X(50).Y(50).F(1500).Build(),
		).
		Build()

	var sb strings.Builder
	if err := gcode.Format(&sb, prog); err != nil {
		panic(err)
	}
	fmt.Print(sb.String())
	// Output:
	// ; start print
	// G28
	// G1 X50 Y50 F1500
}

func ExampleNewFormatter_emitComments() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.CommentLine(" this is a comment"),
			gcode.G(28).Comment("home all").Build(),
		).
		Build()

	// With comments (default)
	var sb1 strings.Builder
	if err := gcode.Format(&sb1, prog); err != nil {
		panic(err)
	}
	fmt.Print(sb1.String())

	fmt.Println("---")

	// Without comments
	var sb2 strings.Builder
	f := gcode.NewFormatter(gcode.WithEmitComments(false))
	if err := f.Format(&sb2, prog); err != nil {
		panic(err)
	}
	fmt.Print(sb2.String())
	// Output:
	// ; this is a comment
	// G28 ;home all
	// ---
	//
	// G28
}

func ExampleNewFormatter_lineNumbersAndChecksum() {
	// Line numbers are a Line-level field, so we build the command
	// and then set LineNumber on the resulting Line.
	l1 := gcode.G(28).Build()
	l1.LineNumber = 1
	l2 := gcode.G(1).X(10).Build()
	l2.LineNumber = 2

	prog := gcode.NewProgramBuilder().
		Append(l1, l2).
		Build()

	var sb strings.Builder
	f := gcode.NewFormatter(
		gcode.WithEmitLineNumbers(true),
		gcode.WithComputeChecksum(true),
	)
	if err := f.Format(&sb, prog); err != nil {
		panic(err)
	}
	fmt.Print(sb.String())
	// Output:
	// N1 G28*18
	// N2 G1 X10*83
}

func ExampleFormatter_FormatLine() {
	f := gcode.NewFormatter()
	var sb strings.Builder

	line := gcode.G(0).X(100).Y(200).Build()

	if err := f.FormatLine(&sb, line); err != nil {
		panic(err)
	}
	fmt.Println(sb.String())
	// Output:
	// G0 X100 Y200
}

func ExampleWithLineEnding_crlf() {
	prog := gcode.NewProgramBuilder().
		Append(
			gcode.G(28).Build(),
			gcode.G(0).Z(5).Build(),
		).
		Build()

	var buf bytes.Buffer
	f := gcode.NewFormatter(gcode.WithLineEnding(gcode.LineEndingCRLF))
	if err := f.Format(&buf, prog); err != nil {
		panic(err)
	}

	// Show that each line ends with \r\n
	data := buf.Bytes()
	for _, b := range data {
		switch b {
		case '\r':
			fmt.Print("\\r")
		case '\n':
			fmt.Print("\\n")
		default:
			fmt.Printf("%c", b)
		}
	}
	// Output:
	// G28\r\nG0 Z5\r\n
}
