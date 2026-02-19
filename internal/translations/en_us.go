package translations

func initEnglishTranslation() *Translation {
	translation := createTranslation()

	translation.put("requires-js", "This website requires JavaScript to run properly.")

	translation.put("start-the-game", "Ready up!")
	translation.put("force-start", "Force Start")
	translation.put("force-restart", "Force Restart")
	translation.put("game-not-started-title", "Game hasn't started")
	translation.put("waiting-for-host-to-start", "Please wait for your lobby host to start the game.")
	translation.put("click-to-homepage", "Click here to get back to the Homepage")

	translation.put("now-spectating-title", "You are now spectating")
	translation.put("now-spectating-text", "You can leave the spectator mode by pressing the eye button at the top.")
	translation.put("now-participating-title", "You are now participating")
	translation.put("now-participating-text", "You can enter the spectator mode by pressing the eye button at the top.")

	translation.put("spectation-requested-title", "Spectator mode requested")
	translation.put("spectation-requested-text", "You'll be a spectator after this turn.")
	translation.put("participation-requested-title", "Participation requested")
	translation.put("participation-requested-text", "You'll be participating after this turn.")

	translation.put("spectation-request-cancelled-title", "Spectator mode requested cancelled")
	translation.put("spectation-request-cancelled-text", "Your spectation request has been cancelled, you will keep participating.")
	translation.put("participation-request-cancelled-title", "Participation requested cancelled")
	translation.put("participation-request-cancelled-text", "Your partiticpation request has been cancelled, you will keep spectating.")

	translation.put("round", "Round")
	translation.put("toggle-soundeffects", "Toggle soundeffects")
	translation.put("toggle-pen-pressure", "Toggle pen pressure")
	translation.put("change-your-name", "Nickname")
	translation.put("randomize", "Randomize")
	translation.put("apply", "Apply")
	translation.put("save", "Save")
	translation.put("toggle-fullscreen", "Toggle fullscreen")
	translation.put("toggle-spectate", "Toggle spectator mode")
	translation.put("show-help", "Show help")
	translation.put("votekick-a-player", "Vote to kick a player")

	translation.put("last-turn", "(Last turn: %s)")

	translation.put("drawer-kicked", "Since the kicked player has been drawing, none of you will get any points this round.")
	translation.put("self-kicked", "You have been kicked")
	translation.put("kick-vote", "(%s/%s) players voted to kick %s.")
	translation.put("player-kicked", "Player has been kicked.")
	translation.put("owner-change", "%s is the new lobby owner.")

	translation.put("change-lobby-settings-tooltip", "Change the lobby settings")
	translation.put("change-lobby-settings-title", "Lobby settings")
	translation.put("lobby-settings-changed", "Lobby settings changed")
	translation.put("advanced-settings", "Advanced Settings")
	translation.put("chill", "Chill")
	translation.put("competitive", "Competitive")
	translation.put("chill-alt", "While being fast is rewarded, it's not too bad if you are little slower.\nThe base score is rather high, focus on having fun!")
	translation.put("competitive-alt", "The faster you are, the more points you will get.\nThe base score is a lot lower and the decline is faster.")
	translation.put("score-calculation", "Scoring")
	translation.put("word-language", "Language")
	translation.put("drawing-time-setting", "Drawing Time")
	translation.put("rounds-setting", "Rounds")
	translation.put("max-players-setting", "Maximum Players")
	translation.put("public-lobby-setting", "Public Lobby")
	translation.put("custom-words", "Custom Words")
	translation.put("custom-words-info", "Enter your additional words, separating them by commas")
	translation.put("custom-words-placeholder", "Comma, separated, word, list, here")
	translation.put("custom-words-per-turn-setting", "Custom Words Per Turn")
	translation.put("players-per-ip-limit-setting", "Players per IP Limit")
	translation.put("words-per-turn-setting", "Words Per Turn")
	translation.put("save-settings", "Save settings")
	translation.put("input-contains-invalid-data", "Your input contains invalid data:")
	translation.put("please-fix-invalid-input", "Correct the invalid input and try again.")
	translation.put("create-lobby", "Create Lobby")
	translation.put("create-public-lobby", "Create Public Lobby")
	translation.put("create-private-lobby", "Create Private Lobby")
	translation.put("no-lobbies-yet", "There are no lobbies yet.")
	translation.put("lobby-full", "Sorry, but the lobby is full.")
	translation.put("lobby-ip-limit-excceeded", "Sorry, but you have exceeded the maximum number of clients per IP.")
	translation.put("lobby-open-tab-exists", "It appears you already have an open tab for this lobby.")
	translation.put("lobby-doesnt-exist", "The requested lobby doesn't exist")

	translation.put("refresh", "Refresh")
	translation.put("join-lobby", "Join Lobby")

	translation.put("message-input-placeholder", "Type your guesses and messages here")

	translation.put("word-choice-warning", "Word if you don't choose in time")
	translation.put("choose-a-word", "Choose a word")
	translation.put("waiting-for-word-selection", "Waiting for word selection")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "is choosing a word.")

	translation.put("close-guess", "'%s' is very close.")
	translation.put("correct-guess", "You have correctly guessed the word.")
	translation.put("correct-guess-other-player", "'%s' correctly guessed the word.")
	translation.put("round-over", "Turn over, no word was chosen.")
	translation.put("round-over-no-word", "Turn over, the word was '%s'.")
	translation.put("game-over-win", "Congratulations, you've won!")
	translation.put("game-over-tie", "It's a tie!")
	translation.put("game-over", "You placed %s. with %s points")
	translation.put("drawer-disconnected", "Turn ended early, drawer disconnected.")
	translation.put("guessers-disconnected", "Turn ended early, guessers disconnected.")
	translation.put("word-hint-revealed", "A word hint was revealed!")

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
	translation.put("server-shutting-down-title", "Server shutting down")
	translation.put("server-shutting-down-text", "Sorry, but the server is about to shut down. Please come back at a later time.")

	// Help dialog
	translation.put("controls", "Controls")
	translation.put("pencil", "Pencil")
	translation.put("eraser", "Eraser")
	translation.put("fill-bucket", "Fill bucket")
	translation.put("switch-tools-intro", "You can switch between tools using shortcuts")
	translation.put("switch-pencil-sizes", "You can also switch between pencil sizes using keys %s to %s.")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "Close")
	translation.put("no", "No")
	translation.put("yes", "Yes")
	translation.put("system", "System")
	translation.put("confirm", "Okay")
	translation.put("ready", "Ready")
	translation.put("join", "Join")
	translation.put("ongoing", "Ongoing")
	translation.put("game-over-lobby", "Game Over")

	translation.put("source-code", "Source Code")
	translation.put("help", "Help")
	translation.put("submit-feedback", "Feedback")
	translation.put("stats", "Status")

	translation.put("forbidden", "Forbidden")

	RegisterTranslation("en", translation)

	return translation
}
