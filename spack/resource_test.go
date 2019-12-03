package spack

import (
	"testing"

	"github.com/dmgk/modules2tuple/tuple"
)

// ExpectedError should only occur if regexp match fails or template
// execution fails.  I haven't figured out how to test those cases.
type testCase struct {
	AppVersion    string
	Tuple         *tuple.Tuple
	Expected      string
	ExpectedError string
}

// TestResource tests that an input tuple generates the expected
// resource string.
func TestResource(t *testing.T) {
	testCases := []testCase{
		// This should use "commit"
		{AppVersion: "1.2.3",
			Tuple: tParse("github.com/pkg/errors v1.2.3-20150716171945-2caba252f4dc"),
			Expected: `
    resource(name="github.com/pkg/errors",
             git="https://github.com/pkg/errors",
             commit="2caba252f4dc",
             destination=".",
             when="@1.2.3",
             placement="vendor/github.com/pkg/errors")`,
			ExpectedError: "",
		},
		// This should use "tag"
		{AppVersion: "1.2.3",
			Tuple: tParse("github.com/rogpeppe/go-internal v1.3.0"),
			Expected: `
    resource(name="github.com/rogpeppe/go-internal",
             git="https://github.com/rogpeppe/go-internal",
             tag="v1.3.0",
             destination=".",
             when="@1.2.3",
             placement="vendor/github.com/rogpeppe/go-internal")`,
			ExpectedError: "",
		},
		// empty AppVersion => no when
		{AppVersion: "",
			Tuple: tParse("github.com/rogpeppe/go-internal v1.3.0"),
			Expected: `
    resource(name="github.com/rogpeppe/go-internal",
             git="https://github.com/rogpeppe/go-internal",
             tag="v1.3.0",
             destination=".",
             placement="vendor/github.com/rogpeppe/go-internal")`,
			ExpectedError: "",
		},
		// A nonstandard, GitLab repo
		{AppVersion: "",
			Tuple: tParse("howett.net/plist v0.0.0-20181124034731-591f970eefbb"),
			Expected: `
    resource(name="howett.net/plist",
             git="https://gitlab.howett.net/go/plist",
             commit="591f970eefbb",
             destination=".",
             placement="vendor/howett.net/plist")`,
			ExpectedError: "",
		},
	}
	for _, c := range testCases {
		r, err := Resource(c.AppVersion, c.Tuple)
		if err != nil {
			if err.Error() != c.ExpectedError {
				t.Error(err)
			}
		}
		if c.ExpectedError != "" {
			t.Errorf("Did not get expected error: %s", c.ExpectedError)
		}
		if r != c.Expected {
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
