// Package gcode provides streaming read and write of G-code, the control
// language used by CNC machines and 3D printers.
//
// G-code files commonly run to hundreds of megabytes. The package is
// designed around streaming: a [Reader] decodes one [Line] at a time
// from an [io.Reader], and a [Writer] encodes lines to an [io.Writer].
// Neither holds the whole program in memory.
//
// The data model is intentionally uniform across classic G/M/T codes
// and Klipper-style extended commands. A [Command] carries a canonical
// string Name ("G28", "G92.1", "EXCLUDE_OBJECT_DEFINE") and a slice of
// keyed [Argument] values. See [Argument] for the value representation.
//
// Dialects ([Dialect]) describe which commands a particular firmware
// understands. Built-in dialects live in subpackages — see
// [github.com/lestrrat-go/gcode/dialects/marlin],
// [github.com/lestrrat-go/gcode/dialects/reprap], and
// [github.com/lestrrat-go/gcode/dialects/klipper].
//
// A typical streaming pipeline looks like:
//
//	r := gcode.NewReader(in)
//	w := gcode.NewWriter(out)
//	var line gcode.Line
//	for {
//	    if err := r.Read(&line); err == io.EOF {
//	        break
//	    } else if err != nil {
//	        return err
//	    }
//	    // ... inspect or mutate line ...
//	    if err := w.Write(line); err != nil {
//	        return err
//	    }
//	}
//	if err := w.Flush(); err != nil {
//	    return err
//	}
//
// Lines returned by [Reader.Read] are backed by the Reader's internal
// buffers and remain valid only until the next call to Read; use
// [Line.Clone] to retain a copy.
package gcode
