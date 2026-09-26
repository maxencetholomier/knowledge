package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type DeckSpec struct {
	Name       string   `json:"name"`
	Timestamps []string `json:"timestamps"`
}

type OrphanNote struct {
	ID        int64  `json:"id"`
	Deck      string `json:"deck"`
	Timestamp string `json:"timestamp"`
	Front     string `json:"front"`
}

const findOrphansScript = `
import json
import re
import sys
from anki.collection import Collection

TIMESTAMP = re.compile(r"\b\d{14}\b")
TAGS = re.compile(r"<[^>]+>")

def note_timestamp(tags):
    for tag in tags:
        match = TIMESTAMP.search(tag)
        if match:
            return match.group(0)
    return None

specs = json.load(sys.stdin)

try:
    col = Collection(sys.argv[1])
except Exception as e:
    print(f"  Could not open the Anki collection (is Anki running?): {e}", file=sys.stderr)
    sys.exit(1)

orphans = []
try:
    for spec in specs:
        deck_id = col.decks.id_for_name(spec["name"])
        if deck_id is None:
            continue
        expected = set(spec["timestamps"] or [])
        for cid in col.decks.cids(deck_id, children=False):
            note = col.get_note(col.get_card(cid).nid)
            timestamp = note_timestamp(note.tags)
            if timestamp is None or timestamp in expected:
                continue
            if any((c.odid or c.did) != deck_id for c in note.cards()):
                continue
            front = TAGS.sub(" ", note.fields[0]).strip()
            orphans.append({
                "id": note.id,
                "deck": spec["name"],
                "timestamp": timestamp,
                "front": front[:80],
            })
finally:
    col.close()

json.dump(orphans, sys.stdout)
`

const removeNotesScript = `
import json
import os
import sys
from anki.collection import Collection

note_ids = json.load(sys.stdin)

try:
    col = Collection(sys.argv[1])
except Exception as e:
    print(f"  Could not open the Anki collection (is Anki running?): {e}", file=sys.stderr)
    sys.exit(1)

try:
    backup_folder = os.path.join(os.path.dirname(sys.argv[1]), "backups")
    os.makedirs(backup_folder, exist_ok=True)
    col.create_backup(backup_folder=backup_folder, force=True, wait_for_completion=True)
    col.remove_notes(note_ids)
finally:
    col.close()
`

func FindOrphanNotes(collectionPath string, specs []DeckSpec) ([]OrphanNote, error) {
	input, err := json.Marshal(specs)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := runAnkiScript(findOrphansScript, string(input), []string{collectionPath}, &out); err != nil {
		return nil, err
	}

	var orphans []OrphanNote
	if err := json.Unmarshal(out.Bytes(), &orphans); err != nil {
		return nil, fmt.Errorf("failed to read orphan notes: %w", err)
	}

	return orphans, nil
}

func RemoveNotes(collectionPath string, noteIDs []int64) error {
	input, err := json.Marshal(noteIDs)
	if err != nil {
		return err
	}

	return runAnkiScript(removeNotesScript, string(input), []string{collectionPath}, os.Stdout)
}
