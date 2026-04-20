package gcode

import "strconv"

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

// NewLine returns a Line carrying a command with the given canonical
// Name. Chain [Line.Arg], [Line.ArgF], [Line.Flag], [Line.WithComment],
// [Line.WithParenComment], [Line.LineNo], and [Line.WithChecksum] to
// populate it.
func NewLine(name string) Line {
	return Line{HasCommand: true, Command: Command{Name: name}}
}

// NewComment returns a comment-only Line in semicolon form.
func NewComment(text string) Line {
	return Line{HasComment: true, Comment: Comment{Text: text, Form: CommentSemicolon}}
}

// NewParenComment returns a comment-only Line in parenthesis form.
func NewParenComment(text string) Line {
	return Line{HasComment: true, Comment: Comment{Text: text, Form: CommentParenthesis}}
}

// LineNo returns a copy of l with the source line-number prefix set to n.
func (l Line) LineNo(n int) Line {
	l.LineNumber = n
	return l
}

// Arg returns a copy of l with one Argument appended whose Key is key
// and Raw is raw. Number and IsNumeric are populated automatically when
// raw parses as a finite float64. The call is a no-op when l carries no
// command.
func (l Line) Arg(key, raw string) Line {
	if !l.HasCommand {
		return l
	}
	a := Argument{Key: key, Raw: raw}
	if v, err := strconv.ParseFloat(raw, 64); err == nil {
		a.Number = v
		a.IsNumeric = true
	}
	return l.appendArg(a)
}

// ArgF returns a copy of l with one numeric Argument appended whose Key
// is key, Number is v, and Raw is the canonical formatting of v
// (matching what [Writer] produces). No-op when l carries no command.
func (l Line) ArgF(key string, v float64) Line {
	if !l.HasCommand {
		return l
	}
	return l.appendArg(Argument{
		Key:       key,
		Raw:       strconv.FormatFloat(v, 'f', -1, 64),
		Number:    v,
		IsNumeric: true,
	})
}

// Flag returns a copy of l with one bare-flag Argument appended. No-op
// when l carries no command.
func (l Line) Flag(key string) Line {
	if !l.HasCommand {
		return l
	}
	return l.appendArg(Argument{Key: key})
}

// WithComment returns a copy of l with a trailing semicolon-form
// comment set (replacing any previous comment).
func (l Line) WithComment(text string) Line {
	l.HasComment = true
	l.Comment = Comment{Text: text, Form: CommentSemicolon}
	return l
}

// WithParenComment returns a copy of l with a trailing parenthesis-form
// comment set (replacing any previous comment).
func (l Line) WithParenComment(text string) Line {
	l.HasComment = true
	l.Comment = Comment{Text: text, Form: CommentParenthesis}
	return l
}

// WithChecksum returns a copy of l with the per-line checksum byte set.
func (l Line) WithChecksum(b byte) Line {
	l.Checksum = b
	l.HasChecksum = true
	return l
}

func (l Line) appendArg(a Argument) Line {
	args := make([]Argument, len(l.Command.Args), len(l.Command.Args)+1)
	copy(args, l.Command.Args)
	l.Command.Args = append(args, a)
	return l
}
