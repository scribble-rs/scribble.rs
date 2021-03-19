package translations

func initGermanTranslation() {
	translation := createTranslation()

	translation.Put("start-the-game", "Starte das Spiel")
	translation.Put("requires-js", "Diese Website benötigt JavaScript um korrekt zu funktionieren.")

	RegisterTranslation("de", translation)
	RegisterTranslation("de-DE", translation)
}
