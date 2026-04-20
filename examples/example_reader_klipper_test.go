package examples_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/klipper"
)

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
