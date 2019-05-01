package main

import "testing"

func TestParseName(t *testing.T) {
	examples := [][]string{
		// spec, Account, Project

		// Github
		[]string{"github.com/pkg/errors v1.0.0", "pkg", "errors"},
		[]string{"github.com/konsorten/go-windows-terminal-sequences v1.1.1", "konsorten", "go-windows-terminal-sequences"},
		// Well-known packages
		[]string{"golang.org/x/crypto v1.0.0", "golang", "crypto"},
		[]string{"gopkg.in/yaml.v2 v2.0.0", "go-yaml", "yaml"},
	}

	for i, x := range examples {
		pkg, err := ParsePackage(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if pkg.Account != x[1] {
			t.Errorf("(%d) expected Account to be %q, got %q", i, x[1], pkg.Account)
		}
		if pkg.Project != x[2] {
			t.Errorf("(%d) expected Project to be %q, got %q", i, x[2], pkg.Project)
		}
	}
}

func TestParseVersion(t *testing.T) {
	examples := [][]string{
		// spec, Tag
		[]string{"github.com/pkg/errors v1.0.0", "v1.0.0"},
		[]string{"github.com/pkg/errors v1.0.0+incompatible", "v1.0.0"},
		[]string{"github.com/pkg/errors v1.0.0-rc.1.2.3", "v1.0.0-rc.1.2.3"},
		[]string{"github.com/pkg/errors v1.2.3-alpha", "v1.2.3-alpha"},
		[]string{"github.com/pkg/errors v1.2.3-alpha+incompatible", "v1.2.3-alpha"},
	}

	for i, x := range examples {
		pkg, err := ParsePackage(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if pkg.Tag != x[1] {
			t.Errorf("(%d) expected Tag to be %q, got %q", i, x[1], pkg.Tag)
		}
	}
}

func TestParseTag(t *testing.T) {
	examples := [][]string{
		// spec, Tag
		[]string{"github.com/pkg/errors v0.0.0-20181001143604-e0a95dfd547c", "e0a95df"},
		[]string{"github.com/pkg/errors v1.2.3-20150716171945-2caba252f4dc", "2caba25"},
		[]string{"github.com/pkg/errors v1.2.3-0.20150716171945-2caba252f4dc", "2caba25"},
		[]string{"github.com/pkg/errors v1.2.3-42.20150716171945-2caba252f4dc", "2caba25"},
	}

	for i, x := range examples {
		pkg, err := ParsePackage(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if pkg.Tag != x[1] {
			t.Errorf("(%d) expected Tag to be %q, got %q", i, x[1], pkg.Tag)
		}
	}
}

func TestStringer(t *testing.T) {
	examples := [][]string{
		// spec, String()
		[]string{"github.com/pkg/errors v1.0.0", "pkg:errors:v1.0.0:pkg_errors/vendor/github.com/pkg/errors"},
		[]string{"github.com/pkg/errors v0.0.0-20181001143604-e0a95dfd547c", "pkg:errors:e0a95df:pkg_errors/vendor/github.com/pkg/errors"},
		[]string{"github.com/pkg/errors v1.0.0-rc.1.2.3", "pkg:errors:v1.0.0-rc.1.2.3:pkg_errors/vendor/github.com/pkg/errors"},
		[]string{"github.com/pkg/errors v0.12.3-0.20181001143604-e0a95dfd547c", "pkg:errors:e0a95df:pkg_errors/vendor/github.com/pkg/errors"},
		[]string{"github.com/UserName/project-with-dashes v1.1.1", "UserName:project-with-dashes:v1.1.1:username_project_with_dashes/vendor/github.com/UserName/project-with-dashes"},
	}

	for i, x := range examples {
		pkg, err := ParsePackage(x[0])
		if err != nil {
			t.Fatal(err)
		}
		s := pkg.String()
		if s != x[1] {
			t.Errorf("(%d) expected String() to return %q, got %q", i, x[1], s)
		}
	}
}

func TestPackageRename(t *testing.T) {
	examples := [][]string{
		// spec, renamed package String()
		[]string{"github.com/spf13/cobra v0.0.0-20180412120829-615425954c3b => github.com/rsteube/cobra v0.0.1-zsh-completion-custom", "rsteube:cobra:v0.0.1-zsh-completion-custom:rsteube_cobra/vendor/github.com/spf13/cobra"},
	}

	for i, x := range examples {
		pkg, err := ParsePackage(x[0])
		if err != nil {
			t.Fatal(err)
		}
		s := pkg.String()
		if s != x[1] {
			t.Errorf("(%d) expected renamed package String() to return %q, got %q", i, x[1], s)
		}
	}
}

func TestParseGopkgIn(t *testing.T) {
	examples := [][]string{
		// name, Account, Project
		[]string{"gopkg.in/pkg.v3", "go-pkg", "pkg"},
		[]string{"gopkg.in/user/pkg.v3", "user", "pkg"},
		[]string{"gopkg.in/pkg-with-dashes.v3", "go-pkg-with-dashes", "pkg-with-dashes"},
		[]string{"gopkg.in/UserCaps-With-Dashes/0andCrazyPkgName.v3-alpha", "UserCaps-With-Dashes", "0andCrazyPkgName"},
	}

	for i, x := range examples {
		account, project := parseGopkgInPackage(x[0])
		if account != x[1] {
			t.Errorf("(%d) expected Account to be %q, got %q", i, x[1], account)
		}
		if project != x[2] {
			t.Errorf("(%d) expected Project to be %q, got %q", i, x[2], project)
		}
	}
}

func TestParseGolangOrgIn(t *testing.T) {
	examples := [][]string{
		// name, Account, Project
		[]string{"golang.org/x/pkg", "golang", "pkg"},
		[]string{"golang.org/x/oauth2", "golang", "oauth2"},
	}

	for i, x := range examples {
		account, project := parseGolangOrgPackage(x[0])
		if account != x[1] {
			t.Errorf("(%d) expected Account to be %q, got %q", i, x[1], account)
		}
		if project != x[2] {
			t.Errorf("(%d) expected Project to be %q, got %q", i, x[2], project)
		}
	}
}
