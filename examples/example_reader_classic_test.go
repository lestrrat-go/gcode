package examples_test

import (
	"fmt"
	"io"
	"strings"

	"github.com/lestrrat-go/gcode"
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
