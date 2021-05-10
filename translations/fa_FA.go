package translations

func initPersianTranslation() Translation {
	translation := createTranslation()

	translation.put("requires-js", "این وبسایت به جاوا اسکریپت نیاز دارن")

	translation.put("start-the-game", "شروع بازی")
	translation.put("start", "شروع")
	translation.put("game-not-started-title", "بازی شروع نشده")
	translation.put("waiting-for-host-to-start", "لطفا منتظر بمانید بازی شروع شود")

	translation.put("round", "راند")
	translation.put("toggle-soundeffects", "قطع/وصل کردن صدا")
	translation.put("change-your-name", "تغییر نام کاربری")
	translation.put("randomize", "شانسی")
	translation.put("apply", "تایید")
	translation.put("save", "ذخیره")
	translation.put("votekick-a-player", "رای برای اخراج یک بازیکان")
	translation.put("time-left", "زمان باقیمانده")

	translation.put("last-turn", "(آخرین نوبت: %s)")

	translation.put("drawer-kicked", "نقاش اخراج شد، کسی امتیازی نمیگیرد")
	translation.put("self-kicked", "شما اخراج شدد")
	translation.put("kick-vote", "(%s/%s) رای برای اخراج %s داد.")
	translation.put("player-kicked", "بازیکن اخراج شد")
	translation.put("owner-change", "%s مدیر جدید")

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
	translation.put("correct-guess-other-player", "%s correctly guessed the word.")
	translation.put("round-over", "Turn over, no word was chosen.")
	translation.put("round-over-no-word", "Turn over, the word was '%s'.")
	translation.put("game-over-win", "Congratulations, you've won!")
	translation.put("game-over", "You placed %s. with %s points")

	translation.put("change-active-color", "Change your active color")
	translation.put("use-pencil", "Use pencil")
	translation.put("use-eraser", "Use eraser")
	translation.put("use-fill-bucket", "Use fill bucket (Fills the target area with the selected color)")
	translation.put("change-pencil-size-to", "Change the pencil / eraser size to %s")
	translation.put("clear-canvas", "Clear the canvas")

	translation.put("connection-lost", "Connection lost!")
	translation.put("connection-lost-text", "Attempting to reconnect"+
		" ...\n\nMake sure your internet connection works.\nIf the "+
		"problem persists, contact the webmaster.")
	translation.put("error-connecting", "Error connecting to server")
	translation.put("error-connecting-text",
		"Scribble.rs couldn't establish a socket connection.\n\nWhile your internet "+
			"connection seems to be working, either the\nserver or your firewall hasn't "+
			"been configured correctly.\n\nTo retry, reload the page.")

	//Generic words
	//As "close" in "closing the window"
	translation.put("close", "بستن")
	translation.put("no", "خیر")
	translation.put("yes", "بله")
	translation.put("system", "سیستم")

	translation.put("source-code", "سورس کد")
	translation.put("help", "راهنما")
	translation.put("contact", "تماس با ما")
	translation.put("submit-feedback", "ثبت نظر")
	translation.put("stats", "استاتوس")

	RegisterTranslation("fa", translation)

	return translation
}
