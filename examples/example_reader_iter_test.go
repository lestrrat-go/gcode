package examples_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

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
