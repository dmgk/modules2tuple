// +build e2e

package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const testdataPath = "../testdata"

type example struct {
	modules, expected string
}

func TestParserE2E(t *testing.T) {
	examples, err := loadExamples()
	if err != nil {
		t.Fatal(err)
	}

	for n, x := range examples {
		expected, err := ioutil.ReadFile(x.expected)
		if err != nil {
			t.Fatal(err)
		}

		actual, err := Load(x.modules)
		if err != nil {
			t.Fatal(err)
		}

		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(strings.TrimSpace(string(expected)), strings.TrimSpace(actual.String()), false)
		if dmp.DiffLevenshtein(diffs) > 0 {
			t.Errorf("%s: expected output doesn't match actual:\n%s", n, dmp.DiffPrettyText(diffs))
		}
	}
}

func loadExamples() (map[string]*example, error) {
	res := map[string]*example{}

	dir, err := ioutil.ReadDir(testdataPath)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		name := f.Name()
		parts := strings.SplitN(strings.TrimSuffix(name, filepath.Ext(name)), "_", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("unexpected testdata file name: %q", name)
		}
		key, kind := parts[0], parts[1]

		x, ok := res[key]
		if !ok {
			x = &example{}
			res[key] = x
		}
		switch kind {
		case "modules":
			x.modules = filepath.Join(testdataPath, name)
		case "expected":
			x.expected = filepath.Join(testdataPath, name)
		default:
			return nil, fmt.Errorf("unexpected testdata file name: %q", name)
		}
	}

	return res, nil
}
