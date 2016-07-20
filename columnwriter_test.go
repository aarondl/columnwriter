package columns

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

var flagWriteGoldenFiles = flag.Bool("test.golden", false, "Controls golden file generation")

func TestBaseCase(t *testing.T) {
	buf := &bytes.Buffer{}

	writer := New(buf)

	fmt.Fprintln(writer, "a")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 1, currentLines: 1, maxLines: 1, currentCol: 0}); err != nil {
		t.Error(err)
	}
	fmt.Fprintln(writer, "there")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 5, currentLines: 2, maxLines: 2, currentCol: 0}); err != nil {
		t.Error(err)
	}

	writer.NextCol()
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 5, currentWidth: 0, currentLines: 0, maxLines: 2, currentCol: 1}); err != nil {
		t.Error(err)
	}

	fmt.Fprintln(writer, "hello")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 5, currentWidth: 5, currentLines: 1, maxLines: 2, currentCol: 1}); err != nil {
		t.Error(err)
	}
	fmt.Fprintln(writer, "hec")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 5, currentWidth: 5, currentLines: 2, maxLines: 2, currentCol: 1}); err != nil {
		t.Error(err)
	}

	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", buf.Bytes())
	compareGoldenFile("basecase.txt", buf.Bytes(), t)
}

func TestMissingVal(t *testing.T) {
	buf := &bytes.Buffer{}

	writer := New(buf)

	fmt.Fprintln(writer, "a")
	writer.NextCol()
	fmt.Fprintln(writer, "there")
	fmt.Fprintln(writer, "b")

	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", buf.Bytes())
	compareGoldenFile("missingval.txt", buf.Bytes(), t)
}

func TestMultipleWriteNoNewline(t *testing.T) {
	buf := &bytes.Buffer{}

	writer := New(buf)

	fmt.Fprint(writer, "a")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 1, currentLines: 1, maxLines: 1, currentCol: 0}); err != nil {
		t.Error(err)
	}

	fmt.Fprint(writer, " b")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 3, currentLines: 1, maxLines: 1, currentCol: 0}); err != nil {
		t.Error(err)
	}

	fmt.Fprint(writer, " c")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 5, currentLines: 1, maxLines: 1, currentCol: 0}); err != nil {
		t.Error(err)
	}

	fmt.Fprintln(writer, " d")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 0, currentLines: 2, maxLines: 2, currentCol: 0}); err != nil {
		t.Error(err)
	}

	fmt.Fprintln(writer, "there")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 0, currentWidth: 5, currentLines: 2, maxLines: 2, currentCol: 0}); err != nil {
		t.Error(err)
	}

	writer.NextCol()
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 5, currentWidth: 0, currentLines: 0, maxLines: 2, currentCol: 1}); err != nil {
		t.Error(err)
	}

	fmt.Fprintln(writer, "hello")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 5, currentWidth: 5, currentLines: 1, maxLines: 2, currentCol: 1}); err != nil {
		t.Error(err)
	}
	fmt.Fprintln(writer, "hec")
	if err := stateCheck(t, writer, stateChecker{
		lastColWidth: 5, currentWidth: 5, currentLines: 2, maxLines: 2, currentCol: 1}); err != nil {
		t.Error(err)
	}

	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", buf.Bytes())
	compareGoldenFile("nonewline.txt", buf.Bytes(), t)
}

type stateChecker struct {
	lastColWidth int
	currentWidth int
	currentLines int
	maxLines     int
	currentCol   int
}

func stateCheck(t *testing.T, writer *Writer, state stateChecker) error {
	if writer.currentCol != state.currentCol {
		return fmt.Errorf("current col wrong: %v want: %v", writer.currentCol, state.currentCol)
	}
	if width := writer.colWidths[len(writer.colWidths)-1]; width != state.currentWidth {
		return fmt.Errorf("currentWidth wrong: %v want: %v", width, state.currentWidth)
	}
	if writer.currentLines != state.currentLines {
		return fmt.Errorf("currentLines wrong: %v want: %v", writer.currentLines, state.currentLines)
	}
	if writer.maxLines != state.maxLines {
		return fmt.Errorf("maxLines wrong: %v want: %v", writer.maxLines, state.maxLines)
	}

	return nil
}

func compareGoldenFile(filename string, input []byte, t *testing.T) {
	filename = filepath.Join("_fixtures", filename)
	if *flagWriteGoldenFiles {
		if err := ioutil.WriteFile(filename, input, 0664); err != nil {
			panic(err)
		}
		return
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal("failed to read file:", filename, err)
	}
	if bytes.Compare(b, input) != 0 {
		t.Errorf("test failed to compare: %s\nwant: %s\ngot: %s\n",
			filename, spew.Sdump(b), spew.Sdump(input))
	}
}
