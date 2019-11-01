package apis

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type GithubCommit struct {
	ID string `json:"sha"`
}

func GetGithubCommit(account, project, ref string) (*GithubCommit, error) {
	projectID := fmt.Sprintf("%s/%s", url.PathEscape(account), url.PathEscape(project))
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", projectID, ref)

	resp, err := get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting commit %s for %s/%s: %v", ref, account, project, err)
	}

	var ret GithubCommit
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	return &ret, nil
}
