# Design Document: github.com/lestrrat-go/gcode

## 1. Executive Summary

`github.com/lestrrat-go/gcode` is a Go library for streaming G-code: reading text into a structured representation and writing that representation back out. G-code is the numerical control language used by CNC machines, 3D printers, and laser cutters.

The library targets three firmware dialects out of the box — Marlin, RepRap, and Klipper — with an extensible dialect system that lets downstream packages register additional ones. It also exposes a macro mechanism for user-defined named command sequences.

Key design decisions:

- **Streaming first.** Real-world G-code files commonly run to hundreds of megabytes. The library is built around a [`Reader`] / [`Writer`] pair (mirroring `csv.Reader` / `csv.Writer`) that processes one line at a time without holding the whole program in memory. There is no `Program` aggregate.
- **Unified command model.** Classic G/M/T codes (`G28`, `M104`, `G92.1`) and Klipper-style extended commands (`EXCLUDE_OBJECT_DEFINE`, `SET_FAN_SPEED`) share one `Command` shape — a canonical string `Name` plus a slice of keyed `Argument` values. No discriminator unions, no parallel argument types.
- **Round-trip fidelity.** Whitespace is canonicalised on output, but every meaningful field round-trips: extended-argument key case, list/quoted argument values, comments, line numbers, and checksums. `Argument.Raw` preserves the value's source text verbatim.
- **Bounded memory.** The Reader copies each source line into a single internal buffer and hands out `string` views via `unsafe.String`. Per-line allocations are constant; total allocation is O(file lines) for the small `Argument` slabs alone.
- **Actionable errors.** Parse errors carry source line number, byte column, and a short excerpt of the offending text.

---

## 2. Requirements Analysis

### 2.1 Functional Requirements

1. Stream-decode G-code from any `io.Reader` into a sequence of `Line` values via a `Reader`.
2. Stream-encode `Line` values to any `io.Writer` via a `Writer`.
3. Each `Line` may carry: an optional line-number prefix (`N<int>`), a command, a trailing comment, and an optional checksum (`*<byte>`). Each component is gated by a `HasX` flag.
4. A `Command` is identified by a canonical string `Name`:
   - Classic: letter + integer + optional `.subcode` (`"G28"`, `"M104"`, `"G92.1"`).
   - Extended: bare identifier (`"EXCLUDE_OBJECT_DEFINE"`).
5. A `Command` carries an ordered slice of `Argument` values. Each `Argument` has a `Key` (single letter for classic, identifier for extended), a verbatim `Raw` value, and parsed `Number` / `IsNumeric` fields when the value parses as a finite float.
6. Comments may be `;`-form or `(...)`-form, comment-only or trailing, and round-trip with their form preserved.
7. Dialects (`Dialect`) describe which commands a firmware understands and what arguments each accepts. The Reader can run in strict mode against a dialect to reject unknown commands.
8. Dialects compose by inheritance: `Extend(name)` produces a child dialect with a flattened copy of the parent's command table.
9. Macros (`MacroRegistry`) expand a named entry into a `[]Line` slice for the caller to feed through the Writer.
10. Module path: `github.com/lestrrat-go/gcode`. Go 1.26+ (uses `iter.Seq2`, `unsafe.String`).

### 2.2 Non-Functional Requirements

- **Throughput**: parse a 1.7 MB / 62 k-line OrcaSlicer file end-to-end in under 100 ms with bounded memory (~5 MB resident, <10 MB total alloc) on commodity hardware.
- **Concurrency**: `Reader`, `Writer`, `Macro`, and `Dialect` mutating operations are not concurrent-safe and require external synchronisation. `MacroRegistry.Lookup` / `Expand` and `DialectRegistry.Lookup` are concurrent-read-safe.
- **Per-line invariant**: lines returned by `Reader.Read` share the Reader's internal buffer. They are valid only until the next `Read` call; callers must use `Line.Clone` to retain.
- **Errors**: `errors.Is(err, ErrParse)` matches all parse errors; `errors.As` extracts a `ParseErrorDetail` with line/column/text.
- **Dependencies**: only `github.com/lestrrat-go/option/v2` and `github.com/stretchr/testify` (test-only).

