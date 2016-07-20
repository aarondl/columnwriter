package columns

import (
	"bytes"
	"io"
)

var (
	newlineChar = []byte{'\n'}
	spaceChar   = []byte{' '}
)

type bufList []*bytes.Buffer

// Writer wraps another io.Writer and can perform table output.
type Writer struct {
	writer io.Writer

	bufs   bufList
	column int
}

// New constructs a writer.
func New(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		bufs:   bufList{&bytes.Buffer{}},
	}
}

// NextCol adds a new column, any subsequent writes will be to that column.
func (w *Writer) NextCol() {
	w.column++

	w.bufs = append(w.bufs, &bytes.Buffer{})
}

// Write to the columns internal buffer
func (w *Writer) Write(b []byte) (int, error) {
	return w.bufs[w.column].Write(b)
}

// Flush the output to the wrapped io.Writer.
func (w *Writer) Flush() error {
	colLines := make([][][]byte, w.column+1)
	colWidths := make([]int, w.column+1)
	var maxLines int

	for col, b := range w.bufs {
		splits := bytes.Split(b.Bytes(), newlineChar)

		if len(splits[len(splits)-1]) == 0 {
			splits = splits[:len(splits)-1]
		}

		colLines[col] = splits
		if ln := len(splits); ln > maxLines {
			maxLines = ln
		}

		for _, line := range splits {
			if len(line) > colWidths[col] {
				colWidths[col] = len(line)
			}
		}
	}

	for i := 0; i < maxLines; i++ {
		for col, column := range colLines {
			var ln = colWidths[col]
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
