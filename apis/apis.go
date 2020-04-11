package apis

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var errNotFound = errors.New("not found")

func get(url, username, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if username != "" && token != "" {
		req.SetBasicAuth(username, token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apis.get %s: %v", url, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("apis.get %s: %v", url, err)
		}
		return body, nil
	case http.StatusNotFound:
		return nil, errNotFound
	default:
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("apis.get %s: %d, body: %v", url, resp.StatusCode, string(body))
	}
}
