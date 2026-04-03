# Design Document: github.com/lestrrat-go/gcode

## 1. Executive Summary

`github.com/lestrrat-go/gcode` is a Go library for parsing G-code text into a structured in-memory representation and generating G-code text from that representation. G-code is the numerical control programming language used by CNC machines, 3D printers, and laser cutters.

The library targets the Marlin and RepRap firmware dialects initially, with an extensible dialect system that permits downstream users and packages to register additional dialects. It also exposes a macro expansion mechanism that lets users define named command sequences, which is useful for high-level toolpath generation.

Key success criteria:

- Round-trip fidelity: parsing and re-generating the same G-code file produces equivalent output (comments and whitespace handling is configurable, not silent-drop).
- Correct concurrent use: the registry of dialects/macros is safe to read from multiple goroutines; mutation requires explicit synchronisation that the caller controls.
- Actionable errors: every parse error carries the source line number, column, and the offending text.
- Extensibility without forking: users can add dialects, command definitions, and macros without modifying library source.

---

## 2. Requirements Analysis

### 2.1 Functional Requirements

1. Parse G-code text (from `string`, `[]byte`, or `io.Reader`) into a `Program` — an ordered slice of `Line` values.
2. Each `Line` may represent: a command (G/M/T code), a comment, a blank line, or a line containing both a command and a trailing comment.
3. Each command carries: letter (`G`, `M`, `T`, …), number (integer part), optional subcode (e.g., `G92.1`), and zero or more `Parameter` values (letter + float64 value).
4. Support optional line-number prefix (`N<int>`) and optional checksum suffix (`*<byte>`).
5. Generate G-code text from a `Program`, with options to control: comment output, line-number output, checksum generation, and line-ending style (LF vs CRLF).
6. Dialect definitions specify which commands and parameters are valid; the parser uses a dialect to provide richer errors and to expose dialect-specific metadata on parsed nodes.
7. Dialects may be composed: a child dialect inherits from a parent, adding or overriding command definitions.
8. Users can register custom command definitions against a dialect at run time.
9. Users can register named macros that expand into a `[]Line` slice; macro expansion is explicit (not automatic during parse).
10. The package must be usable with Go modules (module path `github.com/lestrrat-go/gcode`).

### 2.2 Non-Functional Requirements

- **Concurrency**: reading a `Program`, `Dialect`, or `MacroRegistry` is safe from multiple goroutines without external locks. Writing (registering a new command definition or macro) is not concurrent-safe; callers must synchronise.
- **Memory**: the parser must not retain references to the original input buffer after returning.
- **Errors**: parse errors implement `error` and carry line number, column, and context.
- **Dependencies**: minimise external dependencies. `github.com/lestrrat-go/option/v2` may be used for the options pattern consistent with the rest of the lestrrat-go ecosystem.

### 2.3 Constraints and Assumptions

- G-code numbers are parsed as `float64` parameter values and `int` command numbers (no arbitrary precision needed).
- Comments are lines or inline suffixes beginning with `;` or enclosed in `(…)`. Both forms are supported.
- Subcodes (e.g., `G92.1`) are represented as a separate `int` field; the dot is not part of the command number.
- The checksum byte is an XOR of all bytes on the line starting from (and including) the `N` character up to (but not including) the `*` character, matching Marlin/RepRap convention. Whitespace before the `N` is excluded.
- Line numbers (`N`) are positive integers; the parser does not enforce ordering.
- This version does not evaluate or execute G-code; it only parses and generates text.
- Macro expansion does not recurse (macros cannot reference other macros).

### 2.4 Out-of-Scope

- Simulation or toolpath visualisation.
- Binary G-code formats (e.g., Klipper binary).
- Recursive macro expansion.
- Network streaming of G-code.
- Any UI or CLI tool (pure library only).
- Multiple commands per line: each input line is parsed as at most one command. Files with multiple commands on a single line (e.g., `G92 E0 G28 X Y`) produce a parse error in strict mode; in non-strict mode only the first command is parsed and the remainder is ignored.
- Expression syntax and meta-commands: Marlin 2.x conditional/expression syntax (`if`, `while`, `{var.x + 10}`) and RepRapFirmware 3+ meta-commands (`if sensors.filament.runout`, `set`, `var`) are not supported. The parser targets classic (pre-expression) G-code only.
- String-valued parameters: commands like `M23 filename.gco` or `M117 Hello World` that use non-address-value-pair arguments are parsed as command-only (the raw text is preserved in `Line.Raw` but not structured into `Parameter` values).

---

## 3. Technical Design

