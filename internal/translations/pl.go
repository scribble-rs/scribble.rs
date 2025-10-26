package translations

func initPolishTranslation() Translation {
	translation := createTranslation()

	translation.put("requires-js", "Ta strona wymaga włączonej obsługi JavaScript aby działć poprawnie.")

	translation.put("start-the-game", "Gotowi!")
	translation.put("force-start", "Wymuś Start")
	translation.put("force-restart", "Wymuś Restart")
	translation.put("game-not-started-title", "Gra się nie zaczęła")
	translation.put("waiting-for-host-to-start", "Poczekaj aż gospodarz rozpocznie grę.")

	translation.put("now-spectating-title", "Jesteś teraz widzem")
	translation.put("now-spectating-text", "Możesz wyjść z trybu widza naciskając przycisk oka u góry ekranu.")
	translation.put("now-participating-title", "Jesteś teraz uczestnikiem")
	translation.put("now-participating-text", "Możesz wejść w tryb widza naciskając przycisk oka u góry ekranu.")

	translation.put("spectation-requested-title", "Zażądany tryb widza")
	translation.put("spectation-requested-text", "Staniesz się widzem po tej turze.")
	translation.put("participation-requested-title", "Zażądano uczestnictwa")
	translation.put("participation-requested-text", "Staniesz się uczestnikiem po tej turze.")

	translation.put("spectation-request-cancelled-title", "Anulowano żądanie trybu widza")
	translation.put("spectation-request-cancelled-text", "Anulowano twoje żądanie trybu widza, będziesz dalej uczestnikiem.")
	translation.put("participation-request-cancelled-title", "Anulowano żądanie uczestnictwa")
	translation.put("participation-request-cancelled-text", "Anulowano twoje żądanie uczestnictwa, będziesz dalej widzem.")

	translation.put("round", "Runda")
	translation.put("toggle-soundeffects", "Przełącz dźwięki")
	translation.put("change-your-name", "Ksywka")
	translation.put("randomize", "Losowo")
	translation.put("apply", "Zastosuj")
	translation.put("save", "Zapisz")
	translation.put("toggle-fullscreen", "Przełącz pełny ekran")
	translation.put("toggle-spectate", "Przełącz tryb widza")
	translation.put("show-help", "Pokaż pomoc")
	translation.put("votekick-a-player", "Głosuj za wykopaniem gracza")

	translation.put("last-turn", "(Ostatnia tura: %s)")

	translation.put("drawer-kicked", "Ponieważ wykopany gracz rysował, nikt z was nie dostanie punktów w tej rundzie.")
	translation.put("self-kicked", "Zostałeś wykopany")
	translation.put("kick-vote", "(%s/%s) gracze zagłosowali za wykopaniem %s.")
	translation.put("player-kicked", "Gracz został wykopany.")
	translation.put("owner-change", "%s jest nowym gospodarzem.")

	translation.put("change-lobby-settings-tooltip", "Zmień ustawienia lobby")
	translation.put("change-lobby-settings-title", "Ustawienia lobby")
	translation.put("lobby-settings-changed", "Ustawienia lobby zostały zmienione")
	translation.put("advanced-settings", "Ustawienia zaawansowane")
	translation.put("word-language", "Język")
	translation.put("drawing-time-setting", "Czas rysowania")
	translation.put("rounds-setting", "Rundy")
	translation.put("max-players-setting", "Maksymalna ilośc graczy")
	translation.put("public-lobby-setting", "Publiczne Lobby")
	translation.put("custom-words", "Własne słowa")
	translation.put("custom-words-info", "Wporowadź swoje dodatkowe słowa, rozdzielone przecinkami.")
	translation.put("custom-words-per-turn-setting", "Własne słowa na turę")
	translation.put("players-per-ip-limit-setting", "Limit graczy na adres IP")
	translation.put("save-settings", "Zapisz ustawienia")
	translation.put("input-contains-invalid-data", "Twoje wprowadzone dane są nieprawidłowe:")
	translation.put("please-fix-invalid-input", "Popraw błędy i sprobuj ponownie.")
	translation.put("create-lobby", "Stwórz Lobby")
	translation.put("create-public-lobby", "Stwórz Publiczne Lobby")
	translation.put("create-private-lobby", "Stwórz Prywatne Lobby")

	translation.put("refresh", "Odśwież")
	translation.put("join-lobby", "Wejdź do Lobby")

	translation.put("message-input-placeholder", "Tutaj wpisz swoje odpowiedzi i wiadomości")

	translation.put("choose-a-word", "Wybierz słowo")
	translation.put("waiting-for-word-selection", "Czekamy na wybranie słowa")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "wybiera słowo.")

	translation.put("close-guess", "'%s' jest bardzo blisko.")
	translation.put("correct-guess", "Poprawnie zgadłeś(-aś) słowo.")
	translation.put("correct-guess-other-player", "'%s' poprawnie zgadł(a) słowo.")
	translation.put("round-over", "Koniec tury, nie wybrano żadnego słowa.")
	translation.put("round-over-no-word", "Koniec tury, słowo to '%s'.")
	translation.put("game-over-win", "Gratulacje, wygrałeś(-aś)!")
	translation.put("game-over-tie", "Remis!")
	translation.put("game-over", "Zająłeś %s. miejsce z %s punktami")

	translation.put("change-active-color", "Zmień swój aktywny kolor")
	translation.put("use-pencil", "Użyj ołówka")
	translation.put("use-eraser", "Użyj gumki")
	translation.put("use-fill-bucket", "Użyj wiadra (wypełnia obszar zaznaczonym kolorem)")
	translation.put("change-pencil-size-to", "Zmień rozmiar ołówka / gumki na %s")
	translation.put("clear-canvas", "Wyczyść kanwę")
	translation.put("undo", "Cofnij ostatnią zmianę (Nie działa po \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "Utracono połączenie!")
	translation.put("connection-lost-text", "Próbuję pnownie połączyć"+
		" ...\n\nUpewnij się, że twoje połączenie internetowe działa.\nJeśli "+
		"problem się utrzymuje, skontaktuj się z administratorem.")
	translation.put("error-connecting", "Błąd połączenia z serwerem")
	translation.put("error-connecting-text",
		"Scribble.rs nie może połączyć się z socketem.\n\nTwoje połączenie z internetem "+
			"wydaje się działać, ale \nserver lub twoja zapora sieciowa nie "+
			"zostały poprawnie skonfigurowane.\n\nOdśwież stronę aby spróbować ponownie.")

	translation.put("message-too-long", "Twoja wiadomość jest za długa.")

	// Help dialog
	translation.put("controls", "Sterowanie")
	translation.put("pencil", "Ołówek")
	translation.put("eraser", "Gumka")
	translation.put("fill-bucket", "Wiadro")
	translation.put("switch-tools-intro", "Możesz przełączać się pomiędzy narzędziami za pomocą skrótów")
	translation.put("switch-pencil-sizes", "Możesz też zmieniać rozmiary ołówka używają klawiszy %s do %s.")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "Zamknij")
	translation.put("no", "Nie")
	translation.put("yes", "Tak")
	translation.put("system", "System")

	translation.put("source-code", "Kod Żródłowy")
	translation.put("help", "Pomoc")
	translation.put("submit-feedback", "Opinia")
	translation.put("stats", "Stan")

	RegisterTranslation("pl", translation)

	return translation
}
