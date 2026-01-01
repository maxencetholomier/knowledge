package anki

import (
	"fmt"

	genanki "github.com/npcnixel/genanki-go"
)

type AnkiPackage struct {
	decks      map[string]*genanki.Deck
	model      *genanki.Model
	mediaFiles map[string][]byte
}

func CreatePackage() (*AnkiPackage, error) {
	model := CreateBasicModel()

	return &AnkiPackage{
		decks:      make(map[string]*genanki.Deck),
		model:      model,
		mediaFiles: make(map[string][]byte),
	}, nil
}

func (ap *AnkiPackage) CreateDeck(deckName string) error {
	if _, exists := ap.decks[deckName]; exists {
		return fmt.Errorf("deck '%s' already exists", deckName)
	}

	deck := genanki.StandardDeck(deckName, "Exported from Knowledge (kl)")
	ap.decks[deckName] = deck

	return nil
}

func (ap *AnkiPackage) AddNote(deckName string, note *genanki.Note) error {
	deck, exists := ap.decks[deckName]
	if !exists {
		return fmt.Errorf("deck '%s' does not exist", deckName)
	}

	deck.AddNote(note)
	return nil
}

func (ap *AnkiPackage) AddMedia(filename string, data []byte) {
	ap.mediaFiles[filename] = data
}

func (ap *AnkiPackage) WriteToFile(outputPath string) error {
	deckSlice := make([]*genanki.Deck, 0, len(ap.decks))
	for _, deck := range ap.decks {
		deckSlice = append(deckSlice, deck)
	}

	pkg := genanki.NewPackage(deckSlice).AddModel(ap.model)

	for filename, data := range ap.mediaFiles {
		pkg.AddMedia(filename, data)
	}

	err := pkg.WriteToFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to write Anki package: %w", err)
	}
	return nil
}

func (ap *AnkiPackage) GetDeckNoteCount(deckName string) int {
	deck, exists := ap.decks[deckName]
	if !exists {
		return 0
	}
	return len(deck.Notes)
}
