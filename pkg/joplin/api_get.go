package joplin

import (
	"encoding/json"
	"fmt"
	"io"
	"kl/pkg/config"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func wrapConnectionError(err error, url string) error {
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		platformName := "serveur distant"
		if strings.Contains(url, "localhost:41184") {
			platformName = "Joplin"
		}

		return fmt.Errorf("impossible de se connecter à %s. Veuillez vous assurer que le service est ouvert et accessible", platformName)
	}
	return err
}

func GetField(id string, field string) (string, error) {
	value, _ := getField(id, field)
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key " + "title" + "is not a string")
	}
	return stringValue, nil
}

func GetLastUpdate(id string) (time.Time, error) {
	var t time.Time
	value, _ := getField(id, "updated_time")

	updatedTimeFloat, ok := value.(float64)
	if !ok {
		return t, fmt.Errorf("failed to parse updated_time as float64")
	}

	t = time.UnixMilli(int64(updatedTimeFloat))

	return t, nil
}

func GetTimestamps(idsType string) ([]string, error) {
	ids, err := GetIds(idsType)
	if err != nil {
		return nil, err
	}

	var timestamps []string

	if idsType == "resources" {
		for _, id := range ids {
			timestamp := strings.Split(DecryptFilename(id), ".")[0]
			if timestamp != "" {
				timestamps = append(timestamps, timestamp)
			}
		}
	} else {
		ids = filterIdsByExtension(ids, "aaa")

		for _, id := range ids {
			filename := DecryptFilename(id)
			if filename != "" {
				timestamp := strings.Split(filename, ".")[0]
				if len(timestamp) == 14 {
					timestamps = append(timestamps, timestamp)
				}
			}
		}
	}

	return timestamps, nil
}

func GetIds(idType string) ([]string, error) {
	ids := []string{}
	limit := "50"
	page := 0

	token, err := config.GetJoplinToken()
	if err != nil {
		return nil, err
	}
	url := "http://localhost:41184/" + idType + "?token=" + token
	url_formated := url + "&limit=" + limit + "&page=" + strconv.Itoa(page)

	body, err := httpGet(url_formated)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	ids, err = getIdsFromJson(data, ids)
	if err != nil {
		return nil, err
	}

	hasMore, err := jsonReadValue(data, "bool")
	if err != nil {
		return nil, err
	}

	for hasMore == "true" {

		page = page + 1

		url_formated = url + "&limit=" + limit + "&page=" + strconv.Itoa(page)

		body, err := httpGet(url_formated)
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		ids, err = getIdsFromJson(data, ids)
		if err != nil {
			return nil, err
		}

		hasMore, err = jsonReadValue(data, "bool")
		if err != nil {
			return nil, err
		}
	}

	if idType == "resources" {
		ids_image := []string{}

		for _, id := range ids {
			if strings.HasSuffix(id, "bbb") {
				ids_image = append(ids_image, id)
			}
		}
	}

	return ids, nil
}

func GetResourcesFromBody(input string, timestamp string, DirZet string) error {
	pattern := `\[.*?\]\(:/([a-zA-Z0-9]{1,32})\)`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matches := regex.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil
	}
	for index, match := range matches {
		if len(match) >= 2 {
			err := getResource(match[1], timestamp, index, DirZet)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getField(id string, field string) (interface{}, error) {

	token, err := config.GetJoplinToken()
	if err != nil {
		return "", err
	}
	url := "http://localhost:41184/notes/" + id + "?token=" + token + "&fields=id," + field

	bodyAddr, err := httpGet(url)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(bodyAddr, &data)
	if err != nil {
		return "", err
	}

	value, ok := data[field]
	if !ok {
		return "", fmt.Errorf("key " + field + " not found in JSON")
	}
	return value, nil
}

func getResource(id string, name string, index int, DirZet string) error {

	token, err := config.GetJoplinToken()
	if err != nil {
		return err
	}
	url := "http://localhost:41184/resources/" + id + "/file/" + "?token=" + token

	byte, err := httpGet(url)
	if err != nil {
		return err
	}

	if name == "" {
		name = DecryptFilename(id)
	}

	err = os.WriteFile(DirZet+"/"+name+"_"+strconv.Itoa(index)+".png", byte, 0644)
	if err != nil {
		return err
	}

	return nil

}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, wrapConnectionError(fmt.Errorf("HTTP GET request failed for URL %s: %w", url, err), url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Joplin API error (status: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return body, nil
}

func GetNotebookIdByName(notebookName string) (string, error) {
	if notebookName == "" {
		return "", nil
	}

	ids, err := GetIds("folders")
	if err != nil {
		return "", err
	}

	for _, id := range ids {
		title, err := GetNotebookField(id, "title")
		if err != nil {
			continue
		}
		if title == notebookName {
			return id, nil
		}
	}

	return "", fmt.Errorf("notebook '%s' not found", notebookName)
}

func GetNotebookField(id string, field string) (string, error) {
	token, err := config.GetJoplinToken()
	if err != nil {
		return "", err
	}
	url := "http://localhost:41184/folders/" + id + "?token=" + token + "&fields=id," + field

	bodyAddr, err := httpGet(url)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(bodyAddr, &data)
	if err != nil {
		return "", err
	}

	value, ok := data[field]
	if !ok {
		return "", fmt.Errorf("key " + field + " not found in JSON")
	}
	
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key " + field + " is not a string")
	}
	return stringValue, nil
}

func GetNoteParentId(noteId string) (string, error) {
	return GetField(noteId, "parent_id")
}
