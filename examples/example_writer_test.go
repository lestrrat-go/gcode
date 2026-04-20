package examples_test

import (
	"bytes"
	"fmt"

	"github.com/lestrrat-go/gcode"
)

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
