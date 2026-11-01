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

func CreateTimestamp() string {
	now := time.Now()
	time.Sleep(1 * time.Second)
	return now.Format("20060102150405")
}
