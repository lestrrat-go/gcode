package gcode

import (
	"bufio"
	"io"
	"strconv"

	"github.com/lestrrat-go/option/v3"
)

// Writer streams G-code lines to an [io.Writer]. It is buffered;
// callers must invoke [Writer.Flush] (or close-equivalent) when done
// to ensure all output is written.
type Writer struct {
	bw              *bufio.Writer
	emitComments    bool
	emitLineNumbers bool
	computeChecksum bool
	lineEnding      LineEnding
}

// NewWriter returns a Writer that emits G-code lines to w.
func NewWriter(w io.Writer, opts ...WriteOption) *Writer {
	wr := &Writer{
		bw:           bufio.NewWriter(w),
		emitComments: true,
		lineEnding:   LineEndingLF,
	}
	for _, opt := range opts {
		switch opt.Ident() {
		case identEmitComments{}:
			wr.emitComments = option.MustGet[bool](opt)
		case identEmitLineNumbers{}:
			wr.emitLineNumbers = option.MustGet[bool](opt)
		case identComputeChecksum{}:
			wr.computeChecksum = option.MustGet[bool](opt)
		case identLineEnding{}:
			wr.lineEnding = option.MustGet[LineEnding](opt)
		}
	}
	return wr
}

// Write emits a single Line followed by the configured line ending.
// A blank line emits only the line ending.
func (w *Writer) Write(line Line) error {
	body := w.formatBody(line)
	if w.computeChecksum && (line.HasCommand || line.LineNumber > 0) {
		var cs byte
		for i := range len(body) {
			cs ^= body[i]
		}
		body += "*" + strconv.Itoa(int(cs))
	}
	if _, err := w.bw.WriteString(body); err != nil {
		return err
	}
	switch w.lineEnding {
	case LineEndingCRLF:
		_, err := w.bw.WriteString("\r\n")
		return err
	default:
		return w.bw.WriteByte('\n')
	}
}

// Flush flushes any buffered output to the underlying writer.
func (w *Writer) Flush() error { return w.bw.Flush() }

func (w *Writer) formatBody(line Line) string {
	if line.IsBlank() {
		return ""
	}

	var sb stringBuilder

	// Comment-only line.
	if !line.HasCommand {
		if w.emitLineNumbers && line.LineNumber > 0 {
			sb.writeByte('N')
			sb.writeString(strconv.Itoa(line.LineNumber))
			sb.writeByte(' ')
		}
		if line.HasComment && w.emitComments {
			writeComment(&sb, line.Comment, false)
		}
		return sb.String()
	}

	if w.emitLineNumbers && line.LineNumber > 0 {
		sb.writeByte('N')
		sb.writeString(strconv.Itoa(line.LineNumber))
		sb.writeByte(' ')
	}

	sb.writeString(line.Command.Name)

	classic := isClassicName(line.Command.Name)
	for _, a := range line.Command.Args {
		if classic {
			sb.writeByte(' ')
			sb.writeString(a.Key)
			sb.writeString(a.Raw)
		} else {
			sb.writeByte(' ')
			sb.writeString(a.Key)
			if !a.IsFlag() {
				sb.writeByte('=')
				sb.writeString(a.Raw)
			}
		}
	}

	if line.HasComment && w.emitComments {
		writeComment(&sb, line.Comment, true)
	}

	return sb.String()
}

func writeComment(sb *stringBuilder, c Comment, leadingSpace bool) {
	switch c.Form {
	case CommentParenthesis:
		if leadingSpace {
			sb.writeByte(' ')
		}
		sb.writeByte('(')
		sb.writeString(c.Text)
		sb.writeByte(')')
	default: // CommentSemicolon
		if leadingSpace {
			sb.writeByte(' ')
		}
		sb.writeByte(';')
		sb.writeString(c.Text)
	}
}

// isClassicName reports whether name has the form <Letter><digits>
// (with optional .digits) — i.e. a classic G/M/T code.
func isClassicName(name string) bool {
	if len(name) < 2 {
		return false
	}
	if !isLetter(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !isDigit(c) && c != '.' {
			return false
		}
	}
	return true
}

// stringBuilder is a tiny wrapper around a byte slice; using
// strings.Builder here would also work but pulls in extra method
// overhead for what is a small, hot path.
type stringBuilder struct{ b []byte }

func (s *stringBuilder) writeByte(c byte)       { s.b = append(s.b, c) }
func (s *stringBuilder) writeString(str string) { s.b = append(s.b, str...) }
func (s *stringBuilder) String() string         { return string(s.b) }
