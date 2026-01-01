package anki

import (
	genanki "github.com/npcnixel/genanki-go"
)

const BasicModelID = 1607392319

func CreateBasicModel() *genanki.Model {
	model := genanki.NewModel(BasicModelID, "Knowledge Basic")

	model.Fields = []genanki.Field{
		{Name: "Front", Ord: 0, Font: "Arial", Size: 20, Color: "#000000", Align: "left"},
		{Name: "Back", Ord: 1, Font: "Arial", Size: 20, Color: "#000000", Align: "left"},
	}

	model.Templates = []genanki.Template{
		{
			Name: "Card 1",
			Ord:  0,
			Qfmt: "{{Front}}",
			Afmt: "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}",
		},
	}

	model.CSS = GetBasicCSS()

	return model
}

func GetBasicCSS() string {
	return `.card {
	font-family: Arial, sans-serif;
	font-size: 20px;
	text-align: left;
	color: #000000;
	background-color: #ffffff;
	line-height: 1.6;
	padding: 20px;
}

.question {
	font-weight: bold;
	margin-bottom: 15px;
}

hr {
	border: none;
	border-top: 1px solid #cccccc;
	margin: 20px 0;
}

code {
	font-family: 'Monaco', 'Menlo', 'Consolas', 'Courier New', monospace;
	background-color: #f5f5f5;
	padding: 2px 4px;
	border-radius: 3px;
	font-size: 0.9em;
}

pre {
	background-color: #f6f8fa;
	border: 1px solid #d0d7de;
	border-radius: 6px;
	padding: 16px;
	overflow-x: auto;
	line-height: 1.5;
}

pre code {
	background-color: transparent;
	padding: 0;
	border-radius: 0;
	font-size: 14px;
}

img {
	max-width: 100%;
	height: auto;
	display: block;
	margin: 10px 0;
}

.note-link {
	color: #0066cc;
	font-weight: 500;
}

.note-link-dead {
	color: #999999;
	text-decoration: line-through;
}
`
}
