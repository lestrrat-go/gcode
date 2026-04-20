package examples_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

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
