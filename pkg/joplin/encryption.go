package joplin

import (
	"fmt"
	"kl/pkg/files"
	"strings"
)

func EncryptFilename(filename string, index int) string {
	ext := strings.ToLower(files.GetFileType(filename))
	timestamp := filename[:14]

	index_formatted := fmt.Sprintf("%04d", index)

	extMap := map[string]string{
		"md":   "aaa",
		"png":  "bbb",
		"jpeg": "ccc",
		"jpg":  "ccc",
		"svg":  "ddd",
	}

	suffix, exists := extMap[ext]
	if !exists {
		suffix = "aaa"
	}

	if ext == "svg" || ext == "jpg" || ext == "png" {
		return "00000000000" + index_formatted + timestamp + suffix
	} else {
		return "000000000000000" + timestamp + suffix
	}
}

func DecryptFilename(encrypted string) string {
	if len(encrypted) != 32 {
		return ""
	}

	suffix := encrypted[29:32]
	extMap := map[string]string{
		"aaa": "md",
		"bbb": "png",
		"ccc": "jpg",
		"ddd": "svg",
	}

	ext, exists := extMap[suffix]
	if !exists {
		return ""
	}

	if ext == "md" {
		timestamp := encrypted[15:29]
		return timestamp + "." + ext
	} else {
		timestamp := encrypted[15:29]
		index := strings.TrimLeft(encrypted[11:15], "0")
		if index == "" {
			index = "0"
		}
		return timestamp + "_" + index + "." + ext
	}
}
