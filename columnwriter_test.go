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
	fmt.Fprintln(writer, "there")
	if stored := writer.bufs[0].String(); stored != "a\nthere\n" {
		t.Errorf("wrong stored:\ngot:\n%s\n", spew.Sdump([]byte(stored)))
	}

	if writer.column != 0 {
		t.Error("wrong column:", writer.column)
	}
	writer.NextCol()
	if writer.column != 1 {
		t.Error("wrong column:", writer.column)
	}

	fmt.Fprintln(writer, "hello")
	fmt.Fprintln(writer, "hec")

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

func TestNoNewlines(t *testing.T) {
	buf := &bytes.Buffer{}

	writer := New(buf)

	fmt.Fprint(writer, "a")
	fmt.Fprint(writer, "b")
	fmt.Fprint(writer, "c")
	fmt.Fprintln(writer, "d")
	writer.NextCol()
	fmt.Fprint(writer, "there")
	fmt.Fprintln(writer, "s")
	fmt.Fprintln(writer, "b")

	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", buf.Bytes())
	compareGoldenFile("nonewlines.txt", buf.Bytes(), t)
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
