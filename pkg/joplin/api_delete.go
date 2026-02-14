package joplin

import (
	"kl/pkg/config"
	"net/http"
	"time"
)

func httpDelete(url string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return wrapConnectionError(err, url)
	}
	defer resp.Body.Close()

	return nil
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
