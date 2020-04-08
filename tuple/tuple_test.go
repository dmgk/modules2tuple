package tuple

import (
	"strings"
	"testing"
)

func TestParseRegularSpec(t *testing.T) {
	examples := [][]string{
		// spec, expected package, expected version

		// version is tag
		{"github.com/pkg/errors v1.0.0", "github.com/pkg/errors", "v1.0.0"},
		{"github.com/pkg/errors v1.0.0+incompatible", "github.com/pkg/errors", "v1.0.0"},
		{"github.com/pkg/errors v1.0.0-rc.1.2.3", "github.com/pkg/errors", "v1.0.0-rc.1.2.3"},
		{"github.com/pkg/errors v1.2.3-alpha", "github.com/pkg/errors", "v1.2.3-alpha"},
		{"github.com/pkg/errors v1.2.3-alpha+incompatible", "github.com/pkg/errors", "v1.2.3-alpha"},

		// version is commit ID
		{"github.com/pkg/errors v0.0.0-20181001143604-e0a95dfd547c", "github.com/pkg/errors", "e0a95dfd547c"},
		{"github.com/pkg/errors v1.2.3-20150716171945-2caba252f4dc", "github.com/pkg/errors", "2caba252f4dc"},
		{"github.com/pkg/errors v1.2.3-0.20150716171945-2caba252f4dc", "github.com/pkg/errors", "2caba252f4dc"},
		{"github.com/pkg/errors v1.2.3-42.20150716171945-2caba252f4dc", "github.com/pkg/errors", "2caba252f4dc"},
		{"github.com/docker/libnetwork v0.8.0-dev.2.0.20180608203834-19279f049241", "github.com/docker/libnetwork", "19279f049241"},
		{"github.com/hjson/hjson-go v3.0.1-0.20190209023717-9147687966d9+incompatible", "github.com/hjson/hjson-go", "9147687966d9"},

		// filesystem replacement spec
		{"/some/path", "/some/path", ""},
		{"./relative/path", "./relative/path", ""},
	}

	for i, x := range examples {
		pkg, version, err := parseSpec(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if pkg != x[1] {
			t.Errorf("(%d) expected package to be %q, got %q", i, x[1], pkg)
		}
		if version != x[2] {
			t.Errorf("(%d) expected version to be %q, got %q", i, x[1], version)
		}
	}
}

func TestParseRegularSpecFail(t *testing.T) {
	examples := [][]string{
		// spec, expected error

		// missing version
		{"github.com/pkg/errors", "unexpected spec format"},

		// extra stuff
		{"github.com/pkg/errors v1.0.0 v2.0.0", "unexpected number of fields"},

		// unparseable version
		{"github.com/pkg/errors 012345", "unexpected version string"},
	}

	for i, x := range examples {
		_, _, err := parseSpec(x[0])
		if err == nil {
			t.Errorf("(%d) expected to fail: %q", i, x[0])
		}
		if !strings.HasPrefix(err.Error(), x[1]) {
			t.Errorf("(%d) expected error to start with %q, got %q", i, x[1], err.Error())
		}
	}
}

func TestStringer(t *testing.T) {
	examples := [][]string{
		// spec, expected String()
		{"github.com/pkg/errors v1.0.0", "pkg:errors:v1.0.0:pkg_errors/vendor/github.com/pkg/errors"},
		{"github.com/pkg/errors v0.0.0-20181001143604-e0a95dfd547c", "pkg:errors:e0a95dfd547c:pkg_errors/vendor/github.com/pkg/errors"},
		{"github.com/pkg/errors v1.0.0-rc.1.2.3", "pkg:errors:v1.0.0-rc.1.2.3:pkg_errors/vendor/github.com/pkg/errors"},
		{"github.com/pkg/errors v0.12.3-0.20181001143604-e0a95dfd547c", "pkg:errors:e0a95dfd547c:pkg_errors/vendor/github.com/pkg/errors"},
		{"github.com/UserName/project-with-dashes v1.1.1", "UserName:project-with-dashes:v1.1.1:username_project_with_dashes/vendor/github.com/UserName/project-with-dashes"},
	}

	for i, x := range examples {
		tuple, err := Parse(x[0])
		if err != nil {
			t.Fatal(err)
		}
		s := tuple.String()
		if s != x[1] {
			t.Errorf("(%d) expected String() to return %q, got %q", i, x[1], s)
		}
	}
}

func TestPackageReplace(t *testing.T) {
	examples := [][]string{
		// spec, expected replaced package String()
		{"github.com/spf13/cobra v0.0.0-20180412120829-615425954c3b => github.com/rsteube/cobra v0.0.1-zsh-completion-custom", "rsteube:cobra:v0.0.1-zsh-completion-custom:rsteube_cobra/vendor/github.com/spf13/cobra"},
		{"github.com/hashicorp/vault/api v1.0.5-0.20200215224050-f6547fa8e820 => ./api", "hashicorp:vault:f6547fa8e820:hashicorp_vault"},
	}

	for i, x := range examples {
		tuple, err := Parse(x[0])
		if err != nil {
			t.Fatal(err)
		}
		s := tuple.String()
		if s != x[1] {
			t.Errorf("(%d) expected replaced package String() to return\n%s\ngot\n%s", i, x[1], s)
		}
	}
}

func TestPackageReplaceFail(t *testing.T) {
	examples := [][]string{
		// spec, expected error

		// // missing version in the left spec
		// {"github.com/hashicorp/consul/api => ./api", "unexpected spec format"},

		// missing version in the right spec
		{"github.com/spf13/cobra v0.0.0-20180412120829-615425954c3b => github.com/rsteube/cobra", "unexpected spec format"},
	}

	for i, x := range examples {
		_, err := Parse(x[0])
		if err == nil {
			t.Fatal("expected err to not be nil")
		}
		if !strings.HasPrefix(err.Error(), x[1]) {
			t.Errorf("(%d) expected err to start with %q, got %q", i, x[1], err.Error())
		}
	}
}
