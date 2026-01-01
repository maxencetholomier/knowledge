package joplin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kl/pkg/config"
	"kl/pkg/files"
	"kl/pkg/utils"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func PostToJoplin(fileName string, DirZet string) error {

	time.Sleep(200)
	extension := files.GetFileType(fileName)

	if utils.ItemInSlice([]string{"png", "jpg", "svg"}, extension) {

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		err := getBytes(fileName, &b, writer, DirZet)
		if err != nil {
			return err
		}

		token, err := config.GetJoplinToken()
		if err != nil {
			return err
		}
		url := "http://localhost:41184/resources?token=" + token

		return httpSend("POST", url, b, writer.FormDataContentType(), fmt.Sprintf("resource %s", fileName))
	} else {
		token, err := config.GetJoplinToken()
		if err != nil {
			return err
		}
		url := "http://localhost:41184/notes?token=" + token

		jsonData, err := get_data("POST", fileName, DirZet, "")
		if err != nil {
			return err
		}

		b := bytes.NewBuffer(jsonData)
		return httpSend("POST", url, *b, "application/json", fmt.Sprintf("note %s", fileName))
	}

}

func PostToJoplinWithNotebook(fileName string, DirZet string, notebookId string) error {

	time.Sleep(200)
	extension := files.GetFileType(fileName)

	if utils.ItemInSlice([]string{"png", "jpg", "svg"}, extension) {

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		err := getBytes(fileName, &b, writer, DirZet)
		if err != nil {
			return err
		}

		token, err := config.GetJoplinToken()
		if err != nil {
			return err
		}
		url := "http://localhost:41184/resources?token=" + token

		return httpSend("POST", url, b, writer.FormDataContentType(), fmt.Sprintf("resource %s", fileName))
	} else {
		token, err := config.GetJoplinToken()
		if err != nil {
			return err
		}
		url := "http://localhost:41184/notes?token=" + token

		jsonData, err := get_data("POST", fileName, DirZet, notebookId)
		if err != nil {
			return err
		}

		b := bytes.NewBuffer(jsonData)
		return httpSend("POST", url, *b, "application/json", fmt.Sprintf("note %s", fileName))
	}

}

func PutNoteToJoplin(fileName string, DirZet string) error {

	id := EncryptFilename(fileName, 0)

	token, err := config.GetJoplinToken()
	if err != nil {
		return err
	}
	url := "http://localhost:41184/notes/" + id + "?token=" + token

	time.Sleep(200)

	jsonData, err := get_data("PUT", fileName, DirZet, "")
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(jsonData)
	return httpSend("PUT", url, *b, "application/json", fmt.Sprintf("note %s", fileName))
}

func PutNoteToJoplinWithNotebook(fileName string, DirZet string, notebookId string) error {

	id := EncryptFilename(fileName, 0)

	token, err := config.GetJoplinToken()
	if err != nil {
		return err
	}
	url := "http://localhost:41184/notes/" + id + "?token=" + token

	time.Sleep(200)

	jsonData, err := get_data("PUT", fileName, DirZet, notebookId)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(jsonData)
	return httpSend("PUT", url, *b, "application/json", fmt.Sprintf("note %s", fileName))
}

func PostResourceFromBody(input string, DirZet string) error {
	time.Sleep(200)
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
			filename := match[1]
			err := PostToJoplinWithIndex(filename, DirZet, index)
			if err != nil {
				return fmt.Errorf("failed to post resource %s: %w", filename, err)
			}
		}
	}

	return nil
}

func httpSend(method string, url string, b bytes.Buffer, contentType string, context string) error {
	var req *http.Request
	var err error

	req, err = http.NewRequest(method, url, &b)
	if err != nil {
		if context != "" {
			return fmt.Errorf("failed to create request for %s: %w", context, err)
		}
		return err
	}

	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if context != "" {
			return fmt.Errorf("HTTP request failed for %s: %w", context, err)
		}
		return err
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		body, _ := io.ReadAll(resp.Body)
		if context != "" {
			return fmt.Errorf("Joplin API error for %s (status: %s): %s", context, resp.Status, string(body))
		}
		return fmt.Errorf("Joplin API error (status: %s): %s", resp.Status, string(body))
	}

	return nil
}

func get_data(method string, filename string, DirZet string, notebookId string) ([]byte, error) {

	file, err := os.ReadFile(DirZet + "/" + filename)
	if err != nil {
		return nil, err
	}

	var data map[string]string

	if method == "POST" {
		content, err := ReplaceTimestampToIds(string(file))
		if err != nil {
			return nil, err
		}

		title := utils.GetFirstLine(content)
		title = strings.TrimPrefix(title, "#")
		title = strings.Trim(title, " ")

		lines := strings.Split(content, "\n")
		bodyWithoutTitle := ""
		if len(lines) > 1 {
			bodyWithoutTitle = strings.Join(lines[1:], "\n")
			bodyWithoutTitle = strings.TrimLeft(bodyWithoutTitle, "\n")
		}

		id := EncryptFilename(filename, 0)

		data = map[string]string{
			"id":    id,
			"title": title,
			"body":  bodyWithoutTitle,
		}

		if notebookId != "" {
			data["parent_id"] = notebookId
		}
	} else {

		content, err := ReplaceTimestampToIds(string(file))
		if err != nil {
			return nil, err
		}

		data = map[string]string{
			"body": content,
		}

		if notebookId != "" {
			data["parent_id"] = notebookId
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func getBytes(fileName string, b *bytes.Buffer, writer *multipart.Writer, DirZet string) error {
	id := EncryptFilename(fileName, 0)

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
	err = writer.WriteField("props", string(jsonData))
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}

func PostToJoplinWithIndex(fileName string, DirZet string, index int) error {
	time.Sleep(200)
	extension := files.GetFileType(fileName)

	if utils.ItemInSlice([]string{"png", "jpg", "svg"}, extension) {

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		err := getBytesWithIndex(fileName, &b, writer, DirZet, index)
		if err != nil {
			return err
		}

		token, err := config.GetJoplinToken()
		if err != nil {
			return err
		}
		url := "http://localhost:41184/resources?token=" + token

		return httpSend("POST", url, b, writer.FormDataContentType(), fmt.Sprintf("resource %s", fileName))
	} else {
		token, err := config.GetJoplinToken()
		if err != nil {
			return err
		}
		url := "http://localhost:41184/notes?token=" + token

		jsonData, err := get_data("POST", fileName, DirZet, "")
		if err != nil {
			return err
		}

		b := bytes.NewBuffer(jsonData)
		return httpSend("POST", url, *b, "application/json", fmt.Sprintf("note %s", fileName))
	}
}

func getBytesWithIndex(fileName string, b *bytes.Buffer, writer *multipart.Writer, DirZet string, index int) error {
	id := EncryptFilename(fileName, index)

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
	err = writer.WriteField("props", string(jsonData))
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}