### 3.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Public API surface                       │
│                                                                 │
│   gcode.Parse(…)      gcode.Generate(…)      MacroRegistry     │
│        │                     │                     │           │
│        ▼                     ▼                     ▼           │
│   Parser ──────────► Program ◄──────── Generator              │
│   (internal)          (model)          (internal)              │
│        │                                                        │
│        ▼                                                        │
│   Dialect ◄── DialectRegistry                                  │
│   (command + param defs)                                        │
└─────────────────────────────────────────────────────────────────┘
```

Data flows:

- **Parse path**: raw text → `Parser` reads byte by byte via `strcursor`-style cursor → builds `Line` values → assembles into `Program`.
- **Generate path**: `Program` (or individual `Line` values) → `Generator` formats each line to `io.Writer`.
- **Macro path**: caller invokes `MacroRegistry.Expand(name, params)` → returns `[]Line` → caller inserts into or appends to a `Program`.

### 3.2 Detailed Component Specifications

#### 3.2.1 Package Layout

```
github.com/lestrrat-go/gcode/
├── gcode.go              # package doc, top-level Parse/Generate functions
├── program.go            # Program, Line types
├── command.go            # Command, Parameter types
├── comment.go            # Comment type
├── errors.go             # parseError (unexported), ParseErrorDetail interface, ErrParse sentinel
├── parser.go             # Parser type and implementation
├── generator.go          # Generator type and implementation
├── dialect.go            # Dialect, CommandDef, ParamDef types
├── dialect_registry.go   # DialectRegistry, built-in dialect registration
├── macro.go              # Macro, MacroRegistry types
├── options.go            # shared Option type, ParseOption, GenerateOption interfaces
├── parse_options.go      # parse-specific option constructors
├── generate_options.go   # generate-specific option constructors
├── internal/
│   └── scan/
│       ├── scan.go       # low-level line scanner / cursor helpers
│       └── scan_test.go  # unit tests for parseLine
├── dialects/
│   ├── marlin/
│   │   └── marlin.go     # Marlin dialect definition (command table)
│   └── reprap/
│       └── reprap.go     # RepRap dialect definition
└── example_test.go       # runnable examples in package gcode_test
```

All public types live in the root `gcode` package. Sub-packages `dialects/marlin` and `dialects/reprap` each export a single `Dialect()` function returning a `*gcode.Dialect` ready for registration. This keeps the root package import clean.

#### 3.2.2 Core Model Types

```go
// Program is an ordered sequence of lines that make up a G-code file or stream.
// It is not safe to modify concurrently.
type Program struct {
    lines []Line
}

// Line represents one logical line of G-code. It is one of:
//   - a blank line (IsBlank() == true)
//   - a comment-only line (HasComment && !HasCommand)
//   - a command line (HasCommand)
//   - a command line with trailing comment (HasCommand && HasComment)
//
// LineNumber and Checksum are optional; zero values mean absent.
// Line is a pure value type — no pointer fields. Copying a Line is safe
// and the copy is fully independent.
type Line struct {
    LineNumber  int         // N field value, 0 means absent
    Command     Command     // valid only when HasCommand is true
    HasCommand  bool
    Comment     Comment     // valid only when HasComment is true
    HasComment  bool
    Checksum    byte        // computed checksum; valid only if HasChecksum is true
    HasChecksum bool
    Raw         string      // original source text, populated by parser; useful for
                            // commands with string arguments (M23, M117) that cannot
                            // be fully structured into Parameter values
}

// IsBlank returns true if the line has no command, no comment, and no line number.
func (l Line) IsBlank() bool {
    return !l.HasCommand && !l.HasComment && l.LineNumber == 0
}

// Command represents a single G/M/T (or other letter) code with parameters.
type Command struct {
    Letter     byte        // 'G', 'M', 'T', etc.
    Number     int         // numeric portion, e.g. 28 for G28
    Subcode    int         // subcode after dot, e.g. 1 for G92.1; 0 means absent
    HasSubcode bool
    Params     []Parameter
}

// Parameter represents a single address-value pair, e.g. X10.5 or S200.
type Parameter struct {
    Letter byte    // 'X', 'Y', 'Z', 'E', 'F', 'S', 'T', etc.
    Value  float64
}

// Comment carries the raw comment text (without delimiter characters).
// The form field distinguishes ; vs (…) syntax for round-trip fidelity.
type Comment struct {
    Text string
    Form CommentForm
}

type CommentForm int

const (
    CommentSemicolon    CommentForm = iota // ; text
    CommentParenthesis                    // (text)
)

// String returns the human-readable name of the comment form.
func (f CommentForm) String() string
```

#### 3.2.3 Dialect System

```go
// Dialect describes a set of valid commands and parameters for a firmware variant.
// A Dialect is immutable after construction: its command map is flattened at
// creation time (copy-on-extend from parent). This means LookupCommand is a
// single map lookup with no locking and no parent-chain traversal.
//
// A zero-value Dialect is permissive: it accepts any command letter/number.
type Dialect struct {
    name     string
    commands map[commandKey]*CommandDef // flattened: includes all inherited commands
}

type commandKey struct {
    Letter  byte
    Number  int
    Subcode int // 0 means no subcode; subcode 0 cannot be distinguished from "absent"
                // in lookup keys. This is an accepted limitation — no known firmware
                // uses subcode 0 as a distinct command variant.
}

// CommandDef describes a single command supported by a dialect.
type CommandDef struct {
    // Letter is the command letter, e.g. 'G' or 'M'.
    Letter  byte
    Number  int
    Subcode int
    HasSubcode bool
    // Description is a human-readable summary (for tooling and error messages).
    Description string
    // Params lists the parameter definitions accepted by this command.
    // An empty slice means no parameters are accepted.
    // A nil slice means parameters are unconstrained (permissive).
    Params      []ParamDef
}

// ParamDef describes a single parameter accepted by a command.
type ParamDef struct {
    Letter   byte
    Required bool
    // Description is a human-readable summary.
    Description string
}

