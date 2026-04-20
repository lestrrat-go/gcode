package gcode

// Line represents one source line of G-code. A single line can carry any
// combination of an optional line number (N), a command, a trailing
// comment, and a checksum. Each component is gated by its corresponding
// HasX field; check those before reading the value.
//
// Lines returned by Reader.Read are backed by the Reader's internal
// buffers and remain valid only until the next Read call. To retain a
// Line beyond that point, call Clone.
type Line struct {
	LineNumber  int
	Command     Command
	HasCommand  bool
	Comment     Comment
	HasComment  bool
	Checksum    byte
	HasChecksum bool
	// Raw is the original source text of the line (without the trailing
	// newline). The Reader populates this; the Writer ignores it.
	Raw string
}

// IsBlank reports whether the line carries no command, no comment, no
// line number, and no checksum.
func (l Line) IsBlank() bool {
	return !l.HasCommand && !l.HasComment && !l.HasChecksum && l.LineNumber == 0
}

// Clone returns a deep copy of l with detached storage. Use this to
// retain a Line returned by Reader.Read past the next Read call.
func (l Line) Clone() Line {
	out := Line{
		LineNumber:  l.LineNumber,
		HasCommand:  l.HasCommand,
		HasComment:  l.HasComment,
		Checksum:    l.Checksum,
		HasChecksum: l.HasChecksum,
		Raw:         cloneString(l.Raw),
	}
	if l.HasCommand {
		out.Command.Name = cloneString(l.Command.Name)
		if len(l.Command.Args) > 0 {
			out.Command.Args = make([]Argument, len(l.Command.Args))
			for i, a := range l.Command.Args {
				out.Command.Args[i] = Argument{
					Key:       cloneString(a.Key),
					Raw:       cloneString(a.Raw),
					Number:    a.Number,
					IsNumeric: a.IsNumeric,
				}
			}
		}
	}
	if l.HasComment {
		out.Comment = Comment{
			Text: cloneString(l.Comment.Text),
			Form: l.Comment.Form,
		}
	}
	return out
}

// reset returns the Line to its zero state while preserving any backing
// Args slice capacity, so that the Reader can re-fill the same Line
// without allocating.
func (l *Line) reset() {
	args := l.Command.Args
	if args != nil {
		args = args[:0]
	}
	*l = Line{Command: Command{Args: args}}
}

// cloneString forces a copy of s into a fresh allocation.
// Without this, strings handed out by the Reader would still point into
// the Reader's reusable line buffer.
func cloneString(s string) string {
	if s == "" {
		return ""
	}
	b := make([]byte, len(s))
	copy(b, s)
	return string(b)
}
