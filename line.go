package gcode

import "strconv"

// Line represents one source line of G-code. A single line can carry any
// combination of an optional line number (N), a command, a trailing
// comment, and a checksum. Each component is gated by its corresponding
// HasX field; check those before reading the value.
//
// Lines returned by [Reader.Read] alias the Reader's internal buffers
// and become invalid on the next Read — see [Reader.Read] for the full
// buffer-ownership contract and use [Line.Clone] to retain a Line.
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
		out.Command = l.Command.Clone()
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
// is key, Number is v, and Raw is the shortest float64 representation
// of v that round-trips. No-op when l carries no command.
//
// ArgF stores the shortest form, which means float64 noise can leak
// into output. Use [Line.ArgFP] when the call site knows the required
// decimal precision.
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

// ArgFP returns a copy of l with one numeric Argument appended whose
// Key is key, Number is v, and Raw is v formatted at the given
// decimal precision (matching [strconv.FormatFloat] with the 'f'
// verb). Use this when output convention demands a specific number of
// decimal places — for example, slicer-canonical "X10.000" or
// "E1.03365" — independent of v's underlying float64 representation.
// No-op when l carries no command.
//
// prec values below 0 are clamped to 0; values above 32 are clamped to
// 32 (the maximum FormatFloat accepts).
func (l Line) ArgFP(key string, prec int, v float64) Line {
	if !l.HasCommand {
		return l
	}
	if prec < 0 {
		prec = 0
	}
	if prec > 32 {
		prec = 32
	}
	return l.appendArg(Argument{
		Key:       key,
		Raw:       strconv.FormatFloat(v, 'f', prec, 64),
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
