package translations

func initEnglishTranslation() Translation {
	translation := createTranslation()

	translation.put("requires-js", "This website requires JavaScript to run properly.")

	translation.put("start-the-game", "Start the game")
	translation.put("start", "Start")
	translation.put("game-not-started-title", "Game hasn't started")
	translation.put("waiting-for-host-to-start", "Please wait for your lobby host to start the game.")

	translation.put("round", "Round")
	translation.put("toggle-soundeffects", "Toggle soundeffects")
	translation.put("change-your-name", "Edit your username")
	translation.put("randomize", "Randomize")
	translation.put("apply", "Apply")
	translation.put("save", "Save")
	translation.put("toggle-fullscreen", "Toggle fullscreen")
	translation.put("show-help", "Show help")
	translation.put("votekick-a-player", "Vote to kick a player")
	translation.put("time-left", "Time")

	translation.put("last-turn", "(Last turn: %s)")

	translation.put("drawer-kicked", "Since the kicked player has been drawing, none of you will get any points this round.")
	translation.put("self-kicked", "You have been kicked")
	translation.put("kick-vote", "(%s/%s) players voted to kick %s.")
	translation.put("player-kicked", "Player has been kicked.")
	translation.put("owner-change", "%s is the new lobby owner.")

	translation.put("change-lobby-settings", "Change the lobby settings")
	translation.put("lobby-settings-changed", "Lobby settings changed")
	translation.put("advanced-settings", "Advanced settings")
	translation.put("word-language", "Word-Language")
	translation.put("drawing-time-setting", "Drawing Time")
	translation.put("rounds-setting", "Rounds")
	translation.put("max-players-setting", "Maximum Players")
	translation.put("public-lobby-setting", "Public Lobby")
	translation.put("custom-words", "Custom Words")
	translation.put("custom-words-info", "Enter your additional words, separating them by commas")
	translation.put("custom-words-chance-setting", "Custom Words Chance")
	translation.put("players-per-ip-limit-setting", "Players per IP Limit")
	translation.put("enable-votekick-setting", "Allow Votekick")
	translation.put("save-settings", "Save settings")
	translation.put("input-contains-invalid-data", "Your input contains invalid data:")
	translation.put("please-fix-invalid-input", "Correct the invalid input and try again.")
	translation.put("create-lobby", "Create Lobby")

	translation.put("players", "Players")
	translation.put("refresh", "Refresh")
	translation.put("join-lobby", "Join Lobby")

	translation.put("message-input-placeholder", "Type your guesses and messages here")

	translation.put("choose-a-word", "Choose a word")
	translation.put("waiting-for-word-selection", "Waiting for word selection")
	//This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "is choosing a word.")

	translation.put("close-guess", "'%s' is very close.")
	translation.put("correct-guess", "You have correctly guessed the word.")
	translation.put("correct-guess-other-player", "'%s' correctly guessed the word.")
	translation.put("round-over", "Turn over, no word was chosen.")
	translation.put("round-over-no-word", "Turn over, the word was '%s'.")
	translation.put("game-over-win", "Congratulations, you've won!")
	translation.put("game-over-tie", "It's a tie!")
	translation.put("game-over", "You placed %s. with %s points")

	translation.put("change-active-color", "Change your active color")
	translation.put("use-pencil", "Use pencil")
	translation.put("use-eraser", "Use eraser")
	translation.put("use-fill-bucket", "Use fill bucket (Fills the target area with the selected color)")
	translation.put("change-pencil-size-to", "Change the pencil / eraser size to %s")
	translation.put("clear-canvas", "Clear the canvas")
	translation.put("undo", "Revert the last change you made (Doesn't work after \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "Connection lost!")
	translation.put("connection-lost-text", "Attempting to reconnect"+
		" ...\n\nMake sure your internet connection works.\nIf the "+
		"problem persists, contact the webmaster.")
	translation.put("error-connecting", "Error connecting to server")
	translation.put("error-connecting-text",
		"Scribble.rs couldn't establish a socket connection.\n\nWhile your internet "+
			"connection seems to be working, either the\nserver or your firewall hasn't "+
			"been configured correctly.\n\nTo retry, reload the page.")

	translation.put("message-too-long", "Your message is too long.")

	//Help dialog
	translation.put("controls", "Controls")
	translation.put("pencil", "Pencil")
	translation.put("eraser", "Eraser")
	translation.put("fill-bucket", "Fill bucket")
	translation.put("switch-tools-intro", "You can switch between tools using shortcuts")
	translation.put("switch-pencil-sizes", "You can also switch between pencil sizes using keys %s to %s.")

	//Generic words
	//"close" as in "closing the window"
	translation.put("close", "Close")
	translation.put("no", "No")
	translation.put("yes", "Yes")
	translation.put("system", "System")

	translation.put("source-code", "Source Code")
	translation.put("help", "Help")
	translation.put("contact", "Contact")
	translation.put("submit-feedback", "Feedback")
	translation.put("stats", "Status")

	RegisterTranslation("en", translation)
	RegisterTranslation("en-de", translation)

	return translation
}