// DialectRegistry is a collection of named dialects.
// It is safe to read concurrently; Register must not be called concurrently
// with reads or other writes.
type DialectRegistry struct {
    mu       sync.RWMutex
    dialects map[string]*Dialect
}
```

Key methods on `Dialect`:

```go
// NewDialect creates a new dialect with no parent (root dialect).
func NewDialect(name string) *Dialect

// Extend creates a new child dialect that inherits all command definitions
// from the receiver. The child's command map is a flattened copy — subsequent
// Register calls on the child do not affect the parent, and vice versa.
func (d *Dialect) Extend(name string) *Dialect

func (d *Dialect) Name() string

// Register adds or overwrites a command definition in this dialect.
// Must be called during dialect construction, before the dialect is shared
// with parsers. Register on a shared dialect is not concurrent-safe.
func (d *Dialect) Register(def CommandDef)

// LookupCommand performs a single map lookup (no parent traversal).
// Safe for concurrent use since the command map is not mutated after construction.
func (d *Dialect) LookupCommand(letter byte, number int, subcode int) (*CommandDef, bool)

// Commands returns all command definitions registered in this dialect
// (including inherited ones). Useful for tooling and documentation generation.
func (d *Dialect) Commands() []CommandDef
```

Since the command map is flattened at `Extend` time, `LookupCommand` is a single `O(1)` map lookup with no locking. `Register` mutates the map and is intended for use only during dialect construction (before sharing with parsers). Dialect sub-packages (`dialects/marlin`, `dialects/reprap`) return fresh instances per call to `Dialect()`, so callers get independent copies they can further customise.

#### 3.2.4 Parser

```go
// Parser holds configuration for a parse operation.
// It follows clone-on-update semantics: option methods return a new Parser.
type Parser struct {
    dialect *Dialect
    strict  bool   // if true, unknown commands are parse errors
}

func NewParser(options ...ParseOption) *Parser

// Parse parses G-code from src and returns a Program.
// On error it returns a non-nil *ParseError with line/column information.
func (p *Parser) Parse(src io.Reader) (*Program, error)

// ParseString is a convenience wrapper around Parse.
func (p *Parser) ParseString(src string) (*Program, error)

// ParseBytes is a convenience wrapper around Parse.
func (p *Parser) ParseBytes(src []byte) (*Program, error)
```

Parse options:

```go
// WithDialect attaches a dialect for validation during parsing or generation.
// Returns a value that satisfies both ParseOption and GenerateOption, so the
// same WithDialect call works in both NewParser and NewGenerator.
func WithDialect(d *Dialect) Option  // Option embeds both ParseOption and GenerateOption

// WithStrict enables strict mode: unknown commands produce errors.
// No bool argument — calling WithStrict() enables strict mode.
func WithStrict() ParseOption
```

`Option` is a shared supertype:

```go
type Option interface {
    ParseOption
    GenerateOption
}
```

This resolves the duplicate-name issue: `WithDialect` returns `Option` which satisfies both marker interfaces. Options specific to only parse or generate (e.g., `WithStrict`, `WithEmitComments`) return their specific interface.

Internal parsing algorithm (within `internal/scan`):

1. Read input line by line (splitting on `\n`, preserving `\r\n` awareness).
2. For each raw line:
   a. Strip leading whitespace; if empty → blank line.
   b. If first non-whitespace is `;` → comment-only line.
   c. Attempt to consume optional `N<digits>` prefix.
   d. Consume command token: letter + integer + optional `.` + integer.
   e. Consume zero or more parameter tokens: letter + number (int or float).
   f. Consume optional inline `;` comment.
   g. Consume optional `(…)` comment only after all parameters (not mid-parameter), before any `;` comment or checksum.
   h. Consume optional `*<byte>` checksum at end of line.
   i. Store the original line text in `Line.Raw`.
3. Build a `Line` value; if dialect is set and strict mode is on, validate the command against the dialect.

Error type (see §4.2 for full definition):

```go
// parseError is unexported. Callers match via errors.Is(err, ErrParse).
type parseError struct {
    Line   int
    Column int
    Text   string  // offending text excerpt (max 40 chars)
    Err    error   // underlying cause
}

func (e *parseError) Error() string
func (e *parseError) Unwrap() error
func (e *parseError) Is(target error) bool

// ErrParse is the sentinel for parse errors. Use errors.Is(err, ErrParse).
var ErrParse = errors.New("gcode: parse error")
```

#### 3.2.5 Generator

```go
// Generator holds configuration for a generate operation.
// It follows clone-on-update semantics.
type Generator struct {
    emitComments    bool
    emitLineNumbers bool
    computeChecksum bool
    lineEnding      LineEnding
    dialect         *Dialect
}

type LineEnding int
const (
    LineEndingLF   LineEnding = iota // default
    LineEndingCRLF
)

func NewGenerator(options ...GenerateOption) *Generator

// Generate writes the G-code representation of prog to w.
// Each line is followed by the configured line ending (LF or CRLF).
// Blank lines emit only the line ending.
func (g *Generator) Generate(w io.Writer, prog *Program) error

