package joplin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kl/pkg/config"
	"kl/pkg/files"
	"kl/pkg/httpclient"
	"kl/pkg/utils"
	"mime/multipart"
	"os"
	"regexp"
	"strings"
	"time"
)

type Method string

const (
	POST Method = "POST"
	PUT  Method = "PUT"
)

type WriteQuery struct {
	Method     Method
	FileName   string
	DirZet     string
	NotebookId string
	Index      int
}

func isImageResource(fileName string) bool {
	extension := files.GetFileType(fileName)
	return utils.ItemInSlice([]string{"png", "jpg", "svg"}, extension)
}

func buildJoplinURL(endpoint string, queryParams string) (string, error) {
	token, err := config.GetJoplinToken()
	if err != nil {
		return "", err
	}
	return "http://localhost:41184/" + endpoint + "?token=" + token + queryParams, nil
}

func Send(query WriteQuery) error {
	time.Sleep(200)

	if isImageResource(query.FileName) {
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		if err := getBytes(query.FileName, &b, writer, query.DirZet, query.Index); err != nil {
			return err
		}

		url, err := buildJoplinURL("resources", "")
		if err != nil {
			return err
		}

		return httpSend("POST", url, b, writer.FormDataContentType(), fmt.Sprintf("resource %s", query.FileName))
	}

	endpoint := "notes"
	if query.Method == PUT {
		endpoint = "notes/" + FilenameToNoteID(query.FileName, 0)
	}

	url, err := buildJoplinURL(endpoint, "")
	if err != nil {
		return err
	}

	jsonData, err := noteToJSON(string(query.Method), query.FileName, query.DirZet, query.NotebookId)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(jsonData)
	return httpSend(string(query.Method), url, *b, "application/json", fmt.Sprintf("note %s", query.FileName))
}

func PostResourceFromBody(input string, DirZet string) error {
	pattern := `\[.*?\]\(([0-9]{14}(?:_[0-9]+)?\.(?:jpg|png|svg))\)`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matches := regex.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil
	}

	for index, match := range matches {
		if len(match) > 1 {
			err := Send(WriteQuery{Method: POST, FileName: match[1], DirZet: DirZet, Index: index})
			if err != nil {
				return fmt.Errorf("failed to post resource %s: %w", match[1], err)
			}
		}
	}

	return nil
}

func httpSend(method string, url string, b bytes.Buffer, contentType string, context string) error {
	_, err := httpclient.Send(method, url, &b, contentType, context, platformDetector)
	return err
}

func noteToJSON(method string, filename string, DirZet string, notebookId string) ([]byte, error) {
	file, err := os.ReadFile(DirZet + "/" + filename)
	if err != nil {
		return nil, err
	}

	content, err := replaceTimestampToIds(string(file))
	if err != nil {
		return nil, err
	}

	title := utils.GetFirstLine(content)
	title = strings.TrimPrefix(title, "#")
	title = strings.Trim(title, " ")

	lines := strings.Split(content, "\n")
	body := ""
	if len(lines) > 1 {
		body = strings.Join(lines[1:], "\n")
		body = strings.TrimLeft(body, "\n")
	}

	data := map[string]string{
		"title": title,
		"body":  body,
	}

	if method == "POST" {
		data["id"] = FilenameToNoteID(filename, 0)
	}

	if notebookId != "" {
		data["parent_id"] = notebookId
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func getBytes(fileName string, b *bytes.Buffer, writer *multipart.Writer, DirZet string, index int) error {
	id := FilenameToNoteID(fileName, index)

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
