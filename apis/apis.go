package apis

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	OfflineKey           = "M2T_OFFLINE"
	GithubCredentialsKey = "M2T_GITHUB"
	// GitlabsCredentialsKey = "M2T_GITLAB"
)

func get(url string, credsKey string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if credsKey != "" {
		creds := os.Getenv(credsKey)
		if creds != "" {
			credsSlice := strings.Split(creds, ":")
			if len(credsSlice) == 2 {
				req.SetBasicAuth(credsSlice[0], credsSlice[1])
			}
		}
	}

	resp, err := http.DefaultClient.Do(req)
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
