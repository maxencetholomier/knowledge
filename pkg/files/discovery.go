package files

import (
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name    string
	Path    string
	Content string
}

type Scanner struct {
	Dir        string
	Extensions []string
}

func NewScanner(dir string) *Scanner {
	return &Scanner{
		Dir:        dir,
		Extensions: []string{},
	}
}

func (s *Scanner) WithExtensions(extensions ...string) *Scanner {
	s.Extensions = extensions
	return s
}

func (s *Scanner) ListFiles() ([]FileInfo, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if s.matchesExtensions(name) {
			files = append(files, FileInfo{
				Name: name,
				Path: filepath.Join(s.Dir, name),
			})
		}
	}

	return files, nil
}

func (f *FileInfo) LoadContent() error {
	if f.Content != "" {
		return nil
	}

	content, err := os.ReadFile(f.Path)
	if err != nil {
		return err
	}

	f.Content = string(content)
	return nil
}

func (f *FileInfo) GetTitle() (string, error) {
	if f.Content == "" {
		if err := f.LoadContent(); err != nil {
			return "", err
		}
	}

	lines := strings.Split(f.Content, "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}
	return "", nil
}

func (f *FileInfo) GetBody() (string, error) {
	if f.Content == "" {
		if err := f.LoadContent(); err != nil {
			return "", err
		}
	}

	lines := strings.Split(f.Content, "\n")
	if len(lines) <= 1 {
		return "", nil
	}
	return strings.Join(lines[1:], "\n"), nil
}

func (s *Scanner) matchesExtensions(filename string) bool {
	if len(s.Extensions) == 0 {
		return true
	}

	for _, ext := range s.Extensions {
		if strings.HasSuffix(filename, "."+ext) {
			return true
		}
	}
	return false
}

func GetTimestamps(files []FileInfo) []string {
	timestamps := make([]string, len(files))
	for i, file := range files {
		name := file.Name
		if dotIndex := strings.LastIndex(name, "."); dotIndex != -1 {
			timestamps[i] = name[:dotIndex]
		} else {
			timestamps[i] = name
		}
	}
	return timestamps
}