/*Package golden provides tools for comparing large mutli-line strings.

Golden files are files in the ./testdata/ subdirectory of the package under test.
*/
package golden

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/pmezard/go-difflib/difflib"
)

var flagUpdate = flag.Bool("test.update-golden", false, "update golden file")

type helperT interface {
	Helper()
}

// Get returns the contents of the file in ./testdata
func Get(t assert.TestingT, filename string) []byte {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	expected, err := ioutil.ReadFile(Path(filename))
	assert.NilError(t, err)
	return expected
}

// Path returns the full path to a file in ./testdata
func Path(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join("testdata", filename)
}

func update(filename string, actual []byte, clean bool) error {
	if *flagUpdate {
		if clean {
			actual = bytes.Replace(actual, []byte("\r\n"), []byte("\n"), -1)
		}
		return ioutil.WriteFile(Path(filename), actual, 0644)
	}
	return nil
}

// Assert compares the actual content to the expected content in the golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
// Returns whether the assertion was successful (true) or not (false).
// This is equivalent to assert.Check(t, String(actual, filename))
//
// Deprecated: In a future version this function will change to use assert.Assert
// instead of assert.Check to be consistent with other assert functions.
// Use assert.Check(t, String(actual, filename) if you want to preserve the
// current behaviour.
func Assert(t assert.TestingT, actual string, filename string, msgAndArgs ...interface{}) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	actual = strings.Replace(actual, "\r\n", "\n", -1)
	return assert.Check(t, String(actual, filename), msgAndArgs...)
}

// String compares actual to the contents of filename and returns success
// if the strings are equal.
func String(actual string, filename string) cmp.Comparison {
	return func() cmp.Result {
		result, expected := compare([]byte(actual), filename, true)
		if result != nil {
			return result
		}
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(expected)),
			B:        difflib.SplitLines(actual),
			FromFile: "expected",
			ToFile:   "actual",
			Context:  3,
		})
		if err != nil {
			return cmp.ResultFromError(err)
		}
		return cmp.ResultFailure("\n" + diff)
	}
}

// AssertBytes compares the actual result to the expected result in the golden
// file. If the `-test.update-golden` flag is set then the actual content is
// written to the golden file.
// Returns whether the assertion was successful (true) or not (false)
// This is equivalent to assert.Check(t, Bytes(actual, filename))
//
// Deprecated: In a future version this function will change to use assert.Assert
// instead of assert.Check to be consistent with other assert functions.
// Use assert.Check(t, Bytes(actual, filename) if you want to preserve the
// current behaviour.
func AssertBytes(
	t assert.TestingT,
	actual []byte,
	filename string,
	msgAndArgs ...interface{},
) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	return assert.Check(t, Bytes(actual, filename), msgAndArgs...)
}

// Bytes compares actual to the contents of filename and returns success
// if the bytes are equal.
func Bytes(actual []byte, filename string) cmp.Comparison {
	return func() cmp.Result {
		result, expected := compare(actual, filename, false)
		if result != nil {
			return result
		}
		msg := fmt.Sprintf("%v (actual) != %v (expected)", actual, expected)
		return cmp.ResultFailure(msg)
	}
}

func compare(actual []byte, filename string, clean bool) (cmp.Result, []byte) {
	if err := update(filename, actual, clean); err != nil {
		return cmp.ResultFromError(err), nil
	}
	expected, err := ioutil.ReadFile(Path(filename))
	if err != nil {
		return cmp.ResultFromError(err), nil
	}
	if bytes.Equal(expected, actual) {
		return cmp.ResultSuccess, nil
	}
	return nil, expected
}
