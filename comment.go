package gcode

import "fmt"

// CommentForm distinguishes the syntactic form of a G-code comment
// for round-trip fidelity.
type CommentForm int

const (
	// CommentSemicolon represents a semicolon-delimited comment (; text).
	CommentSemicolon CommentForm = iota
	// CommentParenthesis represents a parenthesis-delimited comment ((text)).
	CommentParenthesis
)

// String returns the human-readable name of the comment form.
func (f CommentForm) String() string {
	switch f {
	case CommentSemicolon:
		return "semicolon"
	case CommentParenthesis:
		return "parenthesis"
	default:
		return fmt.Sprintf("CommentForm(%d)", int(f))
	}
}

// Comment carries the raw comment text (without delimiter characters).
// The Form field distinguishes ; vs (...) syntax for round-trip fidelity.
type Comment struct {
	// Text is the comment content without delimiters.
	Text string
	// Form indicates the syntactic style of the comment.
	Form CommentForm
}
