package gcode

import "sync"

// DialectRegistry is a thread-safe registry of named dialects.
type DialectRegistry struct {
	mu       sync.RWMutex
	dialects map[string]*Dialect
}

// NewDialectRegistry creates a new empty dialect registry.
func NewDialectRegistry() *DialectRegistry {
	return &DialectRegistry{
		dialects: make(map[string]*Dialect),
	}
}

// Register adds a dialect to the registry, keyed by its name.
// If a dialect with the same name already exists, it is overwritten.
func (r *DialectRegistry) Register(d *Dialect) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dialects[d.Name()] = d
}

// Lookup returns the dialect with the given name. The second return value
// indicates whether the dialect was found.
func (r *DialectRegistry) Lookup(name string) (*Dialect, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.dialects[name]
	return d, ok
}
