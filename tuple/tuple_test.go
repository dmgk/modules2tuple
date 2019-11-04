package tuple

import (
	"strings"
	"testing"
)

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

		// gocloud.dev
		{"gocloud.dev v0.16.0", "google", "go-cloud"},

		// Other known mirrors
		{"google.golang.org/api v1.0.0", "googleapis", "google-api-go-client"},
	}

	for i, x := range examples {
		tuple, err := Parse(x[0], "")
		if err != nil {
			t.Fatal(err)
		}
		if tuple.Account != x[1] {
			t.Errorf("(%d) expected account to be %q, got %q", i, x[1], tuple.Account)
		}
		if tuple.Project != x[2] {
			t.Errorf("(%d) expected project to be %q, got %q", i, x[2], tuple.Project)
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
		tuple, err := Parse(x[0], "")
		if err != nil {
			t.Fatal(err)
		}
		if tuple.Tag != x[1] {
			t.Errorf("(%d) expected tag to be %q, got %q", i, x[1], tuple.Tag)
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
		{"github.com/hjson/hjson-go v3.0.1-0.20190209023717-9147687966d9+incompatible", "9147687966d9"},
	}

	for i, x := range examples {
		tuple, err := Parse(x[0], "")
		if err != nil {
			t.Fatal(err)
		}
		if tuple.Tag != x[1] {
			t.Errorf("(%d) expected tag to be %q, got %q", i, x[1], tuple.Tag)
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
		tuple, err := Parse(x[0], "vendor")
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
		{"github.com/spf13/cobra => github.com/rsteube/cobra v0.0.1-zsh-completion-custom", "rsteube:cobra:v0.0.1-zsh-completion-custom:rsteube_cobra/vendor/github.com/spf13/cobra"},
	}

	for i, x := range examples {
		tuple, err := Parse(x[0], "vendor")
		if err != nil {
			t.Fatal(err)
		}
		s := tuple.String()
		if s != x[1] {
			t.Errorf("(%d) expected replaced package String() to return %q, got %q", i, x[1], s)
		}
	}
}

func TestPackageReplaceNoVersion(t *testing.T) {
	examples := [][]string{
		// spec, expected error
		{"github.com/hashicorp/consul/api => ./api", "github.com/hashicorp/consul/api => ./api"},
	}

	for i, x := range examples {
		_, err := Parse(x[0], "vendor")
		if err == nil {
			t.Fatal("expected err to not be nil")
		}
		e := err.Error()
		if e != x[1] {
			t.Errorf("(%d) expected err to be %q, got %q", i, x[1], e)
		}
	}
}

func TestReader(t *testing.T) {
	given := `
# github.com/karrick/godirwalk v1.10.12
## explicit
github.com/karrick/godirwalk
# github.com/rogpeppe/go-internal v1.3.0
## explicit
github.com/rogpeppe/go-internal/modfile
github.com/rogpeppe/go-internal/module
github.com/rogpeppe/go-internal/semver
# some_unknown.vanity_url.net/account/project v1.2.3
some_unknown.vanity_url.net/account/project
# another.vanity_url.org/account/project v1.0.0
another.vanity_url.org/account/project
# gopkg.in/user/pkg.v3 v3.0.0
gopkg.in/user/pkg.v3
# github.com/cockroachdb/cockroach-go v0.0.0-20181001143604-e0a95dfd547c
github.com/cockroachdb/cockroach-go/crdb
# gitlab.com/gitlab-org/labkit v0.0.0-20190221122536-0c3fc7cdd57c
gitlab.com/gitlab-org/labkit/correlation
# gitlab.com/gitlab-org/gitaly-proto v1.32.0
gitlab.com/gitlab-org/gitaly-proto/go/gitalypb
# github.com/golang/lint v0.0.0-20190409202823-959b441ac422 => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
# github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43`

	expected := `GH_TUPLE=	\
		cockroachdb:cockroach-go:e0a95dfd547c:cockroachdb_cockroach_go/vendor/github.com/cockroachdb/cockroach-go \
		golang:lint:959b441ac422:golang_lint/vendor/github.com/golang/lint \
		karrick:godirwalk:v1.10.12:karrick_godirwalk/vendor/github.com/karrick/godirwalk \
		rogpeppe:go-internal:v1.3.0:rogpeppe_go_internal/vendor/github.com/rogpeppe/go-internal \
		ugorji:go:e444a5086c43:ugorji_go/vendor/github.com/ugorji/go \
		user:pkg:v3.0.0:user_pkg/vendor/gopkg.in/user/pkg.v3

GL_TUPLE=	\
		gitlab-org:gitaly-proto:v1.32.0:gitlab_org_gitaly_proto/vendor/gitlab.com/gitlab-org/gitaly-proto \
		gitlab-org:labkit:0c3fc7cdd57c:gitlab_org_labkit/vendor/gitlab.com/gitlab-org/labkit
		# Mirrors for the following packages are not currently known, please look them up and handle these tuples manually:
		#	::v1.0.0:group_name/vendor/another.vanity_url.org/account/project
		#	::v1.2.3:group_name/vendor/some_unknown.vanity_url.net/account/project
`

	tt, err := NewParser("vendor", true).Read(strings.NewReader(given))
	if err == nil {
		t.Fatal("expected err to not be nil")
	}
	out := tt.String() + err.Error()
	if out != expected {
		t.Errorf("expected output\n%s, got\n%s", expected, out)
	}
}

func TestUniqueGroups(t *testing.T) {
	given := `
# github.com/minio/lsync v1.0.1
# github.com/minio/mc v0.0.0-20190924013003-643835013047
# github.com/minio/minio-go v0.0.0-20190327203652-5325257a208f
# github.com/minio/minio-go/v6 v6.0.39
# github.com/minio/parquet-go v0.0.0-20190318185229-9d767baf1679`

	expected := `GH_TUPLE=	\
		minio:lsync:v1.0.1:minio_lsync/vendor/github.com/minio/lsync \
		minio:mc:643835013047:minio_mc/vendor/github.com/minio/mc \
		minio:minio-go:5325257a208f:minio_minio_go/vendor/github.com/minio/minio-go \
		minio:minio-go:v6.0.39:minio_minio_go_1/vendor/github.com/minio/minio-go/v6 \
		minio:parquet-go:9d767baf1679:minio_parquet_go/vendor/github.com/minio/parquet-go
`

	tt, err := NewParser("vendor", true).Read(strings.NewReader(given))
	if err != nil {
		t.Fatal(err)
	}
	out := tt.String()
	if out != expected {
		t.Errorf("expected output %s, got %s", expected, out)
	}
}
