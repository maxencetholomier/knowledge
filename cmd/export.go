package cmd

import (
	"io"
	"os"
	"strings"

	"kl/pkg/files"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export notes to classical markdown format",
	Long:  `Export all notes and associated resources to classical markdown format with readable filenames based on note headers in the export directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(DirExport); os.IsNotExist(err) {
			os.Mkdir(DirExport, 0755)
		}

		if err := exportMarkdownFiles(); err != nil {
			return err
		}

		if err := exportResources(); err != nil {
			return err
		}

		return nil
	},
}

func exportMarkdownFiles() error {
	scanner := files.NewScanner(DirZet).WithExtensions("md")
	fileList, err := scanner.ListFiles()
	if err != nil {
		return err
	}

	for _, file := range fileList {
		err := file.LoadContent()
		if err != nil {
			return err
		}

		header := strings.Split(file.Content, "\n")[0]
		header = strings.TrimPrefix(header, "#")
		header = strings.Trim(header, " ")
		header = strings.Replace(header, "/", ":", -1)

		if header == "" {
			header = strings.TrimSuffix(file.Name, ".md")
		}

		newFiles, err := os.Create(DirExport + header + ".md")
		if err != nil {
			return err
		}
		defer newFiles.Close()
		newFiles.WriteString(file.Content)
	}

	return nil
}

func exportResources() error {
	resourcesFiletypes := []string{"png", "jpg", "jpeg", "svg"}

	for _, filetype := range resourcesFiletypes {
		resourceScanner := files.NewScanner(DirZet).WithExtensions(filetype)
		resourceList, err := resourceScanner.ListFiles()
		if err != nil {
			return err
		}

		for _, file := range resourceList {
			sourceFile, err := os.Open(file.Path)
			if err != nil {
				return err
			}
			defer sourceFile.Close()

			destFile, err := os.Create(DirExport + file.Name)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, sourceFile)
			if err != nil {
				return err
			}

			err = destFile.Sync()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
