package gcode

// Line represents one logical line of G-code. It may be a blank line,
// a comment-only line, a command line, or a command line with trailing comment.
//
// LineNumber and Checksum are optional; zero values mean absent.
// Line is a pure value type — copying a Line produces a fully independent value.
type Line struct {
	// LineNumber is the N field value. 0 means absent.
	LineNumber int
	// Command is valid only when HasCommand is true.
	Command Command
	// HasCommand indicates whether the line contains a command.
	HasCommand bool
	// Comment is valid only when HasComment is true.
	Comment Comment
	// HasComment indicates whether the line contains a comment.
	HasComment bool
	// Checksum is the computed checksum byte; valid only if HasChecksum is true.
	Checksum byte
	// HasChecksum indicates whether a checksum is present.
	HasChecksum bool
	// Raw is the original source text, populated by the parser. Useful for
	// commands with string arguments (M23, M117) that cannot be fully
	// structured into Parameter values.
	Raw string
}

// IsBlank returns true if the line has no command, no comment, and no line number.
func (l Line) IsBlank() bool {
	return !l.HasCommand && !l.HasComment && l.LineNumber == 0
}

// Program is an ordered sequence of lines that make up a G-code file or stream.
// It is immutable after construction; use ProgramBuilder to create one.
type Program struct {
	lines []Line
}

// Lines returns a copy of the line slice. Mutating the returned slice
// does not affect the Program.
func (p *Program) Lines() []Line {
	cp := make([]Line, len(p.lines))
	copy(cp, p.lines)
	return cp
}

// Len returns the number of lines in the program.
func (p *Program) Len() int {
	return len(p.lines)
}

// Line returns the line at index i. It panics if i is out of range.
func (p *Program) Line(i int) Line {
	return p.lines[i]
}

// ProgramBuilder provides a mutable construction API separate from Program.
// Append copies each Line's Command.Params slice to ensure full independence
// from the caller's data.
type ProgramBuilder struct {
	lines []Line
}

// NewProgramBuilder returns a new empty ProgramBuilder.
func NewProgramBuilder() *ProgramBuilder {
	return &ProgramBuilder{}
}

// Append adds lines to the builder. Each line's Command.Params slice is
// copied so that subsequent mutations to the caller's original slice do
// not affect the builder's stored lines.
func (b *ProgramBuilder) Append(lines ...Line) *ProgramBuilder {
	for _, l := range lines {
		if len(l.Command.Params) > 0 {
			cp := make([]Parameter, len(l.Command.Params))
			copy(cp, l.Command.Params)
			l.Command.Params = cp
		}
		b.lines = append(b.lines, l)
	}
	return b
}

// Build creates an immutable Program from the accumulated lines.
// The builder may be reused after calling Build.
func (b *ProgramBuilder) Build() *Program {
	lines := make([]Line, len(b.lines))
	copy(lines, b.lines)
	return &Program{lines: lines}
}
