package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Commit struct {
	ID string `json:"id"`
}

func GetCommit(account, project, commit string) (*Commit, error) {
	projectID := url.PathEscape(fmt.Sprintf("%s/%s", account, project))
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/repository/commits/%s", projectID, commit)

	resp, err := get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting commit %s for %s/%s: %v", commit, account, project, err)
	}

	var ret Commit
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %v, resp: %v", err, string(resp))
	}

	return &ret, nil
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("GET %s: %v", url, err)
		}
		return body, nil
	default:
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s: %d, body: %v", url, resp.StatusCode, string(body))
	}
}
