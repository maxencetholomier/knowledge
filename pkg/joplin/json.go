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

