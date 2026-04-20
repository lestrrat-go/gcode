package examples_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

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
