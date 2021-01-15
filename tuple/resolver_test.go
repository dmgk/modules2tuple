package tuple

import (
	"fmt"
	"testing"

	"github.com/dmgk/modules2tuple/v2/config"
)

func init() {
	config.Offline = true
}

type resolverExample struct {
	pkg                      string
	source                   Source
	account, project, module string
}

func TestGithubResolver(t *testing.T) {
	examples := []resolverExample{
		{"github.com/pkg/errors", GH, "pkg", "errors", ""},
		{"github.com/konsorten/go-windows-terminal-sequences", GH, "konsorten", "go-windows-terminal-sequences", ""},
		{"github.com/account/project/v2", GH, "account", "project", "v2"},
		{"github.com/account/project/api/client", GH, "account", "project", "api/client"},
	}

	for i, x := range examples {
		m, err := githubResolver(x.pkg)
		if err != nil {
			t.Fatal(err)
		}
		if m == nil {
			t.Fatalf("(%d): expected %q to match", i, x.pkg)
		}
		if fmt.Sprintf("%T %v", m.source, m.source) != fmt.Sprintf("%T %v", x.source, x.source) {
			t.Errorf("(%d) expected source to be %q, got %q", i, fmt.Sprintf("%T %v", x.source, x.source), fmt.Sprintf("%T %v", m.source, m.source))
		}
		if m.account != x.account {
			t.Errorf("(%d) expected account to be %q, got %q", i, x.account, m.account)
		}
		if m.project != x.project {
			t.Errorf("(%d) expected project to be %q, got %q", i, x.project, m.project)
		}
	}
}

func TestGitlabResolver(t *testing.T) {
	examples := []resolverExample{
		{"gitlab.com/account/project", GL, "account", "project", ""},
		{"gitlab.com/account/project/api/client", GL, "account", "project", "api/client"},
	}

	for i, x := range examples {
		m, err := gitlabResolver(x.pkg)
		if err != nil {
			t.Fatal(err)
		}
		if m == nil {
			t.Fatalf("(%d): expected %q to match", i, x.pkg)
		}
		if fmt.Sprintf("%T %v", m.source, m.source) != fmt.Sprintf("%T %v", x.source, x.source) {
			t.Errorf("(%d) expected source to be %q, got %q", i, fmt.Sprintf("%T %v", x.source, x.source), fmt.Sprintf("%T %v", m.source, m.source))
		}
		if m.account != x.account {
			t.Errorf("(%d) expected account to be %q, got %q", i, x.account, m.account)
		}
		if m.project != x.project {
			t.Errorf("(%d) expected project to be %q, got %q", i, x.project, m.project)
		}
	}
}

func TestMirrorResolver(t *testing.T) {
	examples := []resolverExample{
		{"camlistore.org", GH, "perkeep", "perkeep", ""},
		{"docker.io/go-docker", GH, "docker", "go-docker", ""},
		{"git.apache.org/thrift.git", GH, "apache", "thrift", ""},
		{"go.bug.st/serial.v1", GH, "bugst", "go-serial", ""},
		{"go.elastic.co/apm", GH, "elastic", "apm-agent-go", ""},
		{"go.elastic.co/fastjson", GH, "elastic", "go-fastjson", ""},
		{"go.mongodb.org/mongo-driver", GH, "mongodb", "mongo-go-driver", ""},
		{"go.opencensus.io", GH, "census-instrumentation", "opencensus-go", ""},
		{"go4.org", GH, "go4org", "go4", ""},
		{"gocloud.dev", GH, "google", "go-cloud", ""},
		{"google.golang.org/api", GH, "googleapis", "google-api-go-client", ""},
		{"google.golang.org/appengine", GH, "golang", "appengine", ""},
		{"google.golang.org/genproto", GH, "google", "go-genproto", ""},
		{"google.golang.org/grpc", GH, "grpc", "grpc-go", ""},
		{"gotest.tools", GH, "gotestyourself", "gotest.tools", ""},
		{"honnef.co/go/tools", GH, "dominikh", "go-tools", ""},
		{"howett.net/plist", GitlabSource("https://gitlab.howett.net"), "go", "plist", ""},
		{"layeh.com/radius", GH, "layeh", "radius", ""},
		{"sigs.k8s.io/yaml", GH, "kubernetes-sigs", "yaml", ""},
		{"tinygo.org/x/go-llvm", GH, "tinygo-org", "go-llvm", ""},
	}

	for i, x := range examples {
		m, err := Resolve(x.pkg, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if m == nil {
			t.Fatalf("(%d): expected %q to match", i, x.pkg)
		}
		if fmt.Sprintf("%T %v", m.source, m.source) != fmt.Sprintf("%T %v", x.source, x.source) {
			t.Errorf("(%d) expected source to be %q, got %q", i, fmt.Sprintf("%T %v", x.source, x.source), fmt.Sprintf("%T %v", m.source, m.source))
		}
		if m.account != x.account {
			t.Errorf("(%d) expected account to be %q, got %q", i, x.account, m.account)
		}
		if m.project != x.project {
			t.Errorf("(%d) expected project to be %q, got %q", i, x.project, m.project)
		}
	}
}

func testResolverFnExamples(t *testing.T, name string, fn mirrorFn, examples []resolverExample) {
	for _, x := range examples {
		m, err := fn(x.pkg)
		if err != nil {
			t.Fatal(err)
		}
		if m == nil {
			t.Fatalf("(%s): expected %q to match", name, x.pkg)
		}
		if fmt.Sprintf("%T %v", m.source, m.source) != fmt.Sprintf("%T %v", x.source, x.source) {
			t.Errorf("(%s) expected source to be %q, got %q", name, fmt.Sprintf("%T %v", x.source, x.source), fmt.Sprintf("%T %v", m.source, m.source))
		}
		if m.account != x.account {
			t.Errorf("(%s) expected account to be %q, got %q", name, x.account, m.account)
		}
		if m.project != x.project {
			t.Errorf("(%s) expected project to be %q, got %q", name, x.project, m.project)
		}
	}
}

func TestBazilOrgResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"bazil.org/fuse", GH, "bazil", "fuse", ""},
	}
	testResolverFnExamples(t, "bazilOrgResolver", bazilOrgResolver, examples)
}

func TestCloudGoogleComResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"cloud.google.com/go", GH, "googleapis", "google-cloud-go", ""},
		{"cloud.google.com/go/storage", GH, "googleapis", "google-cloud-go", ""},
	}
	testResolverFnExamples(t, "cloudGoogleComResolver", cloudGoogleComResolver, examples)
}

// func TestCodeCloudfoundryOrgResolver(t *testing.T) {
// 	examples := []resolverExample{
// 		// name, expected account, expected project
// 		{"code.cloudfoundry.org/gofileutils", GH, "cloudfoundry", "gofileutils", ""},
// 	}
// 	testResolverFnExamples(t, "codeCloudfoundryOrgResolver", codeCloudfoundryOrgResolver, examples)
// }

func TestGoEtcdIoResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"go.etcd.io/bbolt", GH, "etcd-io", "bbolt", ""},
	}
	testResolverFnExamples(t, "goEtcdIoResolver", goEtcdIoResolver, examples)
}

func TestGopkgInResolver(t *testing.T) {
	examples := []resolverExample{
		// pkg, expected account, expected project
		{"gopkg.in/pkg.v3", GH, "go-pkg", "pkg", ""},
		{"gopkg.in/user/pkg.v3", GH, "user", "pkg", ""},
		{"gopkg.in/pkg-with-dashes.v3", GH, "go-pkg-with-dashes", "pkg-with-dashes", ""},
		{"gopkg.in/UserCaps-With-Dashes/0andCrazyPkgName.v3-alpha", GH, "UserCaps-With-Dashes", "0andCrazyPkgName", ""},
	}
	testResolverFnExamples(t, "gopkgInResolver", gopkgInResolver, examples)
}

func TestGolangOrgResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"golang.org/x/pkg", GH, "golang", "pkg", ""},
		{"golang.org/x/oauth2", GH, "golang", "oauth2", ""},
	}
	testResolverFnExamples(t, "golangOrgResolver", golangOrgResolver, examples)
}

func TestK8sIoResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"k8s.io/api", GH, "kubernetes", "api", ""},
	}
	testResolverFnExamples(t, "k8sIoResolver", k8sIoResolver, examples)
}

func TestGoUberOrgResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"go.uber.org/zap", GH, "uber-go", "zap", ""},
	}
	testResolverFnExamples(t, "goUberOrgResolver", goUberOrgResolver, examples)
}

func TestGoMozillaOrgResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"go.mozilla.org/gopgagent", GH, "mozilla-services", "gopgagent", ""},
	}
	testResolverFnExamples(t, "goMozillaOrgResolver", goMozillaOrgResolver, examples)
}

func TestMvdanCcResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"mvdan.cc/editorconfig", GH, "mvdan", "editorconfig", ""},
	}
	testResolverFnExamples(t, "mvdanCcResolver", mvdanCcResolver, examples)
}

func TestRscIoResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"rsc.io/pdf", GH, "rsc", "pdf", ""},
	}
	testResolverFnExamples(t, "rscIoResolver", rscIoResolver, examples)
}

func TestGotestToolsResolver(t *testing.T) {
	examples := []resolverExample{
		// name, expected account, expected project
		{"gotest.tools", GH, "gotestyourself", "gotest.tools", ""},
		{"gotest.tools/gotestsum", GH, "gotestyourself", "gotestsum", ""},
	}
	testResolverFnExamples(t, "gotestToolsResolver", gotestToolsResolver, examples)
}
