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
//
// Build a CommandDef with a struct literal or via the fluent
// constructor [NewCommand] / [CommandDef.Describe] / [CommandDef.Optional] /
// [CommandDef.Required].
type CommandDef struct {
	Name        string
	Description string
	// Params describes the parameters this command accepts.
	// nil means unconstrained; an empty slice means no parameters.
	Params []ParamDef
}

// NewCommand returns a CommandDef with the given canonical Name and an
// empty Params list. Chain [CommandDef.Describe], [CommandDef.Optional],
// and [CommandDef.Required] to populate it.
func NewCommand(name string) CommandDef {
	return CommandDef{Name: name, Params: []ParamDef{}}
}

// Describe returns a copy of c with Description set to s.
func (c CommandDef) Describe(s string) CommandDef {
	c.Description = s
	return c
}

// Optional returns a copy of c with one ParamDef per key appended,
// each marked as optional. Existing params are preserved; the
// underlying slice is reallocated so the returned value is fully
// independent of c.
func (c CommandDef) Optional(keys ...string) CommandDef {
	return c.appendParams(false, keys)
}

// Required returns a copy of c with one ParamDef per key appended,
// each marked as required.
func (c CommandDef) Required(keys ...string) CommandDef {
	return c.appendParams(true, keys)
}

func (c CommandDef) appendParams(required bool, keys []string) CommandDef {
	if len(keys) == 0 {
		return c
	}
	out := c
	out.Params = make([]ParamDef, len(c.Params), len(c.Params)+len(keys))
	copy(out.Params, c.Params)
	for _, k := range keys {
		out.Params = append(out.Params, ParamDef{Key: k, Required: required})
	}
	return out
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

// Register adds a command definition to the dialect and returns d so
// calls can be chained. If a command with the same Name already exists
// it is silently overwritten.
func (d *Dialect) Register(def CommandDef) *Dialect {
	d.commands[def.Name] = &def
	return d
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
