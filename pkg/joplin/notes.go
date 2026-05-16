package joplin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kl/pkg/utils"
	"mime/multipart"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Note struct {
	ID          string
	Title       string
	Body        string
	ParentID    string
	UpdatedTime time.Time
}

type LocalNote struct {
	Timestamp string
	Title     string
}

type Method string

const (
	POST Method = "POST"
	PUT  Method = "PUT"
)

type GetQuery struct {
	Fields      []string
	NotebookID  string
	OnlyDeleted bool
}

type WriteQuery struct {
	Method     Method
	FileName   string
	DirZet     string
	NotebookId string
	Index      int
}

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

func fetchAllPages(baseURL string, process func(map[string]interface{})) error {
	page := 1
	for {
		body, err := httpGet(baseURL + "&page=" + strconv.Itoa(page))
		if err != nil {
			return err
		}
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
		items, _ := data["items"].([]interface{})
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				process(itemMap)
			}
		}
		hasMore, err := jsonReadValue(data, "bool")
		if err != nil || hasMore != "true" {
			break
		}
		page++
	}
	return nil
}

func GetNotes(q GetQuery) ([]Note, error) {
	fieldsParam := "id," + strings.Join(q.Fields, ",")
	if q.OnlyDeleted {
		fieldsParam += ",deleted_time"
	}

	resource := "notes"
	if q.NotebookID != "" {
		resource = "folders/" + q.NotebookID + "/notes"
	}

	url, err := buildJoplinURL(resource, "&fields="+fieldsParam+"&limit=50")
	if err != nil {
		return nil, err
	}
	if q.OnlyDeleted {
		url += "&include_deleted=1"
	}

	var notes []Note
	err = fetchAllPages(url, func(item map[string]interface{}) {
		if q.OnlyDeleted {
			deletedTime, _ := item["deleted_time"].(float64)
			if deletedTime == 0 {
				return
			}
		}
		note := Note{}
		if id, ok := item["id"].(string); ok {
			note.ID = id
		}
		if title, ok := item["title"].(string); ok {
			note.Title = title
		}
		if b, ok := item["body"].(string); ok {
			note.Body = b
		}
		if parentID, ok := item["parent_id"].(string); ok {
			note.ParentID = parentID
		}
		if updatedTime, ok := item["updated_time"].(float64); ok {
			note.UpdatedTime = time.UnixMilli(int64(updatedTime))
		}
		notes = append(notes, note)
	})
	return notes, err
}

func GetNoteIDs(query GetQuery) ([]string, error) {
	resource := "notes"
	if query.NotebookID != "" {
		resource = "folders/" + query.NotebookID + "/notes"
	}
	return getRawIds(resource)
}

func getRawIds(idType string) ([]string, error) {
	url, err := buildJoplinURL(idType, "&limit=50")
	if err != nil {
		return nil, err
	}

	var ids []string
	err = fetchAllPages(url, func(item map[string]interface{}) {
		if id, ok := item["id"].(string); ok {
			ids = append(ids, id)
		}
	})
	return ids, err
}

func FilterLocalNotes(notes []Note) []LocalNote {
	var result []LocalNote
	for _, note := range notes {
		if !strings.HasSuffix(note.ID, "aaa") {
			continue
		}
		filename := IdToFilename(note.ID)
		if filename == "" {
			continue
		}
		timestamp := strings.Split(filename, ".")[0]
		if len(timestamp) != 14 {
			continue
		}
		title := strings.Split(note.Title, "\n")[0]
		if strings.HasPrefix(title, "#") {
			title = strings.TrimSpace(strings.TrimPrefix(title, "#"))
		}
		result = append(result, LocalNote{Timestamp: timestamp, Title: title})
	}
	return result
}

func Send(query WriteQuery) error {
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
		endpoint = "notes/" + FilenameToId(query.FileName, 0)
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

func SendResourceFromBody(body string, DirZet string) error {
	pattern := `\[.*?\]\(([0-9]{14}(?:_[0-9]+)?\.(?:jpg|png|svg))\)`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matches := regex.FindAllStringSubmatch(body, -1)
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

func DeleteNote(id string) error {
	return deleteFromJoplin("notes", id, "&permanent=1")
}

func deleteFromJoplin(endpoint string, id string, queryParams string) error {
	time.Sleep(200)
	url, err := buildJoplinURL(endpoint+"/"+id, queryParams)
	if err != nil {
		return err
	}
	return httpDelete(url)
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
		data["id"] = FilenameToId(filename, 0)
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