// GenerateLine writes a single Line to w WITHOUT a trailing line ending.
// Callers using GenerateLine directly must append their own line endings.
// Generate uses GenerateLine internally and appends the line ending itself.
func (g *Generator) GenerateLine(w io.Writer, line Line) error
```

Generate options:

```go
func WithEmitComments(v bool) GenerateOption      // default true
func WithEmitLineNumbers(v bool) GenerateOption   // default false
func WithComputeChecksum(v bool) GenerateOption   // default false
func WithLineEnding(le LineEnding) GenerateOption  // default LF
// WithDialect is shared — see §3.2.4 for its definition.
```

Output format per line (all tokens separated by single space):

```
[N<num>] <Letter><Number>[.<Subcode>] [<ParamLetter><Value>]... [(comment)] [;<comment>] [*<checksum>]
```

Both parser and generator treat `(…)` comments as appearing after all parameters but before any `;` comment or checksum. The parser does not accept `(…)` comments between parameters — this keeps the model simple and the round-trip fidelity consistent.

#### 3.2.6 Macro System

```go
// Macro defines a named expansion that produces a sequence of Lines.
// The library provides SimpleMacro (fixed lines, no substitution) as a
// convenience. Users who need parameter substitution implement this
// interface themselves — the Expand method gives full control over how
// args are used to construct the returned Lines. The library does not
// provide template markers or expression evaluation; substitution logic
// is entirely the implementer's responsibility.
type Macro interface {
    Name() string
    Expand(args map[string]float64) ([]Line, error)
}

// SimpleMacro is a Macro backed by a fixed slice of Lines.
// It ignores the args parameter and always returns a deep copy of its
// stored lines. Use SimpleMacro for fixed command sequences like
// "preheat PLA" that need no parameter substitution.
type SimpleMacro struct {
    name  string
    lines []Line
}

// MacroRegistry is a collection of named macros.
// Reading is concurrent-safe; Register must not be called concurrently.
type MacroRegistry struct {
    mu     sync.RWMutex
    macros map[string]Macro
}

func NewMacroRegistry() *MacroRegistry
func (r *MacroRegistry) Register(m Macro)
func (r *MacroRegistry) Lookup(name string) (Macro, bool)

// Expand looks up the named macro and calls its Expand method.
// Returns a deep copy of the lines so the caller may mutate them safely.
// Returns an error if the macro is not registered.
func (r *MacroRegistry) Expand(name string, args map[string]float64) ([]Line, error)
```

`Macro` is an interface, not a function callback, consistent with the lestrrat-go design preference for interfaces over function callbacks in public APIs. The interface intentionally does not define a template/substitution mechanism — `SimpleMacro` covers the fixed-sequence case, and custom implementations have full freedom to build lines dynamically from the `args` map.

---

### 3.3 Data Design

There is no database. The in-memory model is described in §3.2.2.

**Immutability contract**: `Line`, `Command`, `Parameter`, and `Comment` are pure value types (no pointer fields). Copying a `Line` produces a fully independent value — the only shared backing memory is the `[]Parameter` slice within `Command`, which is copy-on-write safe because `Parameter` is a small value type. Once a `Program` is built by the parser, its contents are not mutated by any library function. Callers who need to transform a program build a new `Program` by iterating over lines.

**Program construction**:

```go
// ProgramBuilder provides a mutable construction API separate from Program.
// Append copies each Line's Command.Params slice to ensure full independence
// from the caller's data. After Append, mutations to the caller's original
// Parameter slices do not affect the builder's stored lines.
type ProgramBuilder struct {
    lines []Line
}

func NewProgramBuilder() *ProgramBuilder
func (b *ProgramBuilder) Append(lines ...Line) *ProgramBuilder
func (b *ProgramBuilder) Build() *Program
```

`Program` itself is immutable post-construction (only read methods are exported):

```go
func (p *Program) Lines() []Line
func (p *Program) Len() int
func (p *Program) Line(i int) Line
```

---

### 3.4 API Specifications

Top-level convenience functions in the root package:

```go
// Parse parses src using default parser settings and returns a Program.
func Parse(src io.Reader, options ...ParseOption) (*Program, error)

// ParseString parses a string.
func ParseString(src string, options ...ParseOption) (*Program, error)

// ParseBytes parses a byte slice.
func ParseBytes(src []byte, options ...ParseOption) (*Program, error)

// Generate writes prog to w using default generator settings.
func Generate(w io.Writer, prog *Program, options ...GenerateOption) error

// GenerateString returns G-code as a string.
func GenerateString(prog *Program, options ...GenerateOption) (string, error)

