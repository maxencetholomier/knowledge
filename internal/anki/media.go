package anki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type MediaFile struct {
	Filename string
	Data     []byte
	Path     string
}

func ExtractImages(content, baseDir string) ([]MediaFile, error) {
	imageRegex := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	matches := imageRegex.FindAllStringSubmatch(content, -1)

	var mediaFiles []MediaFile
	seenFiles := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		imagePath := match[1]
		fullPath := filepath.Join(baseDir, imagePath)

		if seenFiles[fullPath] {
			continue
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("image not found: %s", imagePath)
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read image %s: %w", imagePath, err)
		}

		mediaFiles = append(mediaFiles, MediaFile{
			Filename: filepath.Base(imagePath),
			Data:     data,
			Path:     fullPath,
		})

		seenFiles[fullPath] = true
	}

	return mediaFiles, nil
}
