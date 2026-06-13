Read every file in `cmd/` and compare it against the documentation.
Then edit the documentation to match the code exactly. Follow these steps:

## Step 1 — Extract the CLI structure from code

For each `.go` file in `cmd/`:
- Extract the command `Use`, `Short`, `Long`, `Example` strings
- Extract every flag defined in `init()`: name, short alias, default value, description
- Note any subcommands (e.g. `anki export`, `joplin merge`)

## Step 2 — Read the existing documentation

Read these files in full:
- `docs/usage.md`
- `docs/comprehensive_exemple.md`
- `README.md`
- Any other `docs/*.md` file that documents commands

## Step 3 — Identify every discrepancy

Build a list of all gaps between code and docs:
- **Missing command**: in `cmd/` but not in `docs/usage.md` → must add a section
- **Removed command**: documented but no longer exists in `cmd/` → must remove the section
- **Missing flag**: flag exists in code but not documented → must add to the command section
- **Removed flag**: flag documented but no longer exists in code → must remove from docs
- **Inaccurate description**: `Short`/`Long`/`Example` in `cmd/*.go` does not match actual behavior (cross-reference `pkg/` if needed) → update in the Go file

## Step 4 — Edit the files

Apply all edits:

**`docs/usage.md`:**
- Update the table of contents to reflect added/removed commands
- For each missing command: add a new section following the style of existing sections (heading, short description, usage line, flags as a bullet list or table, example)
- For each removed command: delete its section
- For each missing/removed flag: update the relevant command section

**`cmd/*.go`:**
- Update `Short`, `Long`, or `Example` strings only if they are factually inaccurate
- Do not change code logic, only documentation strings

**`docs/comprehensive_exemple.md`:**
- Update the walkthrough only if a command it uses changed its name, flags, or behavior
- Keep the narrative style; do not turn it into reference documentation

**`README.md`:**
- Update the feature list only if a feature was added or removed

## Style rules

- Match the exact markdown style already used in `docs/usage.md` (headings level, flag formatting, example formatting)
- Keep descriptions concise — one sentence per flag
- Do not add commentary or notes about what you changed; just make the edits
