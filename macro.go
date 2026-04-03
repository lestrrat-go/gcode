package gcode

import (
	"fmt"
	"sync"
)

// Macro defines a named expansion that produces a sequence of Lines.
// The library provides SimpleMacro (fixed lines, no substitution) as a
// convenience. Users who need parameter substitution implement this
// interface themselves — the Expand method gives full control over how
// args are used to construct the returned Lines. The library does not
// provide template markers or expression evaluation; substitution logic
// is entirely the implementer's responsibility.
type Macro interface {
	Name() string
	Expand(args map[string]float64) ([]Line, error)
}

// SimpleMacro is a Macro backed by a fixed slice of Lines.
// It ignores the args parameter and always returns a deep copy of its
// stored lines. Use SimpleMacro for fixed command sequences like
// "preheat PLA" that need no parameter substitution.
type SimpleMacro struct {
	name  string
	lines []Line
}

// NewSimpleMacro returns a new SimpleMacro with the given name and lines.
// The provided lines are deep-copied so that subsequent mutations to the
// caller's slice do not affect the macro.
func NewSimpleMacro(name string, lines []Line) *SimpleMacro {
	cp := make([]Line, len(lines))
	for i, l := range lines {
		if len(l.Command.Params) > 0 {
			params := make([]Parameter, len(l.Command.Params))
			copy(params, l.Command.Params)
			l.Command.Params = params
		}
		cp[i] = l
	}
	return &SimpleMacro{
		name:  name,
		lines: cp,
	}
}

// Name returns the macro's name.
func (m *SimpleMacro) Name() string {
	return m.name
}

// Expand returns a deep copy of the stored lines. The args parameter is
// ignored; SimpleMacro always returns the same sequence regardless of
// arguments.
func (m *SimpleMacro) Expand(_ map[string]float64) ([]Line, error) {
	cp := make([]Line, len(m.lines))
	for i, l := range m.lines {
		if len(l.Command.Params) > 0 {
			params := make([]Parameter, len(l.Command.Params))
			copy(params, l.Command.Params)
			l.Command.Params = params
		}
		cp[i] = l
	}
	return cp, nil
}

// MacroRegistry is a collection of named macros.
// Lookup and Expand are safe for concurrent use; Register acquires a
// write lock and must not be called concurrently with itself.
type MacroRegistry struct {
	mu     sync.RWMutex
	macros map[string]Macro
}

// NewMacroRegistry returns a new empty MacroRegistry.
func NewMacroRegistry() *MacroRegistry {
	return &MacroRegistry{
		macros: make(map[string]Macro),
	}
}

// Register adds or replaces a macro in the registry.
func (r *MacroRegistry) Register(m Macro) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.macros[m.Name()] = m
}

// Lookup returns the named macro and true, or nil and false if not found.
func (r *MacroRegistry) Lookup(name string) (Macro, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.macros[name]
	return m, ok
}

// Expand looks up the named macro and calls its Expand method.
// Returns an error if the macro is not registered.
func (r *MacroRegistry) Expand(name string, args map[string]float64) ([]Line, error) {
	r.mu.RLock()
	m, ok := r.macros[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("gcode: macro %q not registered", name)
	}
	return m.Expand(args)
}
