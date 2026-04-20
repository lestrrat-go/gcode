package examples_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/klipper"
)

// ExampleReader_klipper enables strict parsing against the Klipper
// dialect. EXCLUDE_OBJECT_DEFINE is a config-dependent command, so
// the [klipper.WithExcludeObject] helper must be applied to the base
// dialect before strict mode will accept it.
func ExampleReader_klipper() {
	d := klipper.WithExcludeObject(klipper.Dialect())
	r := gcode.NewReader(
		strings.NewReader("EXCLUDE_OBJECT_DEFINE NAME=part_0\n"),
		gcode.WithDialect(d),
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
