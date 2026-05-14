package joplin

import (
	"bytes"
	"kl/pkg/config"
	"kl/pkg/httpclient"
	"strings"
)

type JoplinPlatformDetector struct{}

var platformDetector = &JoplinPlatformDetector{}

func (d *JoplinPlatformDetector) DetectPlatform(url string) string {
	if strings.Contains(url, "localhost:41184") {
		return "Joplin"
	}
	return "serveur distant"
}

func buildJoplinURL(endpoint string, queryParams string) (string, error) {
	token, err := config.GetJoplinToken()
	if err != nil {
		return "", err
	}
	return "http://localhost:41184/" + endpoint + "?token=" + token + queryParams, nil
}

func httpGet(url string) ([]byte, error) {
	return httpclient.Get(url, platformDetector)
}

func httpSend(method string, url string, b bytes.Buffer, contentType string, context string) error {
	_, err := httpclient.Send(method, url, &b, contentType, context, platformDetector)
	return err
}

func httpDelete(url string) error {
	return httpclient.Delete(url, platformDetector)
}
