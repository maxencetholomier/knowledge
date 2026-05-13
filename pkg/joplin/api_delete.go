package joplin

import (
	"kl/pkg/httpclient"
	"time"
)

var platformDetector = &JoplinPlatformDetector{}

func httpDelete(url string) error {
	return httpclient.Delete(url, platformDetector)
}

func deleteFromJoplin(endpoint string, id string, queryParams string) error {
	time.Sleep(200)
	url, err := buildJoplinURL(endpoint+"/"+id, queryParams)
	if err != nil {
		return err
	}
	return httpDelete(url)
}

func DeleteNoteFromJoplin(id string) error {
	return deleteFromJoplin("notes", id, "&permanent=1")
}
