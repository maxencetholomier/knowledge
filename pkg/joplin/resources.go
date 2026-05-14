package joplin

import (
	"bytes"
	"encoding/json"
	"io"
	"kl/pkg/files"
	"kl/pkg/utils"
	"mime/multipart"
	"os"
	"regexp"
	"strconv"
)

func isImageResource(fileName string) bool {
	extension := files.GetFileType(fileName)
	return utils.ItemInSlice([]string{"png", "jpg", "svg"}, extension)
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
		name = IdToFilename(id)
	}

	err = os.WriteFile(DirZet+"/"+name+"_"+strconv.Itoa(index)+".png", byte, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getBytes(fileName string, b *bytes.Buffer, writer *multipart.Writer, DirZet string, index int) error {
	id := FilenameToId(fileName, index)

	filePath := DirZet + "/" + fileName
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("data", file.Name())
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	data := map[string]string{
		"id":    id,
		"title": id,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err = writer.WriteField("props", string(jsonData)); err != nil {
		return err
	}

	return writer.Close()
}
