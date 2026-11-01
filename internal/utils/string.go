package utils

import (
	"strings"
	"time"
)

func CreateTemplate(timeStamp string, option string) string {
	if option == "image" {
		return "# \n\n![](" + timeStamp + ".png)"
	} else if option == "schema" {
		return "# \n\n![](" + timeStamp + ".svg)"
	} else if option == "video" {
		return "# \n\n[Video](" + timeStamp + ".mp4)"
	} else {
		return "# "
	}
}

func GetFirstLine(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

var lastTimestamp time.Time

func CreateTimestamp() string {
	now := time.Now().Truncate(time.Second)
	if !now.After(lastTimestamp) {
		now = lastTimestamp.Add(time.Second)
	}
	lastTimestamp = now
	return now.Format("20060102150405")
}