// GenerateBytes returns G-code as a byte slice.
func GenerateBytes(prog *Program, options ...GenerateOption) ([]byte, error)
```

These are thin wrappers around `NewParser().Parse(…)` and `NewGenerator().Generate(…)`.

---

### 3.5 State Management

There is no persistent or shared mutable state at the package level. All state is caller-owned:

- `Parser` is immutable after construction (clone-on-write semantics, as seen in `helium.Parser`).
- `Generator` follows the same pattern.
- `Program` is immutable post-construction.
- `Dialect` is immutable after construction. `Register` mutates the command map and must only be called during setup, before the dialect is shared with parsers. `LookupCommand` and `Commands` are lock-free reads on the flattened map.
- `DialectRegistry` and `MacroRegistry` use `sync.RWMutex` internally; `Register` acquires a write lock, `Lookup` acquires a read lock. The mutex in `DialectRegistry` protects only the registry map (name → `*Dialect`). The returned `*Dialect` pointer is safe to use without further locking because `Dialect` is immutable after construction.
- Dialect sub-packages (`dialects/marlin`, `dialects/reprap`) return fresh instances per call to `Dialect()`. Callers may further customise them without affecting other callers.

---

## 4. Implementation Guidelines

### 4.1 Coding Standards Compliance

The following conventions are drawn directly from the lestrrat-go ecosystem and the project's Go style rules:

- No named return values.
- For existence-check sets, use `map[T]struct{}` not `map[T]bool`.
- Early returns from functions; shorter `if` branch first; remove `else` via early return where possible.
- Prefer interfaces over function callbacks for public-facing extension points (e.g., `Macro` is an interface, not `func`).
- Options are typed using the `github.com/lestrrat-go/option/v2` pattern: a private ident struct, a public `With…` constructor, and a typed option interface. Example:

```go
// in options.go
type identDialect struct{}
type identStrict struct{}

type ParseOption interface {
    option.Interface
    parseOption()
}

type GenerateOption interface {
    option.Interface
    generateOption()
}

// Option satisfies both ParseOption and GenerateOption.
// Used by WithDialect and any future option shared across parse and generate.
type Option interface {
    ParseOption
    GenerateOption
}

// parseOption is the carrier for parse-only options.
type parseOption struct {
    option.Interface
}
func (parseOption) parseOption() {}

// generateOption is the carrier for generate-only options.
type generateOption struct {
    option.Interface
}
func (generateOption) generateOption() {}

// sharedOption is the carrier for options satisfying both interfaces.
type sharedOption struct {
    option.Interface
}
func (sharedOption) parseOption()    {}
func (sharedOption) generateOption() {}

// WithDialect returns Option (satisfies both ParseOption and GenerateOption).
func WithDialect(d *Dialect) Option {
    return sharedOption{option.New(identDialect{}, d)}
}

// WithStrict returns ParseOption only. No bool arg — presence enables strict mode.
func WithStrict() ParseOption {
    return parseOption{option.New(identStrict{}, true)}
}
```

- Error types: the error struct is unexported (`parseError`). A sentinel variable `ErrParse` is exposed. Callers match via `errors.Is(err, ErrParse)`. The `Is` method on `*parseError` matches against `ErrParse`. See §4.2 for full definition.
- File organisation: one primary type per file (e.g., `parser.go` for `Parser`, `generator.go` for `Generator`), with the exception of small companion types that have no natural home.
- Tests are in `_test` external package form (`package gcode_test`).
- Tests use `github.com/stretchr/testify/require` only (not `assert`).
- `t.Context()` is used instead of `context.Background()` in tests.

### 4.2 Error Handling

**Parse errors** must include:

- Source line number (1-based).
- Column within the line (1-based, byte position).
- A short excerpt of the offending text (max 40 chars).
- A wrapped underlying error describing the problem.

```go
// ParseErrorDetail is the exported interface callers use with errors.As
// to extract structured information from parse errors.
//
// Usage:
//   var detail gcode.ParseErrorDetail
//   if errors.As(err, &detail) {
//       fmt.Printf("line %d, col %d: near %q\n", detail.Line(), detail.Column(), detail.Text())
//   }
type ParseErrorDetail interface {
    error
    Line() int
    Column() int
    Text() string
}

// parseError is the unexported implementation of ParseErrorDetail.
type parseError struct {
    line   int
    column int
    text   string
    err    error
}

// ErrParse is the sentinel for all parse errors. Use errors.Is(err, ErrParse).
var ErrParse = errors.New("gcode: parse error")

func (e *parseError) Error() string {
    return fmt.Sprintf("gcode: parse error at line %d col %d: %s (near %q)", e.line, e.column, e.err, e.text)
}

func (e *parseError) Unwrap() error { return e.err }

func (e *parseError) Is(target error) bool {
    return target == ErrParse
}

func (e *parseError) Line() int    { return e.line }
func (e *parseError) Column() int  { return e.column }
func (e *parseError) Text() string { return e.text }

