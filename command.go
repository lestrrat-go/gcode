package gcode

// Parameter represents a single address-value pair, e.g. X10.5 or S200.
type Parameter struct {
	// Letter is the parameter address letter ('X', 'Y', 'Z', 'E', 'F', 'S', etc.).
	Letter byte
	// Value is the numeric value of the parameter.
	Value float64
}

// Command represents a single G/M/T (or other letter) code with parameters.
type Command struct {
	// Letter is the command letter, e.g. 'G', 'M', 'T'.
	Letter byte
	// Number is the numeric portion, e.g. 28 for G28.
	Number int
	// Subcode is the subcode after the dot, e.g. 1 for G92.1; 0 means absent.
	Subcode int
	// HasSubcode indicates whether a subcode is present.
	HasSubcode bool
	// Params holds the address-value pairs following the command code.
	Params []Parameter
}