### 2.3 Constraints and Assumptions

- Parameter numeric values are `float64`. No arbitrary precision.
- Comments: `;` to end-of-line, or `(...)` (single-line, balanced).
- Subcodes are part of the canonical `Name` (e.g., `"G92.1"`); there is no separate field.
- Checksum byte is XOR of all bytes from the `N` (or first non-whitespace if no `N`) up to but not including the `*`, matching Marlin/RepRap convention.
- Source case folding: classic command letters and classic argument letters are upper-cased to canonical form. Extended command names and extended argument keys preserve source case to retain emitter convention on round-trip.
- This version does not evaluate or execute G-code; it only reads and writes text.
- Macro expansion does not recurse.
- The Reader is tolerant of free-form trailing text after a classic command (e.g., `M117 Hello World`): the structured args list is empty and `Line.Raw` preserves the original source.

### 2.4 Out of Scope

- Simulation or toolpath visualisation.
- Binary G-code formats (e.g., the Klipper binary protocol).
- Recursive macro expansion.
- Network streaming protocols.
- Any UI or CLI tool — pure library only.
- Multiple commands per line.
- Expression syntax / meta-commands (Marlin 2.x `if`/`while`/`{var.x+10}`, RepRapFirmware 3+ `if sensors.filament.runout`).
- A persistent in-memory `Program` aggregate. Callers who genuinely want an in-memory representation can collect cloned `Line` values into a slice themselves.

---

## 3. Technical Design

### 3.1 Architecture Overview

```
                    ┌──────────────────────┐
   io.Reader ────►  │  gcode.Reader        │  ──► *Line (reused buffer)
                    │  (line scanner +     │
                    │   in-place parser)   │
                    └──────────────────────┘
                              │
                    ┌──────────────────────┐
   *Line          ► │  gcode.Writer        │  ──► io.Writer
                    │  (formatter +        │
                    │   buffered output)   │
                    └──────────────────────┘

   Dialect attaches to Reader (via WithDialect) for strict-mode validation.
   MacroRegistry produces []Line for the caller to feed through Writer.
```

A typical pipeline:

```go
r := gcode.NewReader(in)
w := gcode.NewWriter(out)
var line gcode.Line
for {
    err := r.Read(&line)
    if err == io.EOF { break }
    if err != nil { return err }
    // mutate or inspect line ...
    if err := w.Write(line); err != nil { return err }
}
return w.Flush()
```

Memory is bounded by the longest single source line (default cap: 16 MiB).

### 3.2 Package Layout

```
github.com/lestrrat-go/gcode/
├── gcode.go              package doc and overview
├── command.go            Command, Argument, Arg helper
├── line.go               Line, Clone, IsBlank, internal reset
├── comment.go            Comment, CommentForm
├── reader.go             Reader, NewReader, Read, All, parseInto
├── writer.go             Writer, NewWriter, Write, Flush
├── dialect.go            Dialect, CommandDef, ParamDef
├── dialect_registry.go   DialectRegistry
├── macro.go              Macro, SimpleMacro, MacroRegistry
├── errors.go             ErrParse, ParseErrorDetail, parseError
├── options.go            ReadOption / WriteOption / Option type system
├── read_options.go       WithStrict, WithMaxLineSize
├── write_options.go      WithEmitComments, WithEmitLineNumbers,
│                         WithComputeChecksum, WithLineEnding
├── dialects/
│   ├── marlin/marlin.go
│   ├── reprap/reprap.go  (extends marlin)
│   └── klipper/klipper.go (extends marlin + extended commands)
├── examples/
│   └── stream_test.go    runnable Example* functions
└── testdata/             corpus for round-trip tests
```

### 3.3 Core Types

