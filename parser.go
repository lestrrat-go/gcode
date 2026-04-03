package gcode

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Parser holds configuration for a parse operation.
// Use NewParser to create one with the desired options.
type Parser struct {
	dialect *Dialect
	strict  bool
}

// NewParser returns a Parser configured with the given options.
func NewParser(options ...ParseOption) *Parser {
	p := &Parser{}
	for _, opt := range options {
		switch opt.Ident() {
		case identDialect{}:
			var d any
			if err := opt.Value(&d); err == nil {
				if dialect, ok := d.(*Dialect); ok {
					p.dialect = dialect
				}
			}
		case identStrict{}:
			_ = opt.Value(&p.strict)
		}
	}
	return p
}

// Parse reads G-code from src and returns an immutable Program.
// In strict mode with a dialect, unknown commands produce an error.
func (p *Parser) Parse(src io.Reader) (*Program, error) {
	sc := newLineScanner(src)
	pb := NewProgramBuilder()

	for sc.scan() {
		line, err := parseLine(sc.text(), sc.lineNumber())
		if err != nil {
			return nil, err
		}

		if p.strict && p.dialect != nil && line.HasCommand {
			cmd := line.Command
			subcode := 0
			if cmd.HasSubcode {
				subcode = cmd.Subcode
			}
			if _, ok := p.dialect.LookupCommand(cmd.Letter, cmd.Number, subcode); !ok {
				return nil, NewParseError(sc.lineNumber(), 1, sc.text(),
					fmt.Errorf("unknown command %c%d in dialect %s", cmd.Letter, cmd.Number, p.dialect.Name()))
			}
		}

		pb.Append(line)
	}

	if err := sc.err(); err != nil {
		return nil, err
	}

	return pb.Build(), nil
}

// ParseString parses G-code from a string and returns an immutable Program.
func (p *Parser) ParseString(src string) (*Program, error) {
	return p.Parse(strings.NewReader(src))
}

// ParseBytes parses G-code from a byte slice and returns an immutable Program.
func (p *Parser) ParseBytes(src []byte) (*Program, error) {
	return p.Parse(bytes.NewReader(src))
}
