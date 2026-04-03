package gcode

// commandKey uniquely identifies a command by its letter, number, and subcode.
type commandKey struct {
	Letter  byte
	Number  int
	Subcode int // 0 means no subcode (subcode 0 cannot be distinguished from absent)
}

// ParamDef describes a single parameter that a command accepts.
type ParamDef struct {
	Letter      byte
	Required    bool
	Description string
}

// CommandDef describes a G-code command recognized by a dialect.
type CommandDef struct {
	Letter      byte
	Number      int
	Subcode     int
	HasSubcode  bool
	Description string
	Params      []ParamDef // nil = unconstrained, empty = no params
}

// Dialect is a named collection of command definitions that describes
// which G-code commands are recognized and what parameters they accept.
type Dialect struct {
	name     string
	commands map[commandKey]*CommandDef // flattened, includes inherited
}

// NewDialect creates a new dialect with the given name and an empty command set.
func NewDialect(name string) *Dialect {
	return &Dialect{
		name:     name,
		commands: make(map[commandKey]*CommandDef),
	}
}

// Name returns the dialect name.
func (d *Dialect) Name() string {
	return d.name
}

// Register adds a command definition to the dialect. If a command with the
// same letter, number, and subcode already exists, it is silently overwritten.
func (d *Dialect) Register(def CommandDef) {
	key := commandKey{
		Letter:  def.Letter,
		Number:  def.Number,
		Subcode: def.Subcode,
	}
	d.commands[key] = &def
}

// LookupCommand returns the command definition for the given letter, number,
// and subcode. The second return value indicates whether the command was found.
func (d *Dialect) LookupCommand(letter byte, number int, subcode int) (*CommandDef, bool) {
	key := commandKey{
		Letter:  letter,
		Number:  number,
		Subcode: subcode,
	}
	def, ok := d.commands[key]
	return def, ok
}

// Commands returns all registered command definitions.
func (d *Dialect) Commands() []CommandDef {
	out := make([]CommandDef, 0, len(d.commands))
	for _, def := range d.commands {
		out = append(out, *def)
	}
	return out
}

// Extend creates a new dialect that inherits all commands from this dialect.
// The returned dialect is independent; registering commands on the child
// does not affect the parent.
func (d *Dialect) Extend(name string) *Dialect {
	child := &Dialect{
		name:     name,
		commands: make(map[commandKey]*CommandDef, len(d.commands)),
	}
	for k, v := range d.commands {
		cp := *v
		child.commands[k] = &cp
	}
	return child
}