```go
type Command struct {
    Name string      // "G28", "G92.1", "EXCLUDE_OBJECT_DEFINE"
    Args []Argument  // in source order
}

func (c Command) Arg(key string) (Argument, bool)

type Argument struct {
    Key       string  // "X" (classic) or "FAN" (extended)
    Raw       string  // verbatim source after the key (or after '=')
    Number    float64 // valid only if IsNumeric
    IsNumeric bool
}

func (a Argument) IsFlag() bool  // true when Raw == ""

type Line struct {
    LineNumber  int
    Command     Command
    HasCommand  bool
    Comment     Comment
    HasComment  bool
    Checksum    byte
    HasChecksum bool
    Raw         string  // original source line, populated by Reader
}

func (l Line) IsBlank() bool
func (l Line) Clone() Line       // detach from Reader buffer
```

### 3.4 Reader

`Reader` wraps a `bufio.Scanner` configured with a generous max-line buffer (default 16 MiB) so it can absorb long Klipper extended-command lines.

For each `Read(line *Line)` call:

1. `line.reset()` returns the line to its zero state while preserving any backing `Args` slice capacity.
2. The Scanner advances; if exhausted, return `io.EOF`.
3. The source line bytes are copied into the Reader's internal `buf` (single `append` reuse; no per-call allocation when `buf` is large enough).
4. `line.Raw` is set to a `string` view of `buf` via `unsafe.String`.
5. `line.Command.Args = r.args[:0]` reuses the cached `Argument` slab.
6. `parseInto(line)` walks `buf`:
   - Optional `N<digits>` line number.
   - Comment-only `;` or `(…)` line.
   - **Command:** if the byte after the first letter is a digit, parse as **classic** (letter + digits + optional `.digits`); otherwise scan a **extended** identifier (`[A-Za-z_][A-Za-z0-9_]*`).
   - **Arguments:** the grammar follows the command kind. Classic accepts `<letter><number>` or `<letter>` (flag). Extended accepts `<identifier>=<value>` with balanced bracket / quoted value scanning, or a bare `<identifier>` flag.
   - Trailing `(…)` and `;` comments, optional `*<digits>` checksum.
7. `r.args` is updated to retain the (possibly grown) slab for the next call.
8. Strict-mode dialect check: if `WithStrict()` and a dialect are set, the canonical `Name` must exist in the dialect.

`Reader.All() iter.Seq2[Line, error]` wraps `Read` for `range`-loop convenience.

**Buffer ownership.** The strings returned on `*line` (`Name`, `Args[i].Key`, `Args[i].Raw`, `Comment.Text`, `Raw`) all alias the Reader's `buf`. Calling `Read` again overwrites `buf`. `Line.Clone()` deep-copies into fresh allocations.

**Tolerance.** Free-form trailing text after a classic command (e.g., `M117 Hello World`) is detected when an argument letter is followed by another letter; the structured args list is cleared and parsing stops. `Line.Raw` still has the original source for callers that need it.

### 3.5 Writer

`Writer` wraps a `bufio.Writer`. Each `Write(line Line)` formats `line` into a single string and emits it followed by the configured line ending.

Formatting rules:

- Commands are emitted by `Name` followed by either `' '<Key><Raw>` (classic) or `' '<Key>=<Raw>` (extended). Bare flags emit `' '<Key>` only.
- Optional `N<n> ` prefix when `WithEmitLineNumbers(true)`.
- Optional trailing `; comment` or ` (comment)` when `WithEmitComments(true)`.
- Optional `*<checksum>` suffix when `WithComputeChecksum(true)`. Checksum is the XOR of the formatted body bytes.
- Line endings: `\n` (default) or `\r\n` (`WithLineEnding(LineEndingCRLF)`).

`Writer.Flush()` flushes the buffered writer.

### 3.6 Dialect

```go
type Dialect struct {
    name     string
    commands map[string]*CommandDef
}

type CommandDef struct {
    Name        string
    Description string
    Params      []ParamDef  // nil = unconstrained, empty = no params
}

type ParamDef struct {
    Key         string  // canonical key (single-letter classic or identifier)
    Required    bool
    Description string
}
```

