package apis

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type GitlabCommit struct {
	SHA string `json:"id"`
}

func GetGitlabCommit(site, account, project, commit string) (string, error) {
	projectID := url.PathEscape(fmt.Sprintf("%s/%s", account, project))
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits/%s", site, projectID, commit)

	resp, err := get(url, "")
	if err != nil {
		return "", fmt.Errorf("error getting commit %s for %s/%s: %v", commit, account, project, err)
	}

	var res GitlabCommit
	if err := json.Unmarshal(resp, &res); err != nil {
		return "", fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	return res.SHA, nil
}
