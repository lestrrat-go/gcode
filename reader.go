package gcode

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"iter"
	"strconv"
	"unsafe"
)

// defaultMaxLineSize bounds the longest single source line the Reader
// will accept. Real-world Klipper extended commands (e.g.
// EXCLUDE_OBJECT_DEFINE with a POLYGON= list of hundreds of vertices)
// can grow well past 64 KiB; 16 MiB is comfortably above any input
// observed in the wild while still bounding pathological cases.
const defaultMaxLineSize = 16 * 1024 * 1024

// Reader streams G-code lines from an [io.Reader].
//
// Each call to [Reader.Read] reuses the Reader's internal buffers, so
// the strings and slices on the returned [Line] remain valid only
// until the next call. Use [Line.Clone] to retain a Line beyond that
// point.
type Reader struct {
	sc      *bufio.Scanner
	buf     []byte
	lineNum int
	args    []Argument
	dialect *Dialect
	strict  bool
}

// NewReader returns a Reader that decodes G-code lines from r.
func NewReader(r io.Reader, opts ...ReadOption) *Reader {
	rd := &Reader{
		sc: bufio.NewScanner(r),
	}
	maxSize := defaultMaxLineSize
	for _, opt := range opts {
		switch opt.Ident() {
		case identDialect{}:
			var v any
			if err := opt.Value(&v); err == nil {
				if d, ok := v.(*Dialect); ok {
					rd.dialect = d
				}
			}
		case identStrict{}:
			_ = opt.Value(&rd.strict)
		case identMaxLineSize{}:
			_ = opt.Value(&maxSize)
		}
	}
	rd.sc.Buffer(make([]byte, 4096), maxSize)
	return rd
}

