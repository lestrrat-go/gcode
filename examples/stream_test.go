package examples_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/klipper"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

// ExampleReader_classic shows the simplest streaming read of classic
// G/M codes.
func ExampleReader_classic() {
	src := `; warm up
G28
M104 S200
G1 X10 Y20 F1500
`
	r := gcode.NewReader(strings.NewReader(src))
	var line gcode.Line
	for {
		err := r.Read(&line)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		if line.HasCommand {
			fmt.Println(line.Command.Name, len(line.Command.Args), "args")
		}
	}
	// Output:
	// G28 0 args
	// M104 1 args
	// G1 3 args
}

// ExampleReader_extended shows that Klipper-style extended commands
// flow through the same Reader. Argument keys retain their source
// case so emitter conventions round-trip.
func ExampleReader_extended() {
	src := `EXCLUDE_OBJECT_DEFINE NAME=part_0 CENTER=120,120 POLYGON=[[1,2],[3,4]]
SET_FAN_SPEED FAN=cooling SPEED=0.5
`
	r := gcode.NewReader(strings.NewReader(src))
	for line, err := range r.All() {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		if line.HasCommand {
			fmt.Println(line.Command.Name)
			for _, a := range line.Command.Args {
				fmt.Printf("  %s = %s\n", a.Key, a.Raw)
			}
		}
	}
	// Output:
	// EXCLUDE_OBJECT_DEFINE
	//   NAME = part_0
	//   CENTER = 120,120
	//   POLYGON = [[1,2],[3,4]]
	// SET_FAN_SPEED
	//   FAN = cooling
	//   SPEED = 0.5
}

// ExampleWriter shows constructing lines programmatically and writing
// them through a Writer.
func ExampleWriter() {
	var buf bytes.Buffer
	w := gcode.NewWriter(&buf)
	_ = w.Write(gcode.Line{HasCommand: true, Command: gcode.Command{Name: "G28"}})
	_ = w.Write(gcode.Line{HasCommand: true, Command: gcode.Command{
		Name: "G1",
		Args: []gcode.Argument{
			{Key: "X", Raw: "10"},
			{Key: "Y", Raw: "20"},
			{Key: "F", Raw: "1500"},
		},
	}})
	_ = w.Write(gcode.Line{HasCommand: true, Command: gcode.Command{
		Name: "SET_FAN_SPEED",
		Args: []gcode.Argument{
			{Key: "FAN", Raw: "cooling"},
			{Key: "SPEED", Raw: "0.5"},
		},
	}})
	_ = w.Flush()
	fmt.Print(buf.String())
	// Output:
	// G28
	// G1 X10 Y20 F1500
	// SET_FAN_SPEED FAN=cooling SPEED=0.5
}

// ExampleReader_strict illustrates dialect-aware strict-mode parsing.
// Unknown commands surface as parse errors that can be matched with
// errors.Is(err, gcode.ErrParse).
func ExampleReader_strict() {
	r := gcode.NewReader(
		strings.NewReader("G999\n"),
		gcode.WithDialect(marlin.Dialect()),
		gcode.WithStrict(),
	)
	var line gcode.Line
	err := r.Read(&line)
	fmt.Println(errors.Is(err, gcode.ErrParse))
	// Output:
	// true
}

// ExampleReader_klipper enables strict parsing against the Klipper
// dialect, accepting extended commands like EXCLUDE_OBJECT_DEFINE.
func ExampleReader_klipper() {
	r := gcode.NewReader(
		strings.NewReader("EXCLUDE_OBJECT_DEFINE NAME=part_0\n"),
		gcode.WithDialect(klipper.Dialect()),
		gcode.WithStrict(),
	)
	var line gcode.Line
	if err := r.Read(&line); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(line.Command.Name)
	// Output:
	// EXCLUDE_OBJECT_DEFINE
}

// ExampleMacroRegistry shows expanding a fixed macro into a stream.
func ExampleMacroRegistry() {
	reg := gcode.NewMacroRegistry()
	reg.Register(gcode.NewSimpleMacro("preheat-pla", []gcode.Line{
		{HasCommand: true, Command: gcode.Command{
			Name: "M140",
			Args: []gcode.Argument{{Key: "S", Raw: "60"}},
		}},
		{HasCommand: true, Command: gcode.Command{
			Name: "M104",
			Args: []gcode.Argument{{Key: "S", Raw: "200"}},
		}},
	}))

	expanded, _ := reg.Expand("preheat-pla", nil)

	var buf bytes.Buffer
	w := gcode.NewWriter(&buf)
	for _, l := range expanded {
		_ = w.Write(l)
	}
	_ = w.Flush()
	fmt.Print(buf.String())
	// Output:
	// M140 S60
	// M104 S200
}

// ExampleReader_iter uses the Go 1.23 range-over-func iterator.
func ExampleReader_iter() {
	r := gcode.NewReader(strings.NewReader("G28\nG1 X10\nG1 Y20\n"))
	for line, err := range r.All() {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(line.Command.Name)
	}
	// Output:
	// G28
	// G1
	// G1
}
