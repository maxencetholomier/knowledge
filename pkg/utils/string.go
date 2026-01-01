package utils

import (
	"regexp"
	"strings"
	"time"
)


func ContainingTimestamp(line string) (bool, error) {
	pattern := `# [0-9]{14}\s*:`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	if re.MatchString(line) {
		return true, nil
	} else {
		return false, nil
	}
}

func AddTimestamp(input, timeStamp string, title string) (string, error) {
	lines := strings.Split(input, "\n")
	result := ""
	headerPresent, _ := containingHeader(lines[0])

	if len(lines) < 1 {
		return input, nil
	}

	if headerPresent {
		lines[0] = strings.Replace(lines[0], "^# ", "# "+timeStamp+" : ", 1)
		result = strings.Join(lines, "\n")
	} else {
		header := "# " + timeStamp + ": " + title + "\n"
		lines = append([]string{header}, lines...)
		result = strings.Join(lines, "\n")
	}

	return result, nil
}

func RemoveTimestampHeader(line string) (string, error) {
	pattern := `^#\s*[0-9]{14}\s*:`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	result := re.ReplaceAllString(line, "")
	result = strings.Trim(result, " ")
	return result, nil
}

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

func GetTimestamp(line string) string {
	re := regexp.MustCompile(`#\s([0-9]{14})\s*:`)
	matches := re.FindStringSubmatch(line)

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func CreateTimestamp() string {
	now := time.Now()
	time.Sleep(1 * time.Second)
	return now.Format("20060102150405")
}


func containingHeader(line string) (bool, error) {
	pattern := `^#`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	if re.MatchString(line) {
		return true, nil
	} else {
		return false, nil
	}
}
