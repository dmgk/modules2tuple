package tuple

import "testing"

func testExamples(t *testing.T, name string, fn vanityParser, examples [][]string) {
	for i, x := range examples {
		tuple := fn(x[0], "vendor")
		if tuple == nil {
			t.Fatalf("%s: expected %q to match", name, x[0])
		}
		if tuple.Account != x[1] {
			t.Errorf("%s: expected account to be %q, got %q (example %d)", name, x[1], tuple.Account, i)
		}
		if tuple.Project != x[2] {
			t.Errorf("%s: expected project to be %q, got %q (example %d)", name, x[2], tuple.Project, i)
		}
	}
}

func TestCloudGoogleCom(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"cloud.google.com/go", "googleapis", "google-cloud-go"},
		{"cloud.google.com/go/storage", "googleapis", "google-cloud-go"},
	}
	testExamples(t, "cloudGoogleComParser", cloudGoogleComParser, examples)
}

func TestParseCodeCloudfoundryOrg(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"code.cloudfoundry.org/gofileutils", "cloudfoundry", "gofileutils"},
	}
	testExamples(t, "codeCloudfoundryOrgParser", codeCloudfoundryOrgParser, examples)
}

func TestParseGoEtcdIoParser(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"go.etcd.io/bbolt", "etcd-io", "bbolt"},
	}
	testExamples(t, "goEtcdIoParser", goEtcdIoParser, examples)
}

func TestParseGopkgInName(t *testing.T) {
	examples := [][]string{
		// pkg, expected account, expected project
		{"gopkg.in/pkg.v3", "go-pkg", "pkg"},
		{"gopkg.in/user/pkg.v3", "user", "pkg"},
		{"gopkg.in/pkg-with-dashes.v3", "go-pkg-with-dashes", "pkg-with-dashes"},
		{"gopkg.in/UserCaps-With-Dashes/0andCrazyPkgName.v3-alpha", "UserCaps-With-Dashes", "0andCrazyPkgName"},
	}
	testExamples(t, "gopkgInParser", gopkgInParser, examples)
}

func TestParseGolangOrgName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"golang.org/x/pkg", "golang", "pkg"},
		{"golang.org/x/oauth2", "golang", "oauth2"},
	}
	testExamples(t, "golangOrgParser", golangOrgParser, examples)
}

func TestParseK8sIoName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"k8s.io/api", "kubernetes", "api"},
	}
	testExamples(t, "k8sIoParser", k8sIoParser, examples)
}

func TestParseGoUberOrgName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"go.uber.org/zap", "uber-go", "zap"},
	}
	testExamples(t, "goUberOrgParser", goUberOrgParser, examples)
}

func TestParseGoMozillaOrgName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"go.mozilla.org/gopgagent", "mozilla-services", "gopgagent"},
	}
	testExamples(t, "goMozillaOrgParser", goMozillaOrgParser, examples)
}

func TestParseMvdanCcName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"mvdan.cc/editorconfig", "mvdan", "editorconfig"},
	}
	testExamples(t, "mvdanCcParser", mvdanCcParser, examples)
}
