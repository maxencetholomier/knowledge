package httpclient

import (
	"fmt"
	"net/url"
	"strings"
)

func FormatHTTPError(err error, context string, message string) error {
	if context != "" {
		return fmt.Errorf("%s for %s: %w", message, context, err)
	}
	return fmt.Errorf("%s: %w", message, err)
}

func WrapConnectionError(err error, rawURL string) error {
	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		return err
	}
	host := rawURL
	if parsed, parseErr := url.Parse(rawURL); parseErr == nil {
		host = parsed.Host
	}
	return fmt.Errorf("cannot connect to %s, make sure the service is running and reachable", host)
}
