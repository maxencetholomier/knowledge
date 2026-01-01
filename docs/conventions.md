# Conventions

## File Naming

The knowledge system uses a timestamp-based naming convention for files to ensure:

- uniqueness
- chronological ordering
- compatibility with all file systems (in case people put emojis or special characters in filenames)

## File Naming Format

All note names follow this format:

```
YYYYMMDDHHMMSS.md
```

Where:

- `YYYY` - 4-digit year
- `MM` - 2-digit month (01-12)
- `DD` - 2-digit day (01-31)
- `HH` - 2-digit hour (00-23)
- `MM` - 2-digit minute (00-59)
- `SS` - 2-digit second (00-59)

The `knowledge new` command automatically generates filenames using this convention, so manual naming is not required.

**Examples:**

- `20240315143022.md` - Note created on March 15, 2024 at 14:30:22
- `20240315143045.md` - Note created on March 15, 2024 at 14:30:45


## File structure

All notes are stored in a single flat directory structure. This means:

- There are no subdirectories or folders
- Every note file exists at the same hierarchical level

**Examples:**

```
knowledge/
├── 20240315143022.md
├── 20240315143045.md
├── 20240316091500.md
├── 20240316102345.md
├── 20240317084512.md
├── 20240318150233.md
└── 20240319163821.md
```
