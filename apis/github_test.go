// +build online

package apis

import "testing"

func TestGetGithubCommit(t *testing.T) {
	examples := []struct {
		site, account, project, ref, ID string
	}{
		{"https://gitlab.com", "dmgk", "modules2tuple", "v1.9.0", "fc09878b93db35aafc74311f7ea6684ac08a3b83"},
		{"https://gitlab.com", "dmgk", "modules2tuple", "a0cdb416ca2c", "a0cdb416ca2cbf6d3dad67a97f4fdcfac954503e"},
	}

	for i, x := range examples {
		c, err := GetGithubCommit(x.account, x.project, x.ref)
		if err != nil {
			t.Fatal(err)
		}
		if x.ID != c.ID {
			t.Errorf("expected commit ID %s, got %s (example %d)", x.ID, c.ID, i)
		}
	}
}
