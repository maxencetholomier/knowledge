package cmd

import (
	"bufio"
	"fmt"
	"kl/pkg/files"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var translateCmd = &cobra.Command{
	Use:     "translate [timestamp...]",
	Aliases: []string{"t"},
	Short:   "Convert timestamps to formatted titles",
	Long: `Convert timestamp(s) to format: timestamp # Title
Accepts timestamps from arguments or stdin (one per line).
Timestamps can be with or without .md extension.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var timestamps []string

		if len(args) == 0 {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line != "" {
						timestamps = append(timestamps, line)
					}
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("error reading stdin: %w", err)
				}
			} else {
				return fmt.Errorf("no timestamps provided. Use: kl translate <timestamp> or pipe timestamps via stdin")
			}
		} else {
			timestamps = args
		}

		return translateTimestamps(timestamps, DirZet)
	},
}

func normalizeTimestamp(ts string) string {
	ts = strings.TrimSpace(ts)
	ts = strings.TrimSuffix(ts, ".md")
	return ts
}

func isValidTimestamp(ts string) bool {
	if len(ts) != 14 {
		return false
	}
	matched, _ := regexp.MatchString(`^\d{14}$`, ts)
	return matched
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.TrimPrefix(title, "#")
	title = strings.TrimSpace(title)
	return title
}

func translateTimestamps(timestamps []string, dirZet string) error {
	for _, ts := range timestamps {
		cleanTS := normalizeTimestamp(ts)

		if !isValidTimestamp(cleanTS) {
			fmt.Fprintf(os.Stderr, "Invalid timestamp: %s\n", ts)
			continue
		}

		filePath := filepath.Join(dirZet, cleanTS+".md")
		fileInfo := files.FileInfo{Path: filePath, Name: cleanTS + ".md"}

		title, err := fileInfo.GetTitle()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", cleanTS, err)
			continue
		}

		cleanedTitle := cleanTitle(title)
		fmt.Printf("%s # %s\n", cleanTS, cleanedTitle)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(translateCmd)
}
