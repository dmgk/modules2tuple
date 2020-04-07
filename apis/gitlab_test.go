// +build online

package apis

import "testing"

func TestGetGitlabCommit(t *testing.T) {
	examples := []struct {
		site, account, project, ref, hash string
	}{
		{"https://gitlab.com", "gitlab-org", "gitaly-proto", "v1.32.0", "f4db5d05d437abe1154d7308ca044d3577b5ccba"},
		{"https://gitlab.com", "gitlab-org", "labkit", "0c3fc7cdd57c", "0c3fc7cdd57c57da5ab474aa72b6640d2bdc9ebb"},
	}

	for i, x := range examples {
		hash, err := GetGitlabCommit(x.site, x.account, x.project, x.ref)
		if err != nil {
			t.Fatal(err)
		}
		if x.hash != hash {
			t.Errorf("expected commit hash %s, got %s (example %d)", x.hash, hash, i)
		}
	}
}
