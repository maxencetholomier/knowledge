package anki

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const importScript = `
import os
import sys
from anki.collection import Collection, ImportAnkiPackageOptions, ImportAnkiPackageRequest

try:
    col = Collection(sys.argv[1])
except Exception as e:
    print(f"  Could not open the Anki collection (is Anki running?): {e}", file=sys.stderr)
    sys.exit(1)

try:
    backup_folder = os.path.join(os.path.dirname(sys.argv[1]), "backups")
    col.create_backup(backup_folder=backup_folder, force=True, wait_for_completion=True)
    for path in sys.argv[2:]:
        request = ImportAnkiPackageRequest(
            package_path=path,
            options=ImportAnkiPackageOptions(merge_notetypes=True),
        )
        log = col.import_anki_package(request).log
        name = os.path.basename(path)
        print(f"  {name}: {len(log.new)} new, {len(log.updated)} updated, {len(log.duplicate)} unchanged")
finally:
    col.close()
`

func IsAnkiRunning() bool {
	if exec.Command("pgrep", "-x", "anki").Run() == nil {
		return true
	}
	return exec.Command("pgrep", "-f", "aqt").Run() == nil
}

func findAnkiPython() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	venvPython := filepath.Join(home, ".local", "share", "AnkiProgramFiles", ".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err == nil {
		return venvPython, nil
	}

	if exec.Command("python3", "-c", "import anki").Run() == nil {
		return "python3", nil
	}

	return "", fmt.Errorf("anki Python library not found (looked for %s and python3 with the 'anki' package)", venvPython)
}

func FindCollection() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	base := filepath.Join(home, ".local", "share", "Anki2")
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", fmt.Errorf("failed to read Anki data directory %s: %w", base, err)
	}

	var newest string
	var newestModTime int64
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		collection := filepath.Join(base, entry.Name(), "collection.anki2")
		info, err := os.Stat(collection)
		if err != nil {
			continue
		}
		if newest == "" || info.ModTime().Unix() > newestModTime {
			newest = collection
			newestModTime = info.ModTime().Unix()
		}
	}

	if newest == "" {
		return "", fmt.Errorf("no Anki profile with a collection found in %s", base)
	}

	return newest, nil
}

func ImportPackages(collectionPath string, apkgPaths []string) error {
	python, err := findAnkiPython()
	if err != nil {
		return err
	}

	args := append([]string{"-", collectionPath}, apkgPaths...)
	cmd := exec.Command(python, args...)
	cmd.Stdin = strings.NewReader(importScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
