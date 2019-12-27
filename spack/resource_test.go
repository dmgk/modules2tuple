package spack

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/dmgk/modules2tuple/tuple"
)

type testCase struct {
	AppVersion string
	Tuple      *tuple.Tuple
	Expected   string
}

// TestResource tests that an input tuple generates the expected
// resource string.
func TestResource(t *testing.T) {
	testCases := []testCase{
		// This should use "commit"
		{AppVersion: "1.2.3",
			Tuple: tParse("github.com/rivo/uniseg v0.0.0-20190313204849-f699dde9c340"),
			Expected: `
        {
          "name": "github.com/rivo/uniseg",
          "git": "https://github.com/rivo/uniseg",
          "commit": "f699dde9c340",
          "placement": "vendor/github.com/rivo/uniseg",
          "when": "@1.2.3",
          "destination": "."
        }`,
		},
		// This should use "tag" and resolve to github
		{AppVersion: "1.2.3",
			Tuple: tParse("golang.org/x/text v0.3.2"),
			Expected: `
        {
          "name": "golang.org/x/text",
          "git": "https://github.com/golang/text",
          "tag": "v0.3.2",
          "placement": "vendor/golang.org/x/text",
          "when": "@1.2.3",
          "destination": "."
        }`,
		},
		// empty AppVersion => no when
		{AppVersion: "",
			Tuple: tParse("golang.org/x/text v0.3.2"),
			Expected: `
        {
          "name": "golang.org/x/text",
          "git": "https://github.com/golang/text",
          "tag": "v0.3.2",
          "placement": "vendor/golang.org/x/text",
          "destination": "."
        }`,
		},
		// A nonstandard, GitLab repo
		{AppVersion: "",
			Tuple: tParse("howett.net/plist v0.0.0-20181124034731-591f970eefbb"),
			Expected: `
        {
          "name": "howett.net/plist",
          "git": "https://gitlab.howett.net/go/plist",
          "commit": "591f970eefbb",
          "placement": "vendor/howett.net/plist",
          "destination": "."
        }`,
		},
	}

	for _, c := range testCases {
		expected := []byte(c.Expected)
		r, err := resourceFromTuple(c.AppVersion, c.Tuple)
		if err != nil {
			t.Error(err)
		}
		got, err := r.ToJson()
		if err != nil {
			t.Error(err)
		}
		equal, err := jsonCompare(expected, got)
		if err != nil {
			t.Error(err)
		}
		if !equal {
			t.Errorf("Got: %s, expected %s", r, c.Expected)
		}
	}
}

// tParse is a helper function that parses a tuple string and returns
// the resulting *tuple.Tuple.  It panics if there's a problem parsing
// the string, since that's not what we're testing here.
func tParse(t string) *tuple.Tuple {
	tuple, err := tuple.Parse(t, "vendor")
	if err != nil {
		panic(err)
	}
	return tuple
}

// After https://gist.github.com/turtlemonvh/e4f7404e28387fadb8ad275a99596f67
func jsonCompare(j1, j2 []byte) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	err := json.Unmarshal([]byte(j1), &o1)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(j2), &o2)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1, o2), nil
}
