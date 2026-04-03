package examples_test

import (
	"errors"
	"fmt"

	"github.com/lestrrat-go/gcode"
	"github.com/lestrrat-go/gcode/dialects/marlin"
)

func ExampleErrParse() {
	// Parse malformed G-code: missing digits after command letter
	_, err := gcode.ParseString("G\n")
	if err == nil {
		panic("expected error")
	}

	fmt.Println(errors.Is(err, gcode.ErrParse))
	// Output:
	// true
}

func ExampleParseErrorDetail() {
	_, err := gcode.ParseString("G\n")
	if err == nil {
		panic("expected error")
	}

	var detail gcode.ParseErrorDetail
	if errors.As(err, &detail) {
		fmt.Printf("line: %d\n", detail.Line())
		fmt.Printf("column: %d\n", detail.Column())
		fmt.Printf("text: %q\n", detail.Text())
	}
	// Output:
	// line: 1
	// column: 2
	// text: ""
}

func ExampleNewParseError() {
	err := gcode.NewParseError(5, 3, "XYZ", fmt.Errorf("custom error"))
	fmt.Println(err)
	fmt.Println(errors.Is(err, gcode.ErrParse))
	// Output:
	// gcode: parse error at line 5 col 3: custom error (near "XYZ")
	// true
}

func ExampleWithStrict() {
	d := marlin.Dialect()

	// G28 is known to Marlin -- parses fine
	_, err := gcode.ParseString("G28\n", gcode.WithStrict(), gcode.WithDialect(d))
	fmt.Printf("G28 error: %v\n", err)

	// G999 is not known to Marlin -- strict mode rejects it
	_, err = gcode.ParseString("G999\n", gcode.WithStrict(), gcode.WithDialect(d))
	fmt.Printf("G999 is parse error: %v\n", errors.Is(err, gcode.ErrParse))

	var detail gcode.ParseErrorDetail
	if errors.As(err, &detail) {
		fmt.Printf("G999 detail line: %d\n", detail.Line())
	}
	// Output:
	// G28 error: <nil>
	// G999 is parse error: true
	// G999 detail line: 1
}
