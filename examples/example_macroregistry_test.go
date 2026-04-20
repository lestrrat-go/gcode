package examples_test

import (
	"bytes"
	"fmt"

	"github.com/lestrrat-go/gcode"
)

// ExampleMacroRegistry shows expanding a fixed macro into a stream
// using the fluent Line constructors.
func ExampleMacroRegistry() {
	reg := gcode.NewMacroRegistry().
		Register(gcode.NewSimpleMacro("preheat-pla",
			gcode.NewLine("M140").ArgF("S", 60),
			gcode.NewLine("M104").ArgF("S", 200),
		))

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