`Extend(name)` returns a child dialect with a flattened copy of the parent's command table. Lookups are a single string-keyed map access.

`DialectRegistry` is a thread-safe `map[string]*Dialect` for looking up dialects by name.

Built-in dialects:
- `dialects/marlin` — common G/M codes, T0-T5.
- `dialects/reprap` — extends marlin with G10/G11/M116/M557/M558.
- `dialects/klipper` — extends marlin with **always-available** Klipper core extended commands (`SET_PRESSURE_ADVANCE`, `SET_VELOCITY_LIMIT`, `SAVE_GCODE_STATE`, `RESTORE_GCODE_STATE`, `SET_PRINT_STATS_INFO`).

Each `Dialect()` constructor returns a **shared singleton** initialised at package load. Mutating it (`dialect.Register(...)`) affects every caller. Callers wanting a private extension must call `Extend(name)` to get a flattened mutable child first.

Klipper has many config-dependent and plugin-supplied commands; the `klipper` package exposes them as opt-in helpers that **clone** the input dialect and return a new one (so they never mutate the caller's input):

- `klipper.WithBedMesh(d)` — `BED_MESH_CALIBRATE`, `BED_MESH_PROFILE`, `BED_MESH_CLEAR` (requires `[bed_mesh]`).
- `klipper.WithExcludeObject(d)` — `EXCLUDE_OBJECT_*` (requires `[exclude_object]`).
- `klipper.WithFanGeneric(d)` — `SET_FAN_SPEED` (requires `[fan_generic ...]`).
- `klipper.WithTimelapse(d)` — `TIMELAPSE_TAKE_FRAME` (third-party moonraker-timelapse plugin).

Compose by chaining: `d := klipper.WithExcludeObject(klipper.WithBedMesh(klipper.Dialect()))`.

### 3.7 Macro

`Macro` is a single-method interface (`Expand(args map[string]float64) ([]Line, error)`). `SimpleMacro` provides a fixed-line implementation. `MacroRegistry` is a thread-safe collection.

The library does not provide template syntax or expression evaluation; callers implementing `Macro` interpret `args` themselves.

### 3.8 Errors

```go
var ErrParse = errors.New("gcode: parse error")

type ParseErrorDetail interface {
    error
    Line() int
    Column() int
    Text() string
}
```

All parse failures wrap with `ErrParse` (matchable via `errors.Is`) and implement `ParseErrorDetail` (extractable via `errors.As`). `Reader.Read` returns the error and stops; the Reader is not designed to recover and resume after a parse error.

---

## 4. Implementation Notes

### 4.1 Coding Standards

- Idiomatic Go; `gofmt` clean; no named return values; testify/require (no assert).
- External `xxx_test` packages where practical.
- Public symbols documented; godoc examples in `examples/`.

### 4.2 Performance

- The Reader is the hot path. Single-line parsing operates entirely on the in-place `buf` byte slice; no intermediate strings are allocated except via `unsafe.String` aliases.
- The `Argument` slab is reused across Reads — typically 0 allocations after warmup.
- The Writer uses an internal `[]byte` builder per line and a `bufio.Writer` for output coalescing.

### 4.3 Tests

- Unit tests for each module live next to the source.
- Round-trip tests stream a corpus (`testdata/*.gcode`) through Reader→Writer and verify a second Reader→Writer pass produces identical output (canonicalisation stability).
- An optional `.tmp/sample.gcode` test exercises a real-world OrcaSlicer Klipper file when present (skipped otherwise).
- Runnable `Example*` functions in `examples/` cover the documented public surface.

---

## 5. Out-of-Tree Roadmap

Not implemented; documented here for callers who may want to fork:

- Per-`Reader` allocation arena with `sync.Pool` for the `buf` slice — useful when many short-lived Readers are created.
- Random-access `LineIndex` (line number → byte offset table) for editor-style consumers that need to jump in the middle of a large file. Today, callers seeking line N must scan from the start.
- Pluggable comment-form policies (e.g., always semicolon-form on output regardless of source).
- Recursive macro expansion (currently a documented non-goal).
