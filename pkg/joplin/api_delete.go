package joplin

import (
	"kl/pkg/config"
	"net/http"
	"time"
)

func DeleteResourceFromJoplin(id string) error {
	time.Sleep(200)
	token, err := config.GetJoplinToken()
	if err != nil {
		return err
	}
	url := "http://localhost:41184/resources/" + id + "?token=" + token

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

func DeleteNoteFromJoplin(id string) error {
	time.Sleep(200)
	token, err := config.GetJoplinToken()
	if err != nil {
		return err
	}
	url := "http://localhost:41184/notes/" + id + "?token=" + token + "&permanent=1"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
