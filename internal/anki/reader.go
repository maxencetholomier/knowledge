package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type CollectionNote struct {
	ID        int64  `json:"id"`
	Deck      string `json:"deck"`
	Timestamp string `json:"timestamp"`
	Front     string `json:"front"`
	Back      string `json:"back"`
	Markdown  string `json:"markdown"`
	Modified  int64  `json:"mod"`
}

type Adoption struct {
	ID        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	GUID      string `json:"guid"`
}

const readNotesScript = pythonPreamble + `
deck_names = json.load(sys.stdin)
model_id = int(sys.argv[2])
col = open_collection()
notes = []
try:
    for deck_name in deck_names:
        deck_id = col.decks.id_for_name(deck_name)
        if deck_id is None:
            continue
        seen = set()
        for cid in col.decks.cids(deck_id, children=False):
            nid = col.get_card(cid).nid
            if nid in seen:
                continue
            seen.add(nid)
            note = col.get_note(nid)
            if note.note_type()["id"] != model_id:
                continue
            fields = note.fields
            notes.append({
                "id": note.id,
                "deck": deck_name,
                "timestamp": note_timestamp(note.tags) or "",
                "front": TAGS.sub(" ", fields[0]).strip(),
                "back": fields[1] if len(fields) > 1 else "",
                "markdown": fields[2] if len(fields) > 2 else "",
                "mod": note.mod,
            })
finally:
    col.close()

json.dump(notes, sys.stdout)
`

const adoptNotesScript = pythonPreamble + `
adoptions = json.load(sys.stdin)
col = open_collection()
try:
    backup(col)
    for adoption in adoptions:
        note = col.get_note(adoption["id"])
        note.tags.append(adoption["timestamp"])
        note.guid = adoption["guid"]
        col.update_note(note)
finally:
    col.close()
`

func ReadNotes(collectionPath string, deckNames []string) ([]CollectionNote, error) {
	input, err := json.Marshal(deckNames)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	args := []string{collectionPath, fmt.Sprint(BasicModelID)}
	if err := runAnkiScript(readNotesScript, string(input), args, &out); err != nil {
		return nil, err
	}

	var notes []CollectionNote
	if err := json.Unmarshal(out.Bytes(), &notes); err != nil {
		return nil, fmt.Errorf("failed to read Anki notes: %w", err)
	}

	return notes, nil
}

func AdoptNotes(collectionPath string, adoptions []Adoption) error {
	for i := range adoptions {
		adoptions[i].GUID = fmt.Sprintf("%x", GenerateDeterministicID(adoptions[i].Timestamp))
	}

	input, err := json.Marshal(adoptions)
	if err != nil {
		return err
	}

	return runAnkiScript(adoptNotesScript, string(input), []string{collectionPath}, os.Stdout)
}