// Read decodes the next line into *line. It returns [io.EOF] when no
// more lines are available.
//
// The fields of *line — including the strings and the Args slice —
// share storage with the Reader's internal buffers and are
// invalidated by the next call to Read. To retain a Line, call
// [Line.Clone].
func (r *Reader) Read(line *Line) error {
	line.reset()
	if !r.sc.Scan() {
		if err := r.sc.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	r.lineNum++

	src := r.sc.Bytes()
	r.buf = append(r.buf[:0], src...)
	line.Raw = bytesToString(r.buf)

	// Reuse the Args slab.
	line.Command.Args = r.args[:0]

	if err := r.parseInto(line); err != nil {
		return err
	}
	r.args = line.Command.Args

	if r.strict && r.dialect != nil && line.HasCommand {
		if _, ok := r.dialect.LookupCommand(line.Command.Name); !ok {
			return makeParseError(r.lineNum, 1, line.Raw,
				fmt.Errorf("unknown command %s in dialect %s", line.Command.Name, r.dialect.Name()))
		}
	}
	return nil
}

// All returns a range-iterator that calls Read repeatedly. It yields
// each Line and any error from the underlying Read; iteration stops
// at the first error or at io.EOF (which is not yielded).
//
// Like Read, the yielded Line is invalidated when the iterator
// advances. Call [Line.Clone] inside the loop to retain.
func (r *Reader) All() iter.Seq2[Line, error] {
	return func(yield func(Line, error) bool) {
		var line Line
		for {
			err := r.Read(&line)
			if err == io.EOF {
				return
			}
			if !yield(line, err) {
				return
			}
			if err != nil {
				return
			}
		}
	}
}

// parseInto walks r.buf and populates *line.
func (r *Reader) parseInto(line *Line) error {
	pos := skipSpace(r.buf, 0)
	if pos >= len(r.buf) {
		return nil // blank line
	}

	// Comment-only line (semicolon).
	if r.buf[pos] == ';' {
		line.HasComment = true
		line.Comment.Form = CommentSemicolon
		line.Comment.Text = bytesToString(r.buf[pos+1:])
		return nil
	}

	// Comment-only line (parenthesis), only when nothing structured follows.
	if r.buf[pos] == '(' {
		end := bytes.IndexByte(r.buf[pos:], ')')
		if end < 0 {
			return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
				fmt.Errorf("unclosed parenthesis comment"))
		}
		end += pos
		rest := skipSpace(r.buf, end+1)
		if rest >= len(r.buf) {
			line.HasComment = true
			line.Comment.Form = CommentParenthesis
			line.Comment.Text = bytesToString(r.buf[pos+1 : end])
			return nil
		}
		// Otherwise fall through: leading "(...)" is a quirk we won't
		// special-case beyond the comment-only form.
	}

	// Optional N<digits> line number.
	if pos < len(r.buf) {
		c := r.buf[pos]
		if c == 'N' || c == 'n' {
			i := pos + 1
			for i < len(r.buf) && isDigit(r.buf[i]) {
				i++
			}
			if i == pos+1 {
				return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
					fmt.Errorf("expected digits after N"))
			}
			n, err := strconv.Atoi(string(r.buf[pos+1 : i]))
			if err != nil {
				return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
					fmt.Errorf("invalid line number: %w", err))
			}
			line.LineNumber = n
			pos = skipSpace(r.buf, i)
		}
	}

	if pos >= len(r.buf) {
		return nil
	}

	// Command.
	if !isLetter(r.buf[pos]) && r.buf[pos] != '_' {
		return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
			fmt.Errorf("expected command letter"))
	}

	// Distinguish classic vs extended by the byte after the first letter.
	classic := pos+1 < len(r.buf) && isDigit(r.buf[pos+1])

	var cmdEnd int
	if classic {
		// Letter + digits + optional .digits
		i := pos + 1
		for i < len(r.buf) && isDigit(r.buf[i]) {
			i++
		}
		if i < len(r.buf) && r.buf[i] == '.' {
			j := i + 1
			for j < len(r.buf) && isDigit(r.buf[j]) {
				j++
			}
			if j == i+1 {
				return makeParseError(r.lineNum, i+1, snippet(r.buf, i),
					fmt.Errorf("expected digits after '.' in subcode"))
			}
			i = j
		}
		cmdEnd = i
		upperASCII(r.buf[pos:pos+1])
	} else {
		// Extended identifier: [A-Za-z_][A-Za-z0-9_]*
		i := pos + 1
		for i < len(r.buf) && isIdentByte(r.buf[i]) {
			i++
		}
		cmdEnd = i
		upperASCII(r.buf[pos:cmdEnd])
	}

	line.HasCommand = true
	line.Command.Name = bytesToString(r.buf[pos:cmdEnd])
	pos = cmdEnd

	// Arguments.
	for {
		pos = skipSpace(r.buf, pos)
		if pos >= len(r.buf) {
			break
		}
		c := r.buf[pos]
		if c == ';' || c == '(' || c == '*' {
			break
		}

		if classic {
			// Free-form-tail tolerance: any non-letter at arg position
			// (e.g. M117 message text, M118 // comment-style payload)
			// signals a string-arg command. Bail cleanly; the command
			// stands as command-only and Line.Raw preserves the source.
			if !isLetter(c) {
				line.Command.Args = line.Command.Args[:0]
				pos = len(r.buf)
				break
			}
			// Single-letter classic key.
			keyStart := pos
			upperASCII(r.buf[keyStart : keyStart+1])
			pos++

			// Same tolerance: classic key followed by another letter is
			// a free-form payload (e.g. "M117 Hello").
			if pos < len(r.buf) && isLetter(r.buf[pos]) {
				line.Command.Args = line.Command.Args[:0]
				pos = len(r.buf)
				break
			}

			arg := Argument{Key: bytesToString(r.buf[keyStart : keyStart+1])}
			if pos < len(r.buf) && isNumberStart(r.buf[pos]) {
				numStart := pos
				numLen := scanNumber(r.buf[pos:])
				if numLen == 0 {
					return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
						fmt.Errorf("invalid number after parameter %c", c))
				}
				pos += numLen
				arg.Raw = bytesToString(r.buf[numStart:pos])
				if v, err := strconv.ParseFloat(arg.Raw, 64); err == nil {
					arg.Number = v
					arg.IsNumeric = true
				}
			}
			line.Command.Args = append(line.Command.Args, arg)
		} else {
			// Extended: identifier '=' value, or bare identifier.
			// Preserve source case on extended arg keys — emitters
			// disagree on convention and round-trip fidelity matters.
			if !isLetter(c) && c != '_' {
				return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
					fmt.Errorf("unexpected character %q in extended argument", c))
			}
			keyStart := pos
			i := pos + 1
			for i < len(r.buf) && isIdentByte(r.buf[i]) {
				i++
			}
			arg := Argument{Key: bytesToString(r.buf[keyStart:i])}
			pos = i

			if pos < len(r.buf) && r.buf[pos] == '=' {
				pos++
				valStart := pos
				valLen := scanExtendedValue(r.buf[pos:])
				pos += valLen
				arg.Raw = bytesToString(r.buf[valStart:pos])
				if arg.Raw != "" {
					if v, err := strconv.ParseFloat(arg.Raw, 64); err == nil {
						arg.Number = v
						arg.IsNumeric = true
					}
				}
			}
			line.Command.Args = append(line.Command.Args, arg)
		}
	}

	pos = skipSpace(r.buf, pos)

	// Optional trailing ( ... ) comment.
	parenStart := -1
	parenEnd := -1
	if pos < len(r.buf) && r.buf[pos] == '(' {
		end := bytes.IndexByte(r.buf[pos:], ')')
		if end < 0 {
			return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
				fmt.Errorf("unclosed parenthesis comment"))
		}
		parenStart = pos + 1
		parenEnd = pos + end
		pos = pos + end + 1
		pos = skipSpace(r.buf, pos)
	}

	// Optional trailing ; comment.
	if pos < len(r.buf) && r.buf[pos] == ';' {
		line.HasComment = true
		line.Comment.Form = CommentSemicolon
		line.Comment.Text = bytesToString(r.buf[pos+1:])
		return nil
	}

	if parenStart >= 0 {
		line.HasComment = true
		line.Comment.Form = CommentParenthesis
		line.Comment.Text = bytesToString(r.buf[parenStart:parenEnd])
	}

	// Optional *<digits> checksum.
	if pos < len(r.buf) && r.buf[pos] == '*' {
		i := pos + 1
		for i < len(r.buf) && isDigit(r.buf[i]) {
			i++
		}
		if i == pos+1 {
			return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
				fmt.Errorf("expected digits after '*' for checksum"))
		}
		v, err := strconv.Atoi(string(r.buf[pos+1 : i]))
		if err != nil || v < 0 || v > 255 {
			return makeParseError(r.lineNum, pos+1, snippet(r.buf, pos),
				fmt.Errorf("invalid checksum value"))
		}
		line.Checksum = byte(v)
		line.HasChecksum = true
		pos = i
	}

	// Trailing whitespace OK; anything else is unexpected.
	if rest := skipSpace(r.buf, pos); rest < len(r.buf) {
		return makeParseError(r.lineNum, rest+1, snippet(r.buf, rest),
			fmt.Errorf("unexpected trailing content"))
	}
	return nil
}

