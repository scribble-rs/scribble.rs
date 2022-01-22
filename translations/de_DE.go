package translations

func initGermanTranslation() {
	translation := createTranslation()

	translation.put("requires-js", "Diese Website benötigt JavaScript um korrekt zu funktionieren.")

	translation.put("start-the-game", "Starte das Spiel")
	translation.put("start", "Starten")
	translation.put("game-not-started-title", "Warte auf Spielstart")
	translation.put("waiting-for-host-to-start", "Bitte warte bis der Lobby Besitzer das Spiel startet.")

	translation.put("last-turn", "(Letzter Zug: %s)")

	translation.put("round", "Runde")
	translation.put("toggle-soundeffects", "Sound ein- / ausschalten")
	translation.put("change-your-name", "Benutzernamen editieren")
	translation.put("randomize", "Zufälliger Name")
	translation.put("apply", "Anwenden")
	translation.put("save", "Speichern")
	translation.put("toggle-fullscreen", "Vollbild aktivieren / deaktivieren")
	translation.put("show-help", "Hilfe anzeigen")
	translation.put("votekick-a-player", "Stimme dafür ab, einen Spieler rauszuwerfen")
	translation.put("time-left", "Zeit")

	translation.put("change-lobby-settings", "Lobby-Einstellungen ändern")
	translation.put("lobby-settings-changed", "Lobby-Einstelungen verändert")
	translation.put("advanced-settings", "Erweiterte Einstellungen")
	translation.put("word-language", "Sprache der Wörter")
	translation.put("drawing-time-setting", "Zeichenzeit")
	translation.put("rounds-setting", "Runden")
	translation.put("max-players-setting", "Maximale Spieler")
	translation.put("public-lobby-setting", "Öffentliche Lobby")
	translation.put("custom-words", "Extrawörter")
	translation.put("custom-words-info", "Gib hier deine Extrawörter ein und trenne einzelne Wörter mit einem Komma")
	translation.put("custom-words-chance-setting", "Chance auf Extrawort")
	translation.put("players-per-ip-limit-setting", "Maximale Spieler pro IP")
	translation.put("enable-votekick-setting", "Kick-Abstimmungen erlauben")
	translation.put("save-settings", "Einstellungen Speichern")
	translation.put("input-contains-invalid-data", "Deine Eingaben enthalten invalide Daten:")
	translation.put("please-fix-invalid-input", "Bitte korrigiere deine Eingaben und versuche es erneut.")
	translation.put("create-lobby", "Lobby erstellen")

	translation.put("players", "Spieler")
	translation.put("refresh", "Aktualisieren")
	translation.put("join-lobby", "Lobby beitreten")

	translation.put("message-input-placeholder", "Antworten und Nachrichten hier eingeben")

	translation.put("choose-a-word", "Wähle ein Wort")
	translation.put("waiting-for-word-selection", "Warte auf Wort-Auswahl")
	//This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "wählt gerade ein Wort.")

	translation.put("close-guess", "'%s' ist nah dran.")
	translation.put("correct-guess", "Du hast das Wort korrekt erraten.")
	translation.put("correct-guess-other-player", "'%s' hat das Wort korrekt erraten.")
	translation.put("round-over", "Zug vorbei, es wurde kein Wort gewählt.")
	translation.put("round-over-no-word", "Zug vorbei, das gewählte Wort war '%s'.")
	translation.put("game-over-win", "Glückwunsch, du hast gewonnen!")
	translation.put("game-over-tie", "Unentschieden!")
	translation.put("game-over", "Du bist %s. mit %s Punkten")

	translation.put("change-active-color", "Ändere die aktive Farbe")
	translation.put("use-pencil", "Stift bentuzen")
	translation.put("use-eraser", "Radiergummi benutzen")
	translation.put("use-fill-bucket", "Fülleimer benutzen (Füllt den Zielbereich mit der gewählten Farbe)")
	translation.put("change-pencil-size-to", "Ändere die Stift / Radiergummi Größe auf %s")
	translation.put("clear-canvas", "Leere die Zeichenfläche")
	translation.put("undo", "Mache deine letzte Änderung ungeschehen (Funktioniert nicht nach \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "Verbindung verloren!")
	translation.put("connection-lost-text", "Versuche Verbindung wiederherzustellen"+
		" ...\n\nStelle sicher, dass deine Internetverbindung funktioniert.\nFalls das "+
		"Problem weiterhin besteht, kontaktiere den Webmaster.")
	translation.put("error-connecting", "Fehler beim Verbindungsaufbau")
	translation.put("error-connecting-text",
		"Scribble.rs war nicht in der Lage eine Socket-Verbindung aufzubauen.\n\nZwar scheint dein "+
			"Internet zu funktionieren, aber entweder wurden der Server oder \ndeine Firewall falsch konfiguriert.\n\n"+
			"Versuche die Seite neu zu laden.")

	translation.put("message-too-long", "Deine Nachricht ist zu lang.")

	//Help dialog
	translation.put("controls", "Steuerung")
	translation.put("pencil", "Stift")
	translation.put("eraser", "Radiergummi")
	translation.put("fill-bucket", "Fülleimer")
	translation.put("switch-tools-intro", "Zwischen den Werkzeugen kannst du mit Tastaturkürzel wechseln")
	translation.put("switch-pencil-sizes", "Die Stiftgröße kannst du mit den Tasten %s bis %s verändern.")

	//Generic words
	//"close" as in "closing the window"
	translation.put("close", "Schließen")
	translation.put("no", "Nein")
	translation.put("yes", "Ja")
	translation.put("system", "System")

	//Footer
	translation.put("source-code", "Source Code")
	translation.put("help", "Hilfe")
	translation.put("contact", "Kontakt")
	translation.put("submit-feedback", "Feedback")
	translation.put("stats", "Status")

	RegisterTranslation("de", translation)
	RegisterTranslation("de-de", translation)
}
