package gcode

import (
	"io"
	"strconv"
	"strings"
)

// Formatter holds configuration for a generate operation.
// Use NewFormatter to create one with the desired options.
type Formatter struct {
	emitComments    bool
	emitLineNumbers bool
	computeChecksum bool
	lineEnding      LineEnding
}

// NewFormatter returns a Formatter configured with the given options.
// Defaults: emitComments=true, emitLineNumbers=false,
// computeChecksum=false, lineEnding=LineEndingLF.
func NewFormatter(options ...FormatOption) *Formatter {
	g := &Formatter{
		emitComments: true,
		lineEnding:   LineEndingLF,
	}
	for _, opt := range options {
		switch opt.Ident() {
		case identEmitComments{}:
			_ = opt.Value(&g.emitComments)
		case identEmitLineNumbers{}:
			_ = opt.Value(&g.emitLineNumbers)
		case identComputeChecksum{}:
			_ = opt.Value(&g.computeChecksum)
		case identLineEnding{}:
			_ = opt.Value(&g.lineEnding)
		}
	}
	return g
}

// FormatLine writes a single Line to w WITHOUT a trailing line ending.
// Callers using FormatLine directly must append their own line endings.
func (g *Formatter) FormatLine(w io.Writer, line Line) error {
	if line.IsBlank() {
		return nil
	}

	var sb strings.Builder

	// Comment-only line (no command)
	if !line.HasCommand {
		if line.HasComment && g.emitComments {
			switch line.Comment.Form {
			case CommentSemicolon:
				sb.WriteByte(';')
				sb.WriteString(line.Comment.Text)
			case CommentParenthesis:
				sb.WriteByte('(')
				sb.WriteString(line.Comment.Text)
				sb.WriteByte(')')
			}
		}
		_, err := io.WriteString(w, sb.String())
		return err
	}

	// Line number
	if g.emitLineNumbers && line.LineNumber > 0 {
		sb.WriteByte('N')
		sb.WriteString(strconv.Itoa(line.LineNumber))
		sb.WriteByte(' ')
	}

	// Command letter and number
	sb.WriteByte(line.Command.Letter)
	sb.WriteString(strconv.Itoa(line.Command.Number))

	// Subcode
	if line.Command.HasSubcode {
		sb.WriteByte('.')
		sb.WriteString(strconv.Itoa(line.Command.Subcode))
	}

	// Parameters
	for _, p := range line.Command.Params {
		sb.WriteByte(' ')
		sb.WriteByte(p.Letter)
		sb.WriteString(strconv.FormatFloat(p.Value, 'f', -1, 64))
	}

	// Comment
	if line.HasComment && g.emitComments {
		switch line.Comment.Form {
		case CommentParenthesis:
			sb.WriteString(" (")
			sb.WriteString(line.Comment.Text)
			sb.WriteByte(')')
		case CommentSemicolon:
			sb.WriteString(" ;")
			sb.WriteString(line.Comment.Text)
		}
	}

	// Checksum
	if g.computeChecksum {
		content := sb.String()
		var cs byte
		for i := range len(content) {
			cs ^= content[i]
		}
		sb.WriteByte('*')
		sb.WriteString(strconv.Itoa(int(cs)))
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

// Format writes the G-code representation of prog to w.
// Each line is followed by the configured line ending.
// Blank lines emit only the line ending.
func (g *Formatter) Format(w io.Writer, prog *Program) error {
	var le string
	switch g.lineEnding {
	case LineEndingCRLF:
		le = "\r\n"
	default:
		le = "\n"
	}

	for i := range prog.Len() {
		line := prog.Line(i)
		if err := g.FormatLine(w, line); err != nil {
			return err
		}
		if _, err := io.WriteString(w, le); err != nil {
			return err
		}
	}
	return nil
}
