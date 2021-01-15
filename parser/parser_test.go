package parser

import (
	"strings"
	"testing"

	"github.com/dmgk/modules2tuple/v2/config"
)

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
		ugorji:go:e444a5086c43:ugorji_go_codec/vendor/github.com/ugorji/go \
		user:pkg:v3.0.0:user_pkg/vendor/gopkg.in/user/pkg.v3

GL_TUPLE=	gitlab-org:gitaly-proto:v1.32.0:gitlab_org_gitaly_proto/vendor/gitlab.com/gitlab-org/gitaly-proto \
		gitlab-org:labkit:0c3fc7cdd57c:gitlab_org_labkit/vendor/gitlab.com/gitlab-org/labkit

		# Mirrors for the following packages are not currently known, please look them up and handle these tuples manually:
		#	::v1.0.0:group_name/vendor/another.vanity_url.org/account/project (from another.vanity_url.org/account/project@v1.0.0)
		#	::v1.2.3:group_name/vendor/some_unknown.vanity_url.net/account/project (from some_unknown.vanity_url.net/account/project@v1.2.3)`

	config.Offline = true
	res, err := Read(strings.NewReader(given))
	if err != nil {
		t.Fatal(err)
	}
	out := res.String()
	if out != expected {
		t.Errorf("expected output\n%q\n, got\n%q\n", expected, out)
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
		minio:minio-go:v6.0.39:minio_minio_go_v6/vendor/github.com/minio/minio-go/v6 \
		minio:parquet-go:9d767baf1679:minio_parquet_go/vendor/github.com/minio/parquet-go`

	config.Offline = true
	res, err := Read(strings.NewReader(given))
	if err != nil {
		t.Fatal(err)
	}
	out := res.String()
	if out != expected {
		t.Errorf("expected output\n%q\n, got\n%q\n", expected, out)
	}
}
