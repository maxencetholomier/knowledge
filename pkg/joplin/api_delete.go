package joplin

import (
	"kl/pkg/config"
	"kl/pkg/httpclient"
	"time"
)

var platformDetector = &JoplinPlatformDetector{}

func httpDelete(url string) error {
	return httpclient.Delete(url, platformDetector)
}

func deleteFromJoplin(endpoint string, id string, queryParams string) error {
	time.Sleep(200)
	token, err := config.GetJoplinToken()
	if err != nil {
		return err
	}
	url := "http://localhost:41184/" + endpoint + "/" + id + "?token=" + token + queryParams

	return httpDelete(url)
}

func DeleteResourceFromJoplin(id string) error {
	return deleteFromJoplin("resources", id, "")
}

func DeleteNoteFromJoplin(id string) error {
	return deleteFromJoplin("notes", id, "&permanent=1")
}
