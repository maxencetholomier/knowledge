package joplin

import (
	"encoding/json"
	"fmt"
	"kl/pkg/httpclient"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Note struct {
	ID          string
	Title       string
	Body        string
	ParentID    string
	UpdatedTime time.Time
}

func fetchAllPages(baseURL string, process func(map[string]interface{})) error {
	page := 0
	for {
		body, err := httpGet(baseURL + "&page=" + strconv.Itoa(page))
		if err != nil {
			return err
		}
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
		items, _ := data["items"].([]interface{})
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				process(itemMap)
			}
		}
		hasMore, err := jsonReadValue(data, "bool")
		if err != nil || hasMore != "true" {
			break
		}
		page++
	}
	return nil
}

type NoteQuery struct {
	Fields      []string
	OnlyDeleted bool
}

func GetNotes(q NoteQuery) ([]Note, error) {
	fieldsParam := "id," + strings.Join(q.Fields, ",")
	if q.OnlyDeleted {
		fieldsParam += ",deleted_time"
	}

	url, err := buildJoplinURL("notes", "&fields="+fieldsParam+"&limit=50")
	if err != nil {
		return nil, err
	}
	if q.OnlyDeleted {
		url += "&include_deleted=1"
	}

	var notes []Note
	err = fetchAllPages(url, func(item map[string]interface{}) {
		if q.OnlyDeleted {
			deletedTime, _ := item["deleted_time"].(float64)
			if deletedTime == 0 {
				return
			}
		}
		note := Note{}
		if id, ok := item["id"].(string); ok {
			note.ID = id
		}
		if title, ok := item["title"].(string); ok {
			note.Title = title
		}
		if b, ok := item["body"].(string); ok {
			note.Body = b
		}
		if parentID, ok := item["parent_id"].(string); ok {
			note.ParentID = parentID
		}
		if updatedTime, ok := item["updated_time"].(float64); ok {
			note.UpdatedTime = time.UnixMilli(int64(updatedTime))
		}
		notes = append(notes, note)
	})
	return notes, err
}

func GetTimestamps(idsType string) ([]string, error) {
	ids, err := getIds(idsType)
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

func getIds(idType string) ([]string, error) {
	url, err := buildJoplinURL(idType, "&limit=50")
	if err != nil {
		return nil, err
	}

	var ids []string
	err = fetchAllPages(url, func(item map[string]interface{}) {
		if id, ok := item["id"].(string); ok {
			ids = append(ids, id)
		}
	})
	return ids, err
}

func DownloadLinkedResources(note string, timestamp string, DirZet string) error {
	pattern := `\[.*?\]\(:/([a-zA-Z0-9]{1,32})\)`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matches := regex.FindAllStringSubmatch(note, -1)
	if len(matches) == 0 {
		return nil
	}
	for index, match := range matches {
		if len(match) >= 2 {
			err := downloadResource(match[1], timestamp, index, DirZet)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func downloadResource(id string, name string, index int, DirZet string) error {
	url, err := buildJoplinURL("resources/"+id+"/file/", "")
	if err != nil {
		return err
	}

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
	return httpclient.Get(url, platformDetector)
}

func getNotebookIdByName(notebookName string) (string, error) {
	if notebookName == "" {
		return "", nil
	}

	ids, err := getIds("folders")
	if err != nil {
		return "", err
	}

	for _, id := range ids {
		title, err := getNotebookField(id, "title")
		if err != nil {
			continue
		}
		if title == notebookName {
			return id, nil
		}
	}

	return "", fmt.Errorf("notebook '%s' not found", notebookName)
}

func getNotebookField(id string, field string) (string, error) {
	url, err := buildJoplinURL("folders/"+id, "&fields=id,"+field)
	if err != nil {
		return "", err
	}

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
