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

func HasGithubTag(account, project, tag string) (bool, error) {
	projectID := fmt.Sprintf("%s/%s", url.PathEscape(account), url.PathEscape(project))
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags/%s", projectID, tag)

	resp, err := get(url, flags.GithubCredentialsKey)
	if err != nil {
		if err == errNotFound {
			return false, nil
		}
		if strings.Contains(err.Error(), "API rate limit exceeded") {
			return false, errors.New(githubRateLimitError)
		}
		return false, fmt.Errorf("error getting refs for %s/%s: %v", account, project, err)
	}

	var ref GithubRef
	if err := json.Unmarshal(resp, &ref); err != nil {
		return false, fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	return true, nil
}

func ListGithubTags(account, project, tag string) ([]string, error) {
	projectID := fmt.Sprintf("%s/%s", url.PathEscape(account), url.PathEscape(project))
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags", projectID)

	resp, err := get(url, flags.GithubCredentialsKey)
	if err != nil {
		if strings.Contains(err.Error(), "API rate limit exceeded") {
			return nil, errors.New(githubRateLimitError)
		}
		return nil, fmt.Errorf("error getting refs for %s/%s: %v", account, project, err)
	}

	var refs []GithubRef
	if err := json.Unmarshal(resp, &refs); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	var res []string
	for _, r := range refs {
		res = append(res, r.Ref)
	}

	return res, nil
}

func LookupGithubTag(account, project, path, tag string) (string, error) {
	hasTag, err := HasGithubTag(account, project, tag)
	if err != nil {
		return "", err
	}

	// tag was found as-is
	if hasTag {
		return tag, nil
	}

	// tag was not found, try to look it up
	allTags, err := ListGithubTags(account, project, tag)
	if err != nil {
		return "", err
	}

	// Github API returns tags sorted by creation time, earliest first.
	// Iterate through them in reverse order to find the most recent matching tag.
	for i := len(allTags) - 1; i >= 0; i-- {
		if strings.HasSuffix(allTags[i], path+"/"+tag) {
			return strings.TrimPrefix(allTags[i], "refs/tags/"), nil
		}
	}

	return "", fmt.Errorf("tag %v doesn't seem to exist in %s/%s", tag, account, project)
}

func HasGithubContentsAtPath(account, project, path, tag string) (bool, error) {
	projectID := fmt.Sprintf("%s/%s", url.PathEscape(account), url.PathEscape(project))
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", projectID, path, tag)

	// Ignore reponse, we care only about errors
	_, err := get(url, flags.GithubCredentialsKey)
	if err != nil && err != errNotFound {
		return false, err
	}
	return err == nil, nil
}
