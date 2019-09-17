package vanity

import "testing"

func testExamples(t *testing.T, p Parser, examples [][]string) {
	for i, x := range examples {
		if !p.Match(x[0]) {
			t.Fatalf("%s parser: expected %q to match", p.Name(), x[0])
		}
		account, project := p.Parse(x[0])
		if account != x[1] {
			t.Errorf("%s parser: expected account to be %q, got %q (example %d)", p.Name(), x[1], account, i)
		}
		if project != x[2] {
			t.Errorf("%s parser: expected project to be %q, got %q (example %d)", p.Name(), x[2], project, i)
		}
	}
}

func TestParseGopkgInName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"gopkg.in/pkg.v3", "go-pkg", "pkg"},
		{"gopkg.in/user/pkg.v3", "user", "pkg"},
		{"gopkg.in/pkg-with-dashes.v3", "go-pkg-with-dashes", "pkg-with-dashes"},
		{"gopkg.in/UserCaps-With-Dashes/0andCrazyPkgName.v3-alpha", "UserCaps-With-Dashes", "0andCrazyPkgName"},
	}
	testExamples(t, newGopkgInParser(), examples)
}

func TestParseGolangOrgName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"golang.org/x/pkg", "golang", "pkg"},
		{"golang.org/x/oauth2", "golang", "oauth2"},
	}
	testExamples(t, newGolangOrgParser(), examples)
}

func TestParseK8sIoName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"k8s.io/api", "kubernetes", "api"},
	}
	testExamples(t, newK8sIoParser(), examples)
}

func TestParseGoUberOrgName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"go.uber.org/zap", "uber-go", "zap"},
	}
	testExamples(t, newGoUberOrgParser(), examples)
}

func TestParseGoMozillaOrgName(t *testing.T) {
	examples := [][]string{
		// name, expected account, expected project
		{"go.mozilla.org/gopgagent", "mozilla-services", "gopgagent"},
	}
	testExamples(t, newGoMozillaOrgParser(), examples)
}