// scanExtendedValue returns the byte length of an extended-command
// argument value beginning at b[0]. It reads until whitespace, ';',
// or '*' at depth 0 — but balances [], (), and {} pairs and treats
// double- and single-quoted strings as opaque tokens.
func scanExtendedValue(b []byte) int {
	i := 0
	depth := 0
	for i < len(b) {
		c := b[i]
		if depth == 0 && (c == ' ' || c == '\t' || c == ';' || c == '*') {
			return i
		}
		switch c {
		case '[', '(', '{':
			depth++
		case ']', ')', '}':
			if depth > 0 {
				depth--
			}
		case '"', '\'':
			j := i + 1
			for j < len(b) && b[j] != c {
				j++
			}
			if j < len(b) {
				i = j
			} else {
				i = len(b) - 1
			}
		}
		i++
	}
	return i
}

func skipSpace(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\t') {
		i++
	}
	return i
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func isIdentByte(b byte) bool { return isLetter(b) || isDigit(b) || b == '_' }

func isNumberStart(b byte) bool {
	return b == '-' || b == '+' || b == '.' || isDigit(b)
}

// scanNumber returns the byte length of a decimal number at b[0].
// Accepts optional sign, integer digits, optional decimal point with
// fractional digits. No exponent — G-code numbers don't use them.
func scanNumber(b []byte) int {
	i := 0
	if i < len(b) && (b[i] == '-' || b[i] == '+') {
		i++
	}
	start := i
	for i < len(b) && isDigit(b[i]) {
		i++
	}
	if i < len(b) && b[i] == '.' {
		i++
		for i < len(b) && isDigit(b[i]) {
			i++
		}
	}
	if i == start {
		return 0
	}
	return i
}

func upperASCII(b []byte) {
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
}

// snippet returns a short excerpt of b starting at pos, suitable for
// error messages.
func snippet(b []byte, pos int) string {
	end := pos + 10
	if end > len(b) {
		end = len(b)
	}
	return string(b[pos:end])
}

// bytesToString returns s as a string sharing b's backing storage.
// The returned string is valid only as long as b is not mutated.
func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
