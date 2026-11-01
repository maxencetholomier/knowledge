package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	buildDir, err := os.MkdirTemp("", "kl-build")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(buildDir, "kl")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		panic(string(out))
	}
	code := m.Run()
	os.RemoveAll(buildDir)
	os.Exit(code)
}

type vault struct {
	t    *testing.T
	home string
	dir  string
}

func newVault(t *testing.T) vault {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, "zet")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	config := `{"dirExport": "` + filepath.Join(home, "export") + `/"}`
	if err := os.MkdirAll(filepath.Join(home, ".config", "kl"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "kl", "config.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	return vault{t: t, home: home, dir: dir}
}

func (v vault) run(args ...string) (string, error) {
	cmd := exec.Command(binary, append([]string{"--dir", v.dir}, args...)...)
	cmd.Env = append(os.Environ(), "HOME="+v.home, "XDG_CACHE_HOME="+v.home, "EDITOR=true", "K_DIR=")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (v vault) mustRun(args ...string) string {
	v.t.Helper()
	out, err := v.run(args...)
	if err != nil {
		v.t.Fatalf("kl %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func (v vault) write(name, content string) {
	v.t.Helper()
	if err := os.WriteFile(filepath.Join(v.dir, name), []byte(content), 0644); err != nil {
		v.t.Fatal(err)
	}
}

func (v vault) exists(name string) bool {
	_, err := os.Stat(filepath.Join(v.dir, name))
	return err == nil
}

func (v vault) notes() []string {
	v.t.Helper()
	matches, err := filepath.Glob(filepath.Join(v.dir, "*.md"))
	if err != nil {
		v.t.Fatal(err)
	}
	for i, match := range matches {
		matches[i] = filepath.Base(match)
	}
	return matches
}

func TestNewListFindGrep(t *testing.T) {
	v := newVault(t)

	v.mustRun("new")
	notes := v.notes()
	if len(notes) != 1 || len(notes[0]) != 17 {
		t.Fatalf("new should create one timestamped note, got %v", notes)
	}
	timestamp := strings.TrimSuffix(notes[0], ".md")
	v.write(notes[0], "# Hello world\n\nsome body text\n")
	v.write("20200101120000.md", "# Other\n\nnothing here\n")

	if out := v.mustRun("list"); !strings.Contains(out, timestamp+" # Hello world") {
		t.Errorf("list output: %q", out)
	}
	if out := v.mustRun("find", "Hello"); !strings.Contains(out, "Hello world") || strings.Contains(out, "Other") {
		t.Errorf("find output: %q", out)
	}
	if out := v.mustRun("grep", "body"); !strings.Contains(out, "Hello world") || strings.Contains(out, "Other") {
		t.Errorf("grep output: %q", out)
	}
}

func TestExportDelete(t *testing.T) {
	v := newVault(t)
	v.write("20200101120000.md", "# Hello world\n\ntext\n")

	v.mustRun("export")
	if _, err := os.Stat(filepath.Join(v.home, "export", "Hello world.md")); err != nil {
		t.Errorf("export should create a readable copy: %v", err)
	}

	v.mustRun("delete", "20200101120000")
	if v.exists("20200101120000.md") {
		t.Error("delete should remove the note")
	}
}

func TestClean(t *testing.T) {
	v := newVault(t)
	v.write("20200101120000.md", "# Full\n\ntext ![](20200101120000.png)\n")
	v.write("20200101120001.md", "# Empty\n\n")
	v.write("20200101120000.png", "")
	v.write("orphan.png", "")
	v.write("anki_export_test", "20200101120000.md # Full\n20190101120000.md # gone\n")

	out := v.mustRun("clean", "--yes")

	if v.exists("20200101120001.md") || v.exists("orphan.png") {
		t.Errorf("clean should delete the empty note and the orphan image:\n%s", out)
	}
	if !v.exists("20200101120000.md") || !v.exists("20200101120000.png") {
		t.Errorf("clean must keep linked files:\n%s", out)
	}
	deck, _ := os.ReadFile(filepath.Join(v.dir, "anki_export_test"))
	if string(deck) != "20200101120000.md # Full\n" {
		t.Errorf("deck file should only keep local notes, got %q", deck)
	}
}

func TestAnkiExportWithoutImport(t *testing.T) {
	v := newVault(t)
	v.write("20200101120000.md", "# Question\n\nAnswer with `code`\n")
	v.write("anki_export_test", "20200101120000.md\n")

	out := v.mustRun("anki", "export", "--no-import")
	if _, err := os.Stat(filepath.Join(v.home, "export", "anki_cards_test.apkg")); err != nil {
		t.Errorf("expected an apkg file: %v\n%s", err, out)
	}
}

func TestErrors(t *testing.T) {
	v := newVault(t)

	if out, err := v.run("joplin", "list"); err == nil || !strings.Contains(out, "joplin config file not found") {
		t.Errorf("joplin without config should fail clearly, got err=%v\n%s", err, out)
	}

	cmd := exec.Command(binary, "list")
	cmd.Env = append(os.Environ(), "HOME="+v.home, "K_DIR=")
	if out, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(out), "K_DIR") {
		t.Errorf("list without a directory should fail, got err=%v\n%s", err, out)
	}

	if out := v.mustRun("--version"); !strings.HasPrefix(out, "kl version") {
		t.Errorf("version output: %q", out)
	}
}
