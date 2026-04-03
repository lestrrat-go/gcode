package examples_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

func ExampleParse_roundtrip() {
	original := "G28\nG1 X10 Y20 F3000\nM104 S200\n"

	// Parse from io.Reader
	prog, err := gcode.Parse(strings.NewReader(original))
	if err != nil {
		panic(err)
	}

	// Generate back to io.Writer
	var buf bytes.Buffer
	if err := gcode.Generate(&buf, prog); err != nil {
		panic(err)
	}

	// Re-parse from bytes
	prog2, err := gcode.ParseBytes(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Verify round-trip preserves structure
	fmt.Printf("original lines: %d\n", prog.Len())
	fmt.Printf("roundtrip lines: %d\n", prog2.Len())

	for i := range prog.Len() {
		l1 := prog.Line(i)
		l2 := prog2.Line(i)
		fmt.Printf("line %d: %c%d == %c%d: %v\n", i,
			l1.Command.Letter, l1.Command.Number,
			l2.Command.Letter, l2.Command.Number,
			l1.Command.Letter == l2.Command.Letter && l1.Command.Number == l2.Command.Number,
		)
	}
	// Output:
	// original lines: 3
	// roundtrip lines: 3
	// line 0: G28 == G28: true
	// line 1: G1 == G1: true
	// line 2: M104 == M104: true
}

func ExampleNewParser() {
	input := "G0 X5 Y10\nG1 Z0.3 E1.5 F600\n"

	p := gcode.NewParser()

	// Parse via ParseString
	prog1, err := p.ParseString(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("ParseString lines: %d\n", prog1.Len())

	// Parse via ParseBytes
	prog2, err := p.ParseBytes([]byte(input))
	if err != nil {
		panic(err)
	}
	fmt.Printf("ParseBytes lines: %d\n", prog2.Len())

	// Parse via Parse (io.Reader)
	prog3, err := p.Parse(strings.NewReader(input))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parse lines: %d\n", prog3.Len())

	// All three produce the same result
	for i := range prog1.Len() {
		l := prog1.Line(i)
		fmt.Printf("%c%d params=%d\n", l.Command.Letter, l.Command.Number, len(l.Command.Params))
	}
	// Output:
	// ParseString lines: 2
	// ParseBytes lines: 2
	// Parse lines: 2
	// G0 params=2
	// G1 params=3
}
