package translations

func initDefaultTranslation() Translation {
	translation := createTranslation()

	translation.Put("start-the-game", "Start the game")
	translation.Put("requires-js", "This website requires JavaScript to run properly.")

	RegisterTranslation("en", translation)
	RegisterTranslation("en-us", translation)

	return translation
}
