package examples_test

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/gcode"
)

// tempMacro is a custom Macro that generates a set-temperature command
// using the "temp" argument for parameter substitution.
type tempMacro struct{}

func (tempMacro) Name() string { return "set-temp" }

func (tempMacro) Expand(args map[string]float64) ([]gcode.Line, error) {
	temp, ok := args["temp"]
	if !ok {
		return nil, fmt.Errorf("missing required arg: temp")
	}
	return []gcode.Line{
		gcode.M(104).S(temp).Build(),
		gcode.M(109).S(temp).Build(),
	}, nil
}

func ExampleMacro_custom() {
	m := tempMacro{}
	fmt.Printf("name: %s\n", m.Name())

	lines, err := m.Expand(map[string]float64{"temp": 210})
	if err != nil {
		panic(err)
	}

	gen := gcode.NewFormatter()
	for _, l := range lines {
		var sb fmt.Stringer = &lineStringer{gen: gen, line: l}
		fmt.Println(sb.String())
	}
	// Output:
	// name: set-temp
	// M104 S210
	// M109 S210
}

// lineStringer formats a line using a generator.
type lineStringer struct {
	gen  *gcode.Formatter
	line gcode.Line
}

func (ls *lineStringer) String() string {
	var sb stringBuilder
	if err := ls.gen.FormatLine(&sb, ls.line); err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return sb.String()
}

type stringBuilder struct {
	buf []byte
}

func (s *stringBuilder) Write(p []byte) (int, error) {
	s.buf = append(s.buf, p...)
	return len(p), nil
}

func (s *stringBuilder) String() string {
	return string(s.buf)
}

func ExampleNewSimpleMacro() {
	// A fixed preheat-PLA macro
	macro := gcode.NewSimpleMacro("preheat-pla", []gcode.Line{
		gcode.M(140).S(60).Build(),
		gcode.M(104).S(200).Build(),
		gcode.M(190).S(60).Build(),
		gcode.M(109).S(200).Build(),
	})

	fmt.Printf("name: %s\n", macro.Name())

	// Expand ignores args for SimpleMacro
	lines, err := macro.Expand(nil)
	if err != nil {
		panic(err)
	}

	prog := gcode.NewProgramBuilder().Append(lines...).Build()
	var sb strings.Builder
	if err := gcode.Format(&sb, prog); err != nil {
		panic(err)
	}
	fmt.Print(sb.String())
	// Output:
	// name: preheat-pla
	// M140 S60
	// M104 S200
	// M190 S60
	// M109 S200
}

func ExampleMacroRegistry() {
	reg := gcode.NewMacroRegistry()

	// Register a simple macro
	reg.Register(gcode.NewSimpleMacro("home-all", []gcode.Line{
		gcode.G(28).Build(),
	}))

	// Register a custom macro
	reg.Register(tempMacro{})

	// Lookup
	m, ok := reg.Lookup("home-all")
	fmt.Printf("home-all found: %v, name: %s\n", ok, m.Name())

	_, ok = reg.Lookup("nonexistent")
	fmt.Printf("nonexistent found: %v\n", ok)

	// Expand home-all
	lines, err := reg.Expand("home-all", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("home-all: %c%d\n", lines[0].Command.Letter, lines[0].Command.Number)

	// Expand set-temp with args
	lines, err = reg.Expand("set-temp", map[string]float64{"temp": 215})
	if err != nil {
		panic(err)
	}
	fmt.Printf("set-temp[0]: %c%d S=%.0f\n",
		lines[0].Command.Letter, lines[0].Command.Number, lines[0].Command.Params[0].Value)
	fmt.Printf("set-temp[1]: %c%d S=%.0f\n",
		lines[1].Command.Letter, lines[1].Command.Number, lines[1].Command.Params[0].Value)

	// Expand unknown macro returns error
	_, err = reg.Expand("unknown", nil)
	fmt.Printf("unknown error: %s\n", err)
	// Output:
	// home-all found: true, name: home-all
	// nonexistent found: false
	// home-all: G28
	// set-temp[0]: M104 S=215
	// set-temp[1]: M109 S=215
	// unknown error: gcode: macro "unknown" not registered
}
