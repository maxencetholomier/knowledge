package httpclient

import (
	"fmt"
	"io"
	"net/http"
)

func Get(url string, detector PlatformDetector) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, WrapConnectionError(fmt.Errorf("HTTP GET request failed for URL %s: %w", url, err), url, detector)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return body, nil
}

func Delete(url string, detector PlatformDetector) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return WrapConnectionError(err, url, detector)
	}
	defer resp.Body.Close()

	return nil
}

func Send(method string, url string, body io.Reader, contentType string, context string, detector PlatformDetector) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, FormatHTTPError(err, context, "failed to create request")
	}

	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, WrapConnectionError(FormatHTTPError(err, context, "HTTP request failed"), url, detector)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		respBody, _ := io.ReadAll(resp.Body)
		if context != "" {
			return nil, fmt.Errorf("API error for %s (status: %s): %s", context, resp.Status, string(respBody))
		}
		return nil, fmt.Errorf("API error (status: %s): %s", resp.Status, string(respBody))
	}

	return nil, nil
}
