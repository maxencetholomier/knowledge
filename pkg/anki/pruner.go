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

const findOrphansScript = pythonPreamble + `
specs = json.load(sys.stdin)
col = open_collection()
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

const removeNotesScript = pythonPreamble + `
note_ids = json.load(sys.stdin)
col = open_collection()
try:
    backup(col)
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
