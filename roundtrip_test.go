package gcode_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/lestrrat-go/gcode"
	"github.com/stretchr/testify/require"
)

// streamCopy reads every line from src through gcode.Reader, writes
// each through gcode.Writer, and returns the formatted output.
func streamCopy(t *testing.T, src io.Reader) []byte {
	t.Helper()
	var buf bytes.Buffer
	r := gcode.NewReader(src)
	w := gcode.NewWriter(&buf)
	var line gcode.Line
	for {
		err := r.Read(&line)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.NoError(t, w.Write(line))
	}
	require.NoError(t, w.Flush())
	return buf.Bytes()
}

func TestRoundTripCorpus(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob("testdata/*.gcode")
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	for _, path := range matches {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			source, err := os.ReadFile(path)
			require.NoError(t, err)

			// First pass canonicalises; second pass must be stable.
			pass1 := streamCopy(t, bytes.NewReader(source))
			pass2 := streamCopy(t, bytes.NewReader(pass1))
			require.Equal(t, string(pass1), string(pass2),
				"round-trip not stable for %s", path)
		})
	}
}

// TestRoundTripOrcaSlicerSample exercises the streaming pipeline
// against a real-world OrcaSlicer Klipper-bound file when one is
// present at .tmp/sample.gcode. The test is skipped if the file is
// not provided locally.
func TestRoundTripOrcaSlicerSample(t *testing.T) {
	t.Parallel()
	const path = ".tmp/sample.gcode"
	source, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("sample file %s not available", path)
	}
	require.NoError(t, err)

	pass1 := streamCopy(t, bytes.NewReader(source))
	pass2 := streamCopy(t, bytes.NewReader(pass1))
	require.Equal(t, len(pass1), len(pass2))
	require.Equal(t, pass1, pass2, "round-trip not stable")
}
