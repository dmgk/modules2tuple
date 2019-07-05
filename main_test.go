package main

import "testing"

func TestParseName(t *testing.T) {
	examples := [][]string{
		// spec, expected account, expected project

		// Github
		{"github.com/pkg/errors v1.0.0", "pkg", "errors"},
		{"github.com/konsorten/go-windows-terminal-sequences v1.1.1", "konsorten", "go-windows-terminal-sequences"},

		// gopkg.in
		{"gopkg.in/yaml.v2 v2.0.0", "go-yaml", "yaml"},
		{"gopkg.in/op/go-logging.v1 v1.0.0", "op", "go-logging"},
		// gopkg.in/fsnotify is a special case
		{"gopkg.in/fsnotify.v1 v1.0.0", "fsnotify", "fsnotify"},

		// golang.org
		{"golang.org/x/crypto v1.0.0", "golang", "crypto"},
		{"golang.org/x/text v0.3.0", "golang", "text"},

		// k8s.io
		{"k8s.io/api v1.0.0", "kubernetes", "api"},
		{"k8s.io/client-go v2.0.0", "kubernetes", "client-go"},

		// go.uber.org
		{"go.uber.org/zap v1.10.0", "uber-go", "zap"},

		// Other known mirrors
		{"google.golang.org/api v1.0.0", "googleapis", "google-api-go-client"},
	}

	for i, x := range examples {
		tuple, err := parseTuple(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if tuple.account != x[1] {
			t.Errorf("(%d) expected account to be %q, got %q", i, x[1], tuple.account)
		}
		if tuple.project != x[2] {
			t.Errorf("(%d) expected project to be %q, got %q", i, x[2], tuple.project)
		}
	}
}

func TestParseVersion(t *testing.T) {
	examples := [][]string{
		// spec, expected tag
		{"github.com/pkg/errors v1.0.0", "v1.0.0"},
		{"github.com/pkg/errors v1.0.0+incompatible", "v1.0.0"},
		{"github.com/pkg/errors v1.0.0-rc.1.2.3", "v1.0.0-rc.1.2.3"},
		{"github.com/pkg/errors v1.2.3-alpha", "v1.2.3-alpha"},
		{"github.com/pkg/errors v1.2.3-alpha+incompatible", "v1.2.3-alpha"},
	}

	for i, x := range examples {
		tuple, err := parseTuple(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if tuple.tag != x[1] {
			t.Errorf("(%d) expected tag to be %q, got %q", i, x[1], tuple.tag)
		}
	}
}

func TestParseTag(t *testing.T) {
	examples := [][]string{
		// spec, expected tag
		{"github.com/pkg/errors v0.0.0-20181001143604-e0a95dfd547c", "e0a95dfd547c"},
		{"github.com/pkg/errors v1.2.3-20150716171945-2caba252f4dc", "2caba252f4dc"},
		{"github.com/pkg/errors v1.2.3-0.20150716171945-2caba252f4dc", "2caba252f4dc"},
		{"github.com/pkg/errors v1.2.3-42.20150716171945-2caba252f4dc", "2caba252f4dc"},
		{"github.com/docker/libnetwork v0.8.0-dev.2.0.20180608203834-19279f049241", "19279f049241"},
	}

	for i, x := range examples {
		tuple, err := parseTuple(x[0])
		if err != nil {
			t.Fatal(err)
		}
		if tuple.tag != x[1] {
			t.Errorf("(%d) expected tag to be %q, got %q", i, x[1], tuple.tag)
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
		tuple, err := parseTuple(x[0])
		if err != nil {
			t.Fatal(err)
		}
		s := tuple.String()
		if s != x[1] {
			t.Errorf("(%d) expected String() to return %q, got %q", i, x[1], s)
		}
	}
}

func TestPackageRename(t *testing.T) {
	examples := [][]string{
		// spec, expected renamed package String()
		{"github.com/spf13/cobra v0.0.0-20180412120829-615425954c3b => github.com/rsteube/cobra v0.0.1-zsh-completion-custom", "rsteube:cobra:v0.0.1-zsh-completion-custom:rsteube_cobra/vendor/github.com/spf13/cobra"},
	}

	for i, x := range examples {
		tuple, err := parseTuple(x[0])
		if err != nil {
			t.Fatal(err)
		}
		s := tuple.String()
		if s != x[1] {
			t.Errorf("(%d) expected renamed package String() to return %q, got %q", i, x[1], s)
		}
	}
}
