# Usage Guide

This guide details the main `kl` features.

Before starting, make sure you've configured your [environment](./environnement_configuration_file.md).

If you just want a quick example, check this [comprehensive example](./comprehensive_exemple.md).

If the information you're searching for is not present here, try calling the help:

```bash
kl [command] --help
```

## Table of Content

- [Core](#core)
  - [`new`](#new)
  - [`list`](#list)
  - [`find`](#find)
  - [`grep`](#grep)
  - [`edit`](#edit)
  - [`delete`](#delete)
  - [`clean`](#clean)
  - [`translate`](#translate)
  - [Useful flags](#useful-flags)
- [Media](#media)
  - [`image`](#image)
  - [`video`](#video)
  - [`schema`](#schema)
- [Local Export](#local-export)
  - [`export`](#export)
  - [`man`](#man)
- [Spaced Repetition](#spaced-repetition)
  - [`anki export`](#anki-export)
  - [`anki diff`](#anki-diff)
- [Cloud Synchronization](#cloud-synchronization)
  - [`joplin list`](#joplin-list)
  - [`joplin diff`](#joplin-diff)
  - [`joplin import`](#joplin-import)
  - [`joplin export`](#joplin-export)
  - [`joplin merge`](#joplin-merge)
  - [`joplin clean`](#joplin-clean)

## Commands

### Core

#### `new`

Create a new note with a timestamp-based filename.

```bash
kl new
```

This command opens your default editor with a new markdown file named using the current timestamp (YYYYMMDDHHMMSS.md).
The file is created in `$K_DIR` with a clean header (no timestamp in the title).

#### `list`

Display all note names and associated timestamps.

```bash
kl list
```

#### `find`

Find notes by searching in their titles (headers).

```bash
kl find [search_term]
kl find [word1] [word2] [word3]
```

Searches through note titles and outputs matching results with line numbers. Results are stored in cache at `/tmp/knowledge-search-cache`.

#### `grep`

Search for text patterns within note content using ripgrep.

```bash
kl grep [pattern]
kl grep [word1] [word2] [word3]
```

Searches through note body content and outputs matching results with line numbers. Results are stored in cache at `/tmp/knowledge-search-cache`.

#### `edit`

Edit a note by line number from the last search or find result cache, or directly by timestamp.

```bash
kl edit [line_number|timestamp]
```

#### `delete`

Delete a note by line number from the last search or find result cache, or directly by timestamp.

```bash
kl delete [line_number|timestamp]
```

Examples:

- `kl delete 1` - Delete the first result from the last search/find
- `kl delete 20240315143022` - Delete note with timestamp 20240315143022.md directly

#### `clean`

Remove empty notes (no content beyond the title) and image files that are not referenced by any notes.

```bash
kl clean
```

Lists the empty notes and unlinked images found, then asks for confirmation before deleting them.

#### `translate`

Convert timestamp(s) to the format `timestamp # Title`.

```bash
kl translate [timestamp...]
```

Accepts timestamps from arguments or stdin (one per line), with or without the `.md` extension.

- `--anki` or `-a`: Output format with `.md` extension (for Anki export)

#### Useful flags

Find and grep support :

- `--case-insensitive` or `-i`: Enable case-insensitive search (for find and grep)
- `--matching-strategy` or `-m`: Set the matching strategy for searches
- `--fzf` or `-f` : Activate interactive selection with fzf
- `--no-fzf`: Disable interactive selection with fzf (use standard output)
- `--dir`: Specify a different directory to search in (overrides K_DIR)

When searching with multiple words, you can control how the search terms are combined using the matching strategy:

- **OR** (default): Find notes containing ANY of the search terms

  ```bash
  kl find project report -m OR
  ```

- **AND**: Find notes containing ALL of the search terms
  ```bash
  kl find project report -m AND
  ```


### Media

#### `image`

Create a new note with screenshot using Flameshot.

```bash
kl image
```

This command takes a screenshot using the configured screenshot tool and creates a new note with the image embedded.

- `--new-terminal`: Open the screenshot tool in a new terminal

#### `video`

Create a new note with video recording using screenrecord. For detailed setup instructions, see the [video recording guide](./video_recording.md).

```bash
kl video
```

Creates a new note with video recording capability using the configured screen recording tool.

- `--new-terminal`: Open the screen recording tool in a new terminal

#### `schema`

Create a new note with diagram using the configured diagram tool.

```bash
kl schema
```

Opens the diagram tool configured in your `~/.config/kl/config.json` (by default Inkscape) for creating diagrams and embeds the result in a new note.

- `--new-terminal` or `-t`: Open the diagram tool in a new terminal

### Local Export

#### `export`

Export all notes to classical markdown format with readable filenames based on note headers.

```bash
kl export
```

Exports notes to the configured export directory, converting timestamp-based filenames (e.g., `20240315143022.md`) to regular markdown files with readable names based on note headers.

#### `man`

Generate man pages for the kl CLI tool and all its subcommands.

```bash
kl man [OUTPUT_DIR]
```

Creates manual pages that can be installed in your system's man page directories.
If no output directory is provided, man pages are generated in `/tmp/kl-man/`.

- `--quiet` or `-q`: Suppress output messages

**Note**: Man pages are automatically generated and installed during the installation process via `install.sh`.

### Spaced Repetition

#### `anki export`

Export selected notes to Anki package (.apkg) format for spaced repetition learning.

```bash
kl anki export
```

This command converts your knowledge notes into Anki flashcards and packages them for import into Anki. Notes are organized into separate decks based on special deck definition files in your zettelkasten directory.

**Quick Setup:**

1. Create a deck definition file in your `$K_DIR`:
   ```bash
   echo "20240315100000.md" > anki_export_vocabulary
   echo "20240316120000.md" >> anki_export_vocabulary
   ```

2. Export to Anki format:
   ```bash
   kl anki export
   ```

3. Import the generated `.apkg` file into Anki

**Features:**
- Converts note titles to flashcard fronts
- Processes markdown content to HTML with syntax highlighting
- Embeds images and media files
- Converts note links to styled text
- Custom CSS styling for optimal readability
- Organizes cards into topic-based decks

#### `anki diff`

Show differences between local notes and the decks defined by your `anki_export_*` files.

```bash
kl anki diff
```

By default shows both local-only notes and Anki-only notes.

- `--local` or `-l`: Show only notes not in Anki
- `--anki` or `-a`: Show only Anki notes not found locally


### Cloud Synchronization

The `joplin` command provides integration with the Joplin note-taking application:

#### `joplin list`

List all notes in Joplin with their timestamps and titles.

```bash
kl joplin list
```

This command retrieves all notes from Joplin via the API and displays them in the same format as `kl list`, making it easy to compare local and Joplin notes.

Output format: `TIMESTAMP - TITLE`

#### `joplin diff`

Compare local notes with Joplin notes by timestamp to identify synchronization issues.

```bash
kl joplin diff
```

This command compares the output of `kl list` and `kl joplin list` to identify which notes are missing in either location. The comparison is based solely on timestamps, not content or titles.

Shows:
- **Only in local** - Notes that exist locally but not in Joplin
- **Only in Joplin** - Notes that exist in Joplin but not locally
- **Summary** - Count of notes in each category

Flags:
- `--debug`: Show detailed content comparison for debugging

#### `joplin import`

Import notes from Joplin application into the knowledge base.

```bash
kl joplin import
```

- `--notebook` or `-n`: Specify the notebook to import notes from

#### `joplin export`

Export notes from the knowledge base to Joplin application.

```bash
kl joplin export
```

- `--notebook` or `-n`: Specify the notebook to export notes to

#### `joplin merge`

Synchronize notes between the knowledge base and Joplin by merging changes bidirectionally based on modification timestamps.

**Important**: The merge process uses modification timestamps to determine which version is newer. If you have modified the same note both locally and in the cloud, the newest version will be preserved and the older version will be overwritten. There is no conflict resolution - data loss may occur if both versions contain important changes.

```bash
kl joplin merge
```

- `--notebook` or `-n`: Specify the notebook to move local notes to when pushing to Joplin
- `--diff`: Show diff of changes for each file
- `--force-local`: Push all notes from local to Joplin, ignoring timestamps
- `--force-joplin`: Pull all notes from Joplin to local, ignoring timestamps

#### `joplin clean`

Remove notes from Joplin that do not have corresponding local files in the knowledge base.

```bash
kl joplin clean
```
