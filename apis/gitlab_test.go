//go:build online
// +build online

package apis

import "testing"

func TestGitlabGetCommit(t *testing.T) {
	examples := []struct {
		site, account, project, ref, sha string
	}{
		{"https://gitlab.com", "gitlab-org", "gitaly-proto", "v1.32.0", "f4db5d05d437abe1154d7308ca044d3577b5ccba"},
		{"https://gitlab.com", "gitlab-org", "labkit", "0c3fc7cdd57c", "0c3fc7cdd57c57da5ab474aa72b6640d2bdc9ebb"},
	}

	for i, x := range examples {
		sha, err := GitlabGetCommit(x.site, x.account, x.project, x.ref)
		if err != nil {
			t.Fatal(err)
		}
		if x.sha != sha {
			t.Errorf("expected commit hash %s, got %s (example %d)", x.sha, sha, i)
		}
	}
}
