package joplin

import (
	"bytes"

	"github.com/maxencetholomier/knowledge/internal/config"
	"github.com/maxencetholomier/knowledge/internal/httpclient"
)

func buildJoplinURL(endpoint string, queryParams string) (string, error) {
	token, err := config.JoplinToken()
	if err != nil {
		return "", err
	}
	return "http://localhost:41184/" + endpoint + "?token=" + token + queryParams, nil
}

func httpGet(url string) ([]byte, error) {
	return httpclient.Get(url)
}

func httpSend(method string, url string, b bytes.Buffer, contentType string, context string) error {
	_, err := httpclient.Send(method, url, &b, contentType, context)
	return err
}

func httpDelete(url string) error {
	return httpclient.Delete(url)
}
