package gcode

// CommandBuilder provides a fluent API for constructing G-code command lines.
// Builders are pointer-based and should not be reused after calling Line.
type CommandBuilder struct {
	cmd        Command
	comment    Comment
	hasComment bool
}

// G creates a CommandBuilder for a G-code command (e.g., G0, G1, G28).
func G(number int) *CommandBuilder {
	return &CommandBuilder{cmd: Command{Letter: 'G', Number: number}}
}

// M creates a CommandBuilder for an M-code command (e.g., M104, M106).
func M(number int) *CommandBuilder {
	return &CommandBuilder{cmd: Command{Letter: 'M', Number: number}}
}

// T creates a CommandBuilder for a tool-select command (e.g., T0, T1).
func T(number int) *CommandBuilder {
	return &CommandBuilder{cmd: Command{Letter: 'T', Number: number}}
}

// Cmd creates a CommandBuilder for an arbitrary command letter and number.
func Cmd(letter byte, number int) *CommandBuilder {
	return &CommandBuilder{cmd: Command{Letter: letter, Number: number}}
}

// GSub creates a CommandBuilder for a G-code with a subcode (e.g., G92.1).
func GSub(number, subcode int) *CommandBuilder {
	return &CommandBuilder{cmd: Command{
		Letter:     'G',
		Number:     number,
		HasSubcode: true,
		Subcode:    subcode,
	}}
}

// X appends an X parameter and returns the builder.
func (b *CommandBuilder) X(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'X', Value: v})
	return b
}

// Y appends a Y parameter and returns the builder.
func (b *CommandBuilder) Y(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'Y', Value: v})
	return b
}

// Z appends a Z parameter and returns the builder.
func (b *CommandBuilder) Z(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'Z', Value: v})
	return b
}

// E appends an E parameter and returns the builder.
func (b *CommandBuilder) E(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'E', Value: v})
	return b
}

// F appends an F parameter and returns the builder.
func (b *CommandBuilder) F(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'F', Value: v})
	return b
}

// S appends an S parameter and returns the builder.
func (b *CommandBuilder) S(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'S', Value: v})
	return b
}

// P appends a P parameter and returns the builder.
func (b *CommandBuilder) P(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'P', Value: v})
	return b
}

// R appends an R parameter and returns the builder.
func (b *CommandBuilder) R(v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: 'R', Value: v})
	return b
}

// Param appends a parameter with the given letter and value. Use this for
// less common parameters such as I and J in arc commands.
func (b *CommandBuilder) Param(letter byte, v float64) *CommandBuilder {
	b.cmd.Params = append(b.cmd.Params, Parameter{Letter: letter, Value: v})
	return b
}

// Comment sets a trailing semicolon-form comment on the line.
func (b *CommandBuilder) Comment(text string) *CommandBuilder {
	b.comment = Comment{Text: text, Form: CommentSemicolon}
	b.hasComment = true
	return b
}

// ParenComment sets a trailing parenthesis-form comment on the line.
func (b *CommandBuilder) ParenComment(text string) *CommandBuilder {
	b.comment = Comment{Text: text, Form: CommentParenthesis}
	b.hasComment = true
	return b
}

// Build produces the final Line value from the builder.
func (b *CommandBuilder) Build() Line {
	return Line{
		HasCommand: true,
		Command:    b.cmd,
		HasComment: b.hasComment,
		Comment:    b.comment,
	}
}
