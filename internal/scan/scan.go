// Package scan provides a low-level line scanner and line parser for G-code.
// It is an internal package used by the parent module's parser.
package scan

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	gcode "github.com/lestrrat-go/gcode"
)

// Scanner wraps bufio.Scanner with line number tracking.
type Scanner struct {
	scanner *bufio.Scanner
	lineNum int
}

// NewScanner returns a new Scanner that reads from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan advances to the next line and returns true if a line was read.
func (s *Scanner) Scan() bool {
	if s.scanner.Scan() {
		s.lineNum++
		return true
	}
	return false
}

// Text returns the text of the current line.
func (s *Scanner) Text() string {
	return s.scanner.Text()
}

// LineNum returns the 1-based line number of the current line.
func (s *Scanner) LineNum() int {
	return s.lineNum
}

// Err returns the first non-EOF error encountered by the scanner.
func (s *Scanner) Err() error {
	return s.scanner.Err()
}

// ParseLine parses a single raw line of G-code into a gcode.Line value.
// It does NOT do dialect validation -- that is the Parser's job.
func ParseLine(rawLine string, lineNum int) (gcode.Line, error) {
	line := gcode.Line{Raw: rawLine}
	s := rawLine
	pos := 0 // byte position in rawLine for error reporting

	// 1. Strip leading whitespace
	trimmed := strings.TrimLeftFunc(s, unicode.IsSpace)
	pos += len(s) - len(trimmed)
	s = trimmed

	// If empty after trimming → blank line
	if len(s) == 0 {
		return line, nil
	}

	// 2. Semicolon comment-only line
	if s[0] == ';' {
		line.HasComment = true
		line.Comment = gcode.Comment{
			Text: s[1:],
			Form: gcode.CommentSemicolon,
		}
		return line, nil
	}

	// 3. Parenthesis comment-only line
	if s[0] == '(' {
		end := strings.IndexByte(s, ')')
		if end == -1 {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s, fmt.Errorf("unclosed parenthesis comment"))
		}
		// Check if the rest of the line (after closing paren) is only whitespace
		rest := strings.TrimSpace(s[end+1:])
		if len(rest) == 0 {
			line.HasComment = true
			line.Comment = gcode.Comment{
				Text: s[1:end],
				Form: gcode.CommentParenthesis,
			}
			return line, nil
		}
		// Not a comment-only line; fall through to command parsing
	}

	// Track where 'N' starts for checksum calculation
	nStart := pos

	// 4. Consume optional N<digits> line number prefix
	if len(s) > 0 && (s[0] == 'N' || s[0] == 'n') {
		i := 1
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == 1 {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("expected digits after N"))
		}
		n, err := strconv.Atoi(s[1:i])
		if err != nil {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:i], fmt.Errorf("invalid line number: %w", err))
		}
		line.LineNumber = n
		pos += i
		s = s[i:]
		// Skip whitespace after line number
		trimmed = strings.TrimLeftFunc(s, unicode.IsSpace)
		pos += len(s) - len(trimmed)
		s = trimmed
	}

	// If nothing left after line number, it's just a line number (unusual but valid blank-ish line)
	if len(s) == 0 {
		return line, nil
	}

	// 5. Consume command token: letter + integer [+ '.' + integer]
	if !isLetter(s[0]) {
		return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("expected command letter"))
	}

	cmdLetter := toUpper(s[0])
	s = s[1:]
	pos++

	// Parse command number (digits required)
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("expected digits after command letter %c", cmdLetter))
	}
	cmdNum, _ := strconv.Atoi(s[:i])
	pos += i
	s = s[i:]

	cmd := gcode.Command{
		Letter: cmdLetter,
		Number: cmdNum,
	}

	// Optional subcode: '.' + digits
	if len(s) > 0 && s[0] == '.' {
		s = s[1:]
		pos++
		j := 0
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if j == 0 {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("expected digits after '.' in subcode"))
		}
		sub, _ := strconv.Atoi(s[:j])
		cmd.HasSubcode = true
		cmd.Subcode = sub
		pos += j
		s = s[j:]
	}

	// 7. Consume zero or more parameter tokens
	var params []gcode.Parameter
	for {
		// Skip whitespace
		trimmed = strings.TrimLeftFunc(s, unicode.IsSpace)
		pos += len(s) - len(trimmed)
		s = trimmed

		if len(s) == 0 {
			break
		}

		// Stop if we hit a comment or checksum
		if s[0] == ';' || s[0] == '(' || s[0] == '*' {
			break
		}

		// Must be a letter for a parameter
		if !isLetter(s[0]) {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("unexpected character %q", s[0]))
		}

		paramLetter := toUpper(s[0])
		s = s[1:]
		pos++

		// Try to parse numeric value; if next char is not numeric, value is 0 (flag-style)
		var value float64
		if len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == '.' || (s[0] >= '0' && s[0] <= '9')) {
			numStr, numLen := consumeNumber(s)
			if numLen == 0 {
				return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("invalid number after parameter %c", paramLetter))
			}
			v, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, numStr, fmt.Errorf("invalid parameter value: %w", err))
			}
			value = v
			pos += numLen
			s = s[numLen:]
		}
		// else: flag-style parameter, value stays 0

		params = append(params, gcode.Parameter{Letter: paramLetter, Value: value})
	}
	cmd.Params = params
	line.HasCommand = true
	line.Command = cmd

	// Skip whitespace
	trimmed = strings.TrimLeftFunc(s, unicode.IsSpace)
	pos += len(s) - len(trimmed)
	s = trimmed

	// 8. Consume optional (...) comment after all params
	var parenComment string
	hasParenComment := false
	if len(s) > 0 && s[0] == '(' {
		end := strings.IndexByte(s, ')')
		if end == -1 {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("unclosed parenthesis comment"))
		}
		parenComment = s[1:end]
		hasParenComment = true
		pos += end + 1
		s = s[end+1:]

		// Skip whitespace after paren comment
		trimmed = strings.TrimLeftFunc(s, unicode.IsSpace)
		pos += len(s) - len(trimmed)
		s = trimmed
	}

	// 9. Consume optional ; comment at end
	var semiComment string
	hasSemiComment := false
	if len(s) > 0 && s[0] == ';' {
		semiComment = s[1:]
		hasSemiComment = true
		s = ""
	}

	// 10. If both comments: keep semicolon, discard parenthesis
	if hasSemiComment {
		line.HasComment = true
		line.Comment = gcode.Comment{
			Text: semiComment,
			Form: gcode.CommentSemicolon,
		}
	} else if hasParenComment {
		line.HasComment = true
		line.Comment = gcode.Comment{
			Text: parenComment,
			Form: gcode.CommentParenthesis,
		}
	}

	// Skip whitespace before potential checksum
	trimmed = strings.TrimLeftFunc(s, unicode.IsSpace)
	pos += len(s) - len(trimmed)
	s = trimmed

	// 11. Consume optional *<int> checksum
	if len(s) > 0 && s[0] == '*' {
		s = s[1:]
		pos++
		j := 0
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if j == 0 {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:min(len(s), 10)], fmt.Errorf("expected digits after '*' for checksum"))
		}
		csVal, err := strconv.Atoi(s[:j])
		if err != nil || csVal < 0 || csVal > 255 {
			return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, s[:j], fmt.Errorf("invalid checksum value"))
		}
		line.HasChecksum = true
		line.Checksum = byte(csVal)
		s = s[j:]
	}

	// Compute checksum range: from 'N' (or first non-whitespace if no N) to '*'
	// The checksum in the line should match XOR of bytes from nStart to '*'
	_ = nStart // used for checksum verification if needed in future

	// Remaining should be whitespace only
	rest := strings.TrimSpace(s)
	if len(rest) > 0 {
		return gcode.Line{}, gcode.NewParseError(lineNum, pos+1, rest[:min(len(rest), 10)], fmt.Errorf("unexpected trailing content"))
	}

	return line, nil
}

// consumeNumber reads a numeric literal (integer or float, optionally negative)
// from the start of s and returns the string and its length in bytes.
func consumeNumber(s string) (string, int) {
	i := 0
	// Optional sign
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		i++
	}
	// Integer part
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	// Optional fractional part
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	// Must have consumed at least one digit
	if i == start {
		// Only a sign (or nothing), no digits
		return "", 0
	}
	return s[:i], i
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func toUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}
