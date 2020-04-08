// +build online

package apis

import "testing"

func TestGithubGetCommit(t *testing.T) {
	examples := []struct {
		account, project, ref, hash string
	}{
		{"dmgk", "modules2tuple", "v1.9.0", "fc09878b93db35aafc74311f7ea6684ac08a3b83"},
		{"dmgk", "modules2tuple", "a0cdb416ca2c", "a0cdb416ca2cbf6d3dad67a97f4fdcfac954503e"},
	}

	for i, x := range examples {
		hash, err := GithubGetCommit(x.account, x.project, x.ref)
		if err != nil {
			t.Fatal(err)
		}
		if x.hash != hash {
			t.Errorf("expected commit hash %s, got %s (example %d)", x.hash, hash, i)
		}
	}
}

func TestGithubLookupTag(t *testing.T) {
	examples := []struct {
		account, project, packageSuffix, given, expected string
	}{
		// tag exists as-is
		{"hashicorp", "vault", "", "v1.3.4", "v1.3.4"},
		// tag exists but with prefix
		{"hashicorp", "vault", "api", "v1.0.4", "api/v1.0.4"},
		{"hashicorp", "vault", "sdk", "v0.1.13", "sdk/v0.1.13"},
		// this repo has earlier mathing tag "codec/codecgen/v1.1.7"
		{"ugorji", "go", "", "v1.1.7", "v1.1.7"},
	}

	for i, x := range examples {
		tag, err := GithubLookupTag(x.account, x.project, x.packageSuffix, x.given)
		if err != nil {
			t.Fatal(err)
		}
		if x.expected != tag {
			t.Errorf("expected tag %s, got %s (example %d)", x.expected, tag, i)
		}
	}
}
