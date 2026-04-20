package gcode

// Command is a single G-code command.
//
// Name is the canonical command identifier:
//   - Classic G/M/T codes: letter + number, with optional ".subcode".
//     Examples: "G28", "M104", "G92.1".
//   - Extended (Klipper-style) commands: the bare identifier as written.
//     Examples: "EXCLUDE_OBJECT_DEFINE", "SET_FAN_SPEED".
//
// Args are the address-value pairs that follow the command, in source order.
type Command struct {
	Name string
	Args []Argument
}

// Arg returns the first argument whose Key equals key, and true if found.
// Comparison is case-sensitive against the canonical key.
func (c Command) Arg(key string) (Argument, bool) {
	for i := range c.Args {
		if c.Args[i].Key == key {
			return c.Args[i], true
		}
	}
	return Argument{}, false
}

// Argument is one keyed value within a command.
//
// Examples (source on the left, struct fields on the right):
//
//	X10.5                  Key:"X"        Raw:"10.5"            Number:10.5  IsNumeric:true
//	X (bare flag)          Key:"X"        Raw:""
//	FAN=my_fan             Key:"FAN"      Raw:"my_fan"          Number:0     IsNumeric:false
//	SPEED=0.5              Key:"SPEED"    Raw:"0.5"             Number:0.5   IsNumeric:true
//	POLYGON=[[1,2],[3,4]]  Key:"POLYGON"  Raw:"[[1,2],[3,4]]"
//	MSG="hi"               Key:"MSG"      Raw:`"hi"`
//
// Key uses single-letter form for classic parameters ("X") and the
// identifier form for extended commands ("FAN", "POLYGON"). Raw preserves
// the value's source text verbatim so lists, quoted strings, and other
// non-numeric forms round-trip exactly. Number is populated only when
// Raw parses as a finite float; in that case IsNumeric is true.
//
// A bare flag — a parameter letter with no following value, as in "G28 X" —
// is represented by an empty Raw and IsNumeric=false. Use IsFlag to test.
type Argument struct {
	Key       string
	Raw       string
	Number    float64
	IsNumeric bool
}

// IsFlag reports whether the argument is a bare flag (no value).
func (a Argument) IsFlag() bool { return a.Raw == "" }
