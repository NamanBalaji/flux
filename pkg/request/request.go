package request

import (
	"fmt"
	"io"
	"net/http"
)

func SendHTTPRequest(method string, url string, body io.Reader) (io.Reader, int, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting response: %w", err)
	}

	return resp.Body, resp.StatusCode, nil
}
