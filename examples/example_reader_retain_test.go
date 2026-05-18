package examples_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

// ExampleReader_retain shows how to safely retain Lines past the next
// [gcode.Reader.Read] call. Each line returned by Read aliases the
// Reader's internal buffer and is invalidated on the next Read, so
// callers that need to keep a Line must call [gcode.Line.Clone] before
// storing it.
//
// Forgetting Clone here would leave every entry in `moves` pointing at
// whatever line the Reader happens to be parsing at the end of the
// loop.
func ExampleReader_retain() {
	src := `G0 X0 Y0
G1 X10 Y10
M104 S200
G1 X20 Y10
`
	r := gcode.NewReader(strings.NewReader(src))

	var moves []gcode.Line
	for line, err := range r.All() {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		if !line.HasCommand || line.Command.Name != "G1" {
			continue
		}
		moves = append(moves, line.Clone())
	}

	for _, m := range moves {
		fmt.Println(m.Command.Name, len(m.Command.Args))
	}
	// Output:
	// G1 2
	// G1 2
}
