package apis

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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
