// +build online

package gitlab

import "testing"

func TestGetCommit(t *testing.T) {
	examples := []struct {
		account, project, tag, commitID string
	}{
		{"gitlab-org", "gitaly-proto", "v1.32.0", "f4db5d05d437abe1154d7308ca044d3577b5ccba"},
		{"gitlab-org", "labkit", "0c3fc7cdd57c", "0c3fc7cdd57c57da5ab474aa72b6640d2bdc9ebb"},
	}

	for i, x := range examples {
		c, err := GetCommit(x.account, x.project, x.tag)
		if err != nil {
			t.Fatal(err)
		}
		if x.commitID != c.ID {
			t.Errorf("expected commitID %s, got %s (example %d)", x.commitID, c.ID, i)
		}
	}
}
