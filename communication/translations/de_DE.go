package translations

func init() {
	var translation Translation = make(map[string]string)
	translation["start-the-game"] = "Starte das Spiel"
	translation["requires-js"] = "Diese Website benötigt JavaScript um korrekt zu funktionieren."

	DefaultTranslation = translation
	translations["de"] = translation
	translations["de-DE"] = translation
}
