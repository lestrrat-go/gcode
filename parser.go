package gcode

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/lestrrat-go/gcode/internal/scan"
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
			var d *Dialect
			_ = opt.Value(&d)
			p.dialect = d
		case identStrict{}:
			_ = opt.Value(&p.strict)
		}
	}
	return p
}

// Parse reads G-code from src and returns an immutable Program.
// In strict mode with a dialect, unknown commands produce an error.
func (p *Parser) Parse(src io.Reader) (*Program, error) {
	sc := scan.NewScanner(src)
	pb := NewProgramBuilder()

	for sc.Scan() {
		line, err := scan.ParseLine(sc.Text(), sc.LineNum())
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
				return nil, NewParseError(sc.LineNum(), 1, sc.Text(),
					fmt.Errorf("unknown command %c%d in dialect %s", cmd.Letter, cmd.Number, p.dialect.Name()))
			}
		}

		pb.Append(line)
	}

	if err := sc.Err(); err != nil {
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
