package translations

func init() {
	var translation Translation = make(map[string]string)
	translation["start-the-game"] = "Start the game"
	translation["requires-js"] = "This website requires JavaScript to run properly."

	DefaultTranslation = translation
	translations["en"] = translation
	translations["en-us"] = translation
}
