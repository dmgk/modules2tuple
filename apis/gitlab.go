package apis

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type GitlabCommit struct {
	ID string `json:"id"`
}

func GetGitlabCommit(site, account, project, commit string) (*GitlabCommit, error) {
	projectID := url.PathEscape(fmt.Sprintf("%s/%s", account, project))
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits/%s", site, projectID, commit)

	resp, err := get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting commit %s for %s/%s: %v", commit, account, project, err)
	}

	var ret GitlabCommit
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	return &ret, nil
}
