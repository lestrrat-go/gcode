package gcode

// ParamDef describes a single parameter that a command accepts.
type ParamDef struct {
	// Key is the canonical argument key.
	// For classic single-letter parameters this is the upper-case letter
	// as a string ("X", "Y"). For extended commands it is the named
	// argument identifier ("FAN", "SPEED").
	Key         string
	Required    bool
	Description string
}

// CommandDef describes a G-code command recognized by a dialect.
//
// Name is the canonical command identifier as used on Command.Name —
// "G28", "M104", "G92.1", or "EXCLUDE_OBJECT_DEFINE".
type CommandDef struct {
	Name        string
	Description string
	// Params describes the parameters this command accepts.
	// nil means unconstrained; an empty slice means no parameters.
	Params []ParamDef
}

// Dialect is a named collection of command definitions describing which
// G-code commands a particular firmware understands and what arguments
// each one accepts. The Reader can consult a Dialect to validate input
// when strict mode is enabled.
type Dialect struct {
	name     string
	commands map[string]*CommandDef
}

// NewDialect creates an empty dialect with the given name.
func NewDialect(name string) *Dialect {
	return &Dialect{
		name:     name,
		commands: make(map[string]*CommandDef),
	}
}

// Name returns the dialect name.
func (d *Dialect) Name() string { return d.name }

// Register adds a command definition to the dialect. If a command with
// the same Name already exists it is silently overwritten.
func (d *Dialect) Register(def CommandDef) {
	d.commands[def.Name] = &def
}

// LookupCommand returns the command definition for the given canonical
// name and reports whether it was found.
func (d *Dialect) LookupCommand(name string) (*CommandDef, bool) {
	def, ok := d.commands[name]
	return def, ok
}

// Commands returns a copy of all registered command definitions.
func (d *Dialect) Commands() []CommandDef {
	out := make([]CommandDef, 0, len(d.commands))
	for _, def := range d.commands {
		out = append(out, *def)
	}
	return out
}

// Extend creates a child dialect that inherits all commands from this
// dialect. The returned dialect is independent — registering commands on
// the child does not affect the parent.
func (d *Dialect) Extend(name string) *Dialect {
	child := &Dialect{
		name:     name,
		commands: make(map[string]*CommandDef, len(d.commands)),
	}
	for k, v := range d.commands {
		cp := *v
		child.commands[k] = &cp
	}
	return child
}