// makeParseError is the private constructor.
func makeParseError(line, col int, text string, err error) *parseError
```

**Generate errors** are wrapped with context indicating which line index failed.

**Dialect validation errors** wrap `ErrParse` and additionally carry a `commandKey` identifying the unrecognised command.

Logging: this library does not use a logger. Errors are returned; callers log as appropriate.

### 4.3 Testing Strategy

- **Unit tests** for each public type: `Parser`, `Generator`, `Dialect`, `MacroRegistry`.
- **Table-driven tests** are the standard pattern. Each table row has a name, input, expected output/error.
- **Round-trip test**: parse a corpus of G-code snippets, generate, compare string output.
- **Dialect tests**: verify `LookupCommand` finds own and inherited (flattened) commands; verify `Register` overwrites definitions; verify `Extend` produces independent copy.
- **Macro tests**: verify `Expand` returns correct lines; verify unknown macro returns error.
- **Checksum tests**: verify XOR computation matches known-good G-code with `*` checksums.
- **Error tests**: verify `errors.Is(err, ErrParse)` matching; verify `errors.As` extracts line/column.
- **Coverage expectation**: 80 %+ statement coverage on the `gcode` package; dialect sub-packages are data-only and covered implicitly.
- **Test files**:
  - `internal/scan/scan_test.go` — internal scanner/tokeniser tests
  - `parser_test.go` — parsing tests
  - `generator_test.go` — generation tests
  - `dialect_test.go` — dialect system tests
  - `macro_test.go` — macro system tests
  - `roundtrip_test.go` — round-trip corpus tests
  - `testdata/` — sample `.gcode` files for corpus tests

---

## 5. Detailed Task Breakdown

### 5.1 Execution Model

Tasks are executed by **subagents** running in isolated **git worktrees**. Each subagent works on its own branch, and its results are merged back to `main` upon completion. This enables maximum parallelism — independent tasks run concurrently.

**Rules:**
- Each subagent creates a worktree via `git worktree add`, works on a feature branch, and merges back to `main` when done.
- Tasks within the same wave run as **parallel subagents**. A wave does not start until all tasks in the previous wave have merged.
- **Test-first**: each subagent writes tests before writing the implementation. The workflow is: (1) write test file(s) with all test cases, (2) verify tests compile (using stub types/functions if needed) and fail as expected, (3) write the implementation, (4) verify all tests pass. This applies to every task that produces Go source — the only exceptions are Task 1 (directory structure) and Task 18 (smoke test).
- Each subagent must ensure `go build ./...` and `go vet ./...` pass before merging.
- After merging a wave, each subsequent wave's subagents pull the latest `main` into their worktree before starting.

### 5.2 Dependency Graph

```
Wave 1 (foundation — sequential, single agent):
  Task 1: module init + deps

Wave 2 (parallel — 3 subagents):
  ┌─ Agent A: Tasks 2+3 (model types + errors)     → branch: types-and-errors
  ├─ Agent B: Task 4 (option types)                 → branch: options
  └─ Agent C: Task 15+16 (macro system + tests)     → branch: macros
     (Task 15 depends only on model types from Task 2;
      Agent C waits for Agent A to merge before starting)

Wave 3 (parallel — 3 subagents):
  ┌─ Agent D: Tasks 5+6+7+8 (dialect system + both dialects + tests)  → branch: dialects
  ├─ Agent E: Tasks 9+11-partial (scanner + scanner tests)            → branch: scanner
  └─ Agent F: Tasks 12+13 (generator + tests)                         → branch: generator

Wave 4 (parallel — 2 subagents):
  ┌─ Agent G: Tasks 10+11 (parser + parser tests)   → branch: parser
  └─ Agent H: Task 14 (round-trip corpus tests)      → branch: roundtrip
     (Task 14 depends on parser+generator; Agent H waits for Agent G to merge)

Wave 5 (sequential — single agent):
  Task 17+18: package doc, examples, final build smoke test → branch: docs-and-examples
