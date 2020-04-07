package apis

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/dmgk/modules2tuple/flags"
)

type GithubCommit struct {
	SHA string `json:"sha"`
}

type GithubRef struct {
	Ref string `json:"ref"`
}

var githubRateLimitError = fmt.Sprintf(`Github API rate limit exceeded. Please either:
- set %s environment variable to your Github "username:personal_access_token"
  to let modules2tuple call Github API using basic authentication.
  To create a new token, navigate to https://github.com/settings/tokens/new
  (leave all checkboxes unchecked, modules2tuple doesn't need any access to your account)
- set %s=0 and/or remove "-ghtags" flag to turn off Github tags lookup
- set %s=1 or pass "-offline" flag to module2tuple to disable network access`,
	flags.GithubCredentialsKey, flags.LookupGithubTagsKey, flags.OfflineKey)

func GetGithubCommit(account, project, tag string) (string, error) {
	projectID := fmt.Sprintf("%s/%s", url.PathEscape(account), url.PathEscape(project))
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", projectID, tag)

	resp, err := get(url, flags.GithubCredentialsKey)
	if err != nil {
		if strings.Contains(err.Error(), "API rate limit exceeded") {
			return "", errors.New(githubRateLimitError)
		}
		return "", fmt.Errorf("error getting commit %s for %s/%s: %v", tag, account, project, err)
	}

	var res GithubCommit
	if err := json.Unmarshal(resp, &res); err != nil {
		return "", fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	return res.SHA, nil
}

func LookupGithubTag(account, project, tag string) (string, error) {
	projectID := fmt.Sprintf("%s/%s", url.PathEscape(account), url.PathEscape(project))
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags", projectID)

	resp, err := get(url, flags.GithubCredentialsKey)
	if err != nil {
		if strings.Contains(err.Error(), "API rate limit exceeded") {
			return "", errors.New(githubRateLimitError)
		}
		return "", fmt.Errorf("error getting refs for %s/%s: %v", account, project, err)
	}

	var res []GithubRef
	if err := json.Unmarshal(resp, &res); err != nil {
		return "", fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	// Github API returns tags sorted by creation time, earliest first.
	// Iterate through them in reverse order to find the most recent matching tag.
	for i := len(res) - 1; i >= 0; i-- {
		if strings.HasSuffix(res[i].Ref, "/"+tag) {
			return strings.TrimPrefix(res[i].Ref, "refs/tags/"), nil
		}
	}

	return "", fmt.Errorf("tag %v doesn't seem to exist in %s/%s", tag, account, project)
}
