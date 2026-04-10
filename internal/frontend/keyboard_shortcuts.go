package frontend

type keys struct {
	Bucket       string
	Pen          string
	Rubber       string
	Size8        string
	Size16       string
	Size24       string
	Size32       string
	Undo         string
	UndoModifier string
}

var lobbyKeyboardShortcuts = keys{
	Bucket:       "w",
	Pen:          "q",
	Rubber:       "e",
	Size8:        "1",
	Size16:       "2",
	Size24:       "3",
	Size32:       "4",
	Undo:         "z",
	UndoModifier: "ctrl",
}