```

### 5.3 Task Definitions

---

#### Wave 1: Foundation (sequential)

**Task 1: Initialise module and directory structure** [Simple]
- Branch: `init-structure`
- Files: `go.mod` (already exists), create `docs/`, `internal/scan/`, `dialects/marlin/`, `dialects/reprap/`
- Verify `go 1.26.1` directive is present.
- Add `require github.com/lestrrat-go/option/v2` and `require github.com/stretchr/testify` via `go get`.
- No Go source to write yet.
- Merge to `main` before Wave 2 starts.

---

#### Wave 2: Core Types (parallel — 3 subagents)

**Agent A — Task 2+3: Core model types + error types** [Simple]
- Branch: `types-and-errors`
- Worktree: `.worktrees/types-and-errors`
- Files: `program.go`, `command.go`, `comment.go`, `errors.go`
- Task 2: Define `CommentForm`, `Comment` (with `CommentForm.String()`), `Parameter`, `Command`, `Line` (with `IsBlank()`), `Program`, `ProgramBuilder`.
  - `Line` uses value types for `Command` and `Comment` with `HasCommand`/`HasComment` bool flags (no pointer fields). Include `Raw string` field.
  - `ProgramBuilder.Append` copies each Line's `Command.Params` slice for full independence.
  - Trivial accessor methods: `Program.Lines()`, `Program.Len()`, `Program.Line(i)`.
  - Add Go doc comments on every exported type and field.
- Task 3: Define `ParseErrorDetail` exported interface, unexported `parseError` struct with `Error()`, `Unwrap()`, `Is()` methods, and accessor methods `Line()`, `Column()`, `Text()`.
  - Implement `ErrParse` sentinel variable: `var ErrParse = errors.New("gcode: parse error")`.
  - `(*parseError).Is` matches against `ErrParse`.
  - Add `makeParseError(line, col int, text string, err error) *parseError` private constructor.

**Agent B — Task 4: Option ident types and constructors** [Simple]
- Branch: `options`
- Worktree: `.worktrees/options`
- Files: `options.go` (shared), `parse_options.go` (parse-specific), `generate_options.go` (generate-specific)
- Define `Option` (shared supertype satisfying both `ParseOption` and `GenerateOption`), `ParseOption`, `GenerateOption` interface types (each embeds `option.Interface` and has a private marker method).
- Define carrier structs: `parseOption`, `generateOption`, `sharedOption` (see §4.1).
- Define private ident structs: `identDialect{}`, `identStrict{}`, `identEmitComments{}`, `identEmitLineNumbers{}`, `identComputeChecksum{}`, `identLineEnding{}`.
- `WithDialect` returns `Option` (shared). `WithStrict()` returns `ParseOption` (no bool arg). Generate-specific `With…` return `GenerateOption`.

**Agent C — Task 15+16: Macro system + tests** [Simple]
- Branch: `macros`
- Worktree: `.worktrees/macros`
- **Starts after Agent A merges** (depends on model types from Task 2).
- Files: `macro.go`, `macro_test.go`
- Task 15: Implement `Macro` interface, `SimpleMacro` with `NewSimpleMacro(name string, lines []Line) *SimpleMacro`, `Name()`, `Expand(args map[string]float64) ([]Line, error)` (SimpleMacro ignores args, returns a copy of the stored lines).
  - Implement `MacroRegistry` with `NewMacroRegistry()`, `Register`, `Lookup`, `Expand`.
  - `Expand` returns a deep copy of the lines so the caller may mutate them safely.
- Task 16: Tests — register a `SimpleMacro` and expand it; expand unknown macro returns error; expanded lines are a copy; `MacroRegistry` with a custom `Macro` implementation.

---

#### Wave 3: Dialect + Scanner + Generator (parallel — 3 subagents)

All agents in this wave pull latest `main` (which includes Wave 2 merges).

**Agent D — Tasks 5+6+7+8: Dialect system + Marlin + RepRap + tests** [Moderate]
- Branch: `dialects`
- Worktree: `.worktrees/dialects`
- Files: `dialect.go`, `dialect_registry.go`, `dialects/marlin/marlin.go`, `dialects/reprap/reprap.go`, `dialect_test.go`
- Task 5: Implement `commandKey`, `CommandDef`, `ParamDef` types. Implement `Dialect` with `NewDialect`, `Name`, `Register`, `LookupCommand`, `Commands`, `Extend`. `Extend` copies the parent's flattened command map into the new child. `LookupCommand` is a single map lookup (no parent traversal, no locks). `Register` mutates the command map; silently overwrites on duplicate key. Implement `DialectRegistry` with `NewDialectRegistry`, `Register`, `Lookup`.
- Task 6: Package `marlin` exports `Dialect() *gcode.Dialect` returning a fresh instance per call. Populate command table per §3.2.3.
- Task 7: Package `reprap` exports `Dialect() *gcode.Dialect`. Implement by calling `marlin.Dialect().Extend("reprap")` and registering RepRap-specific commands (G10, G11, M116, M557, M558).
- Task 8: Tests — `NewDialect`, `Register`, `LookupCommand`, `Commands()`, `Extend` independence, unknown command lookup, Marlin/RepRap correctness, fresh-instance-per-call.

**Agent E — Task 9: Internal line scanner + tests** [Moderate]
- Branch: `scanner`
- Worktree: `.worktrees/scanner`
- Files: `internal/scan/scan.go`, `internal/scan/scan_test.go`
- Implement `Scanner` struct wrapping `bufio.Scanner` with line number tracking.
- Implement `parseLine(rawLine string) (Line, error)` — core tokenisation:
  - Index-based scanning (no external cursor dependency).
  - Normalise command letters to upper-case.
  - Handle: optional `N<int>`, command letter+number (with optional `.subcode`), parameters, `;` comments, `(…)` comments (after params only), `*<checksum>`.
  - Store original line text in `Line.Raw`.
  - Return `parseError` (wrapping `ErrParse`) on malformed input.
- Keep free of dialect logic.
- Include `scan_test.go` with direct unit tests for `parseLine` covering all token forms.

**Agent F — Tasks 12+13: Generator + tests** [Moderate]
- Branch: `generator`
- Worktree: `.worktrees/generator`
- Files: `generator.go`, `generator_test.go`
- Task 12: Implement `Generator` struct with fields from §3.2.5. `NewGenerator` sets `emitComments: true` as default. `GenerateLine(w io.Writer, line Line) error` (value type). `Generate(w io.Writer, prog *Program) error` appends line ending after each `GenerateLine` call. Top-level `Generate`, `GenerateString`, `GenerateBytes` wrappers.
- Task 13: Tests — blank lines, comment-only lines with emitComments true/false, commands with/without params, line numbers, checksum computation, CRLF/LF, `GenerateString` round-trip.

---

#### Wave 4: Parser + Round-Trip (parallel — 2 subagents)

All agents pull latest `main` (which includes Wave 3 merges).

**Agent G — Tasks 10+11: Parser + tests** [Moderate]
- Branch: `parser`
- Worktree: `.worktrees/parser`
- Files: `parser.go`, `gcode.go` (top-level Parse functions), `parser_test.go`
- Task 10: Implement `Parser` struct with `dialect *Dialect` and `strict bool`. `NewParser(options ...ParseOption) *Parser` using ident-switch. `(*Parser).Parse(src io.Reader) (*Program, error)` — delegates to `internal/scan.parseLine`, validates against dialect in strict mode, accumulates into `ProgramBuilder`. `ParseString`, `ParseBytes` wrappers. Top-level `Parse`, `ParseString`, `ParseBytes` package-level functions.
- Task 11: Tests — blank lines, comment-only lines (`;` and `(…)`), G0/G1 with float params, M-codes, `N<num>` prefix, `*<checksum>` suffix, inline trailing comment, malformed input, strict/non-strict mode, multi-line program.

**Agent H — Task 14: Round-trip corpus tests** [Moderate]
- Branch: `roundtrip`
- Worktree: `.worktrees/roundtrip`
- **Starts after Agent G merges** (depends on parser).
- Files: `roundtrip_test.go`, `testdata/*.gcode`
- Add 3–5 sample `.gcode` files covering: Marlin print start sequence, file with `N` line numbers and checksums, file with both comment forms.
- Test: parse → generate → parse again → compare `Program.Lines()` field-by-field.

---

#### Wave 5: Integration (sequential)

**Task 17+18: Package-level doc, examples, final smoke test** [Simple]
- Branch: `docs-and-examples`
- Worktree: `.worktrees/docs-and-examples`
- Files: `gcode.go` (package doc), `example_test.go`
- Task 17: Write package doc comment in `gcode.go`. Write `Example_parse`, `Example_generate`, `Example_macro` runnable examples.
- Task 18: Run `go build ./...` and `go vet ./...`. Resolve any issues.

---

## 6. Security Considerations

- **Input untrusted G-code**: the parser does not execute any code. Adversarial input can produce `ParseError` but cannot cause unbounded memory allocation. The scanner reads line by line; individual lines exceeding `bufio.MaxScanTokenSize` (64 KiB default) are rejected with a `ParseError`.
- **Integer overflow**: command numbers and line numbers are stored as `int`. Input such as `G999999999999` is parsed using `strconv.Atoi` which will return an error; the parser propagates this as a `ParseError`.
- **No shell invocation**: the library never spawns processes.
- **No file I/O in core package**: file reading is the caller's responsibility. The library operates on `io.Reader`.

---

## 7. Performance Considerations

- **Allocation budget**: each `Line` value is stack-friendly (struct, not pointer). The `[]Parameter` slice on `Command` is the primary allocation per line. For programs with thousands of lines, callers should pre-allocate `ProgramBuilder` capacity if the line count is known.
- **strings.Builder** is used in the generator to avoid `fmt.Sprintf` allocations in the hot path.
- **bufio.Scanner** with default 64 KiB token size handles typical G-code files well. For very large files (SD card prints, hundreds of thousands of lines), the scanner processes one line at a time, so memory usage is proportional to the longest line, not the whole file.
- **Dialect lookups** are `O(1)` single map lookups on a flattened command map. No locking required.
- **No pooling** in v1. If profiling reveals significant GC pressure in macro expansion, `sync.Pool` for `[]Line` can be introduced.

---

## 8. Rollout Plan

This is a new library with no existing users. All phases can be developed and merged to the main branch directly.

Suggested release milestones:

- **v0.1.0**: Phases 1–4 + Marlin dialect (model, parser, generator, Marlin dialect, tests). Enough to parse real 3D printer G-code.
- **v0.2.0**: RepRap dialect, macro system, round-trip corpus tests.
- **v0.3.0**: Phase 7 (examples, docs, corpus tests). Announce for external feedback.
- **v1.0.0**: After user feedback stabilises the API surface. Semantic versioning applies; no breaking changes after v1.

There are no feature flags or gradual rollout concerns for a library. Monitoring is the caller's responsibility.

---

## 9. Open Questions and Risks

| # | Question / Risk | Mitigation |
|---|-----------------|------------|
| 1 | **Subcode representation**: some firmwares use `G92.1` (dot-subcode) while others use `G92 S1` (param-as-mode). Should subcodes be parsed as a separate field or folded into a synthetic parameter? | Current design uses a separate `Subcode int` field. Revisit if firmware corpus shows `G92 S1` patterns that conflict. |
| 2 | **T command form**: some G-code uses bare `T0`, others use `T 0` (with space). The scanner should handle both. | Treat `T` as a command letter like `G`/`M`; the number immediately follows (with optional whitespace). |
| 3 | **Case sensitivity**: G-code is technically case-insensitive (`g0` vs `G0`). | **Resolved**: the parser normalises command and parameter letters to upper-case in the scanner. Documented in §3.2.4 parsing algorithm. |
| 4 | **Float formatting in generator**: some firmware is sensitive to the number of decimal places. `strconv.FormatFloat(v, 'f', -1, 64)` emits the minimum digits. A `WithFloatPrecision(n int)` option may be needed. | Add as a `GenerateOption` in v0.2 if user feedback requests it. |
| 5 | **Checksum validation during parse**: should the parser verify incoming checksums? | Add `WithVerifyChecksum(bool)` as a `ParseOption` in v0.2. Default false to avoid breaking files with wrong checksums. |
| 6 | **Comment preservation order**: a line could theoretically have both a `(…)` comment and a `;` comment. | **Resolved**: the model stores one `Comment` per `Line`. If a line has both forms, the parser keeps the `;` comment (which is more common and typically carries the semantically meaningful text). The `(…)` comment is discarded. This is a documented round-trip fidelity loss for the rare dual-comment case. Revisit if real corpus shows need for `[]Comment`. |
| 7 | **RepRap dialect inheritance**: RepRap and Marlin share most commands. | **Resolved**: dialect command maps are flattened at `Extend` time. `reprap.Dialect()` returns a fresh instance with Marlin commands copied in. Changes to Marlin after extension do not propagate. |
