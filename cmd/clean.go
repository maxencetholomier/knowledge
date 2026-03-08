package cmd

import (
	"fmt"
	"kl/pkg/files"
	"kl/pkg/prompt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"c"},
	Short:   "Remove empty notes and unlinked images",
	Long:    `Remove empty notes (no content beyond title) and image files that are not referenced by any notes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		emptyNotes, err := findEmptyNotes()
		if err != nil {
			return fmt.Errorf("error finding empty notes: %w", err)
		}

		unlinkedImages, err := findUnlinkedImages()
		if err != nil {
			return fmt.Errorf("error finding unlinked images: %w", err)
		}

		totalFiles := len(emptyNotes) + len(unlinkedImages)
		if totalFiles == 0 {
			fmt.Println("No empty notes or unlinked images found.")
			return nil
		}

		displayFilesToClean(emptyNotes, unlinkedImages)

		confirmed, err := prompt.Confirm("Do you want to delete these files?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cleanup cancelled.")
			return nil
		}

		deletedCount := deleteFiles(emptyNotes, unlinkedImages)
		fmt.Printf("Successfully deleted %d files.\n", deletedCount)
		return nil
	},
}

func displayFilesToClean(emptyNotes, unlinkedImages []string) {
	fmt.Printf("Found %d empty notes and %d unlinked images:\n", len(emptyNotes), len(unlinkedImages))

	for _, note := range emptyNotes {
		fmt.Printf("  Empty note: %s\n", note)
	}

	for _, image := range unlinkedImages {
		fmt.Printf("  Unlinked image: %s\n", image)
	}
}

func deleteFiles(emptyNotes, unlinkedImages []string) int {
	deletedCount := 0

	for _, note := range emptyNotes {
		if err := os.Remove(filepath.Join(DirZet, note)); err != nil {
			fmt.Printf("Error deleting %s: %v\n", note, err)
		} else {
			deletedCount++
		}
	}

	for _, image := range unlinkedImages {
		if err := os.Remove(filepath.Join(DirZet, image)); err != nil {
			fmt.Printf("Error deleting %s: %v\n", image, err)
		} else {
			deletedCount++
		}
	}

	return deletedCount
}

func findEmptyNotes() ([]string, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	var emptyNotes []string
	for _, file := range fileList {
		if err := file.LoadContent(); err != nil {
			continue
		}

		lines := strings.Split(file.Content, "\n")
		hasContent := false

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if i == 0 && strings.HasPrefix(line, "#") {
				continue
			}

			hasContent = true
			break
		}

		if !hasContent {
			emptyNotes = append(emptyNotes, file.Name)
		}
	}

	return emptyNotes, nil
}

func findUnlinkedImages() ([]string, error) {
	imageFiles, err := findImageFiles()
	if err != nil {
		return nil, err
	}

	if len(imageFiles) == 0 {
		return []string{}, nil
	}

	linkedImages, err := findLinkedImages()
	if err != nil {
		return nil, err
	}

	var unlinkedImages []string
	for _, image := range imageFiles {
		isLinked := false
		for _, linked := range linkedImages {
			if image == linked {
				isLinked = true
				break
			}
		}
		if !isLinked {
			unlinkedImages = append(unlinkedImages, image)
		}
	}

	return unlinkedImages, nil
}

func findImageFiles() ([]string, error) {
	entries, err := os.ReadDir(DirZet)
	if err != nil {
		return nil, err
	}

	var imageFiles []string
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp"}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range imageExtensions {
			if strings.HasSuffix(strings.ToLower(name), ext) {
				imageFiles = append(imageFiles, name)
				break
			}
		}
	}

	return imageFiles, nil
}

func findLinkedImages() ([]string, error) {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return nil, err
	}

	linkedImagesMap := make(map[string]bool)

	markdownImagePattern := regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)
	htmlImagePattern := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)

	for _, file := range fileList {
		if err := file.LoadContent(); err != nil {
			continue
		}

		markdownMatches := markdownImagePattern.FindAllStringSubmatch(file.Content, -1)
		for _, match := range markdownMatches {
			if len(match) > 1 {
				imagePath := match[1]
				imageName := filepath.Base(imagePath)
				linkedImagesMap[imageName] = true
			}
		}

		htmlMatches := htmlImagePattern.FindAllStringSubmatch(file.Content, -1)
		for _, match := range htmlMatches {
			if len(match) > 1 {
				imagePath := match[1]
				imageName := filepath.Base(imagePath)
				linkedImagesMap[imageName] = true
			}
		}
	}

	var linkedImages []string
	for imageName := range linkedImagesMap {
		linkedImages = append(linkedImages, imageName)
	}

	return linkedImages, nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
