package joplin

import (
	"encoding/json"
	"fmt"
	"kl/pkg/config"
)

func GetNotebookInfo(notebookName string) (string, string, error) {
	if notebookName == "" {
		notebookName = config.GetJoplinNotebook()
	}

	var notebookId string
	if notebookName != "" {
		var err error
		notebookId, err = getNotebookIdByName(notebookName)
		if err != nil {
			return "", "", err
		}
	}

	return notebookName, notebookId, nil
}

func getNotebookIdByName(notebookName string) (string, error) {
	if notebookName == "" {
		return "", nil
	}

	ids, err := getRawIds("folders")
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
