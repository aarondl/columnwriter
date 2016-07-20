package columns

import (
	"bytes"
	"io"
)

var (
	newlineChar = []byte{'\n'}
	spaceChar   = []byte{' '}
)

type bufList [][][]byte

// Writer wraps another io.Writer and can perform table output.
type Writer struct {
	writer io.Writer

	currentLines int
	maxLines     int

	currentCol int
	colWidths  []int
	colLines   bufList
}

// New constructs a writer.
func New(w io.Writer) *Writer {
	return &Writer{
		writer:    w,
		colLines:  bufList{nil},
		colWidths: []int{0},
	}
}

// NextCol adds a new column, any subsequent writes will be to that column.
func (w *Writer) NextCol() {
	w.currentCol++

	w.currentLines = 0

	w.colWidths = append(w.colWidths, 0)
	w.colLines = append(w.colLines, nil)
}

// Write divides the input into lines based on the '\n' character. It stores
// the lines in an internal buffer until the .Render() call. Write always
// writes the lines to the current column, once NextCol is called there is no
// way to go back to the old column to add new lines.
func (w *Writer) Write(b []byte) (int, error) {
	// Copy data to not destroy println buffers
	ln := len(b)
	hasNewlines := false
	if bytes.HasSuffix(b, newlineChar) {
		ln--
		hasNewlines = true
	}
	lineData := make([]byte, ln)
	copy(lineData, b)
	lines := bytes.Split(lineData, newlineChar)

	hasNewlines = hasNewlines || len(lines) > 1

	w.currentLines += len(lines)
	if w.currentLines > w.maxLines {
		w.maxLines = w.currentLines
	}

	for _, line := range lines {
		length := len(line)
		if length > w.colWidths[w.currentCol] {
			w.colWidths[w.currentCol] = length
		}
	}

	w.colLines[w.currentCol] = append(w.colLines[w.currentCol], lines...)

	return len(b), nil
}

// Flush the output to the wrapped io.Writer.
func (w *Writer) Flush() error {
	for i := 0; i < w.maxLines; i++ {
		for j, column := range w.colLines {
			var ln = w.colWidths[j]
			var padding []byte

			if i >= len(column) {
				padding = bytes.Repeat(spaceChar, ln+1)
				if _, err := w.writer.Write(padding); err != nil {
					return err
				}

				continue
			}

			line := column[i]
			if _, err := w.writer.Write(line); err != nil {
				return err
			}
			if delta := ln - len(line) + 1; delta > 0 {
				padding = bytes.Repeat(spaceChar, delta)
			}

			if _, err := w.writer.Write(padding); err != nil {
				return err
			}
		}

		if _, err := w.writer.Write(newlineChar); err != nil {
			return err
		}
	}

	return nil
}
