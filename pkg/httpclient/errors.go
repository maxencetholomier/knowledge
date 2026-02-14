package httpclient

import (
	"fmt"
	"strings"
)

type PlatformDetector interface {
	DetectPlatform(url string) string
}

func FormatHTTPError(err error, context string, message string) error {
	if context != "" {
		return fmt.Errorf("%s for %s: %w", message, context, err)
	}
	return fmt.Errorf("%s: %w", message, err)
}

func WrapConnectionError(err error, url string, detector PlatformDetector) error {
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		platformName := "serveur distant"
		if detector != nil {
			platformName = detector.DetectPlatform(url)
		}
		return fmt.Errorf("impossible de se connecter à %s. Veuillez vous assurer que le service est ouvert et accessible", platformName)
	}
	return err
}
