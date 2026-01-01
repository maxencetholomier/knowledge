package joplin

import (
	"fmt"
	"strconv"
)

func jsonReadValue(data map[string]interface{}, expectedType string) (string, error) {

	value, ok := data["has_more"]
	if !ok {
		return "", fmt.Errorf("key body not found in JSON")
	}

	switch expectedType {
	case "bool":
		boolValue, ok := value.(bool)
		if !ok {
			return "", fmt.Errorf("value for key body is not a string")
		}

		return strconv.FormatBool(boolValue), nil

	default:
		stringValue, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("value for key body is not a string")
		}

		return stringValue, nil

	}

}

func getIdsFromJson(data map[string]interface{}, ids []string) ([]string, error) {
	if items, ok := data["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if id, ok := itemMap["id"].(string); ok {
					ids = append(ids, id)
				} else {
					return nil, fmt.Errorf("id field not found or not a string")
				}
			} else {
				return nil, fmt.Errorf("item is not a map")
			}
		}
	} else {
		return nil, fmt.Errorf("items key not found or is not a list")
	}

	return ids, nil
}
