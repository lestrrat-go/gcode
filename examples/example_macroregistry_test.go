package examples_test

import (
	"bytes"
	"fmt"

	"github.com/lestrrat-go/gcode"
)

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
