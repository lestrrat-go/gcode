package gcode

// BlankLine returns a blank Line (IsBlank returns true).
func BlankLine() Line {
	return Line{}
}

// CommentLine returns a comment-only Line with semicolon form.
func CommentLine(text string) Line {
	return Line{
		HasComment: true,
		Comment:    Comment{Text: text, Form: CommentSemicolon},
	}
}

// ParenCommentLine returns a comment-only Line with parenthesis form.
func ParenCommentLine(text string) Line {
	return Line{
		HasComment: true,
		Comment:    Comment{Text: text, Form: CommentParenthesis},
	}
}
