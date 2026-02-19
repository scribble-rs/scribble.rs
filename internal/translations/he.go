package translations

func initHebrewTranslation() {
	translation := createTranslation()
	translation.IsRtl = true

	translation.put("requires-js", "אתר זה דורש JavaScript על מנת לעבוד בצורה תקינה")

	translation.put("start-the-game", "התכונן!")
	translation.put("force-start", "התחל בלי לחכות לכל השחקנים")
	translation.put("force-restart", "התחל מחדש")
	translation.put("game-not-started-title", "המשחק לא התחיל עדיין")
	translation.put("waiting-for-host-to-start", "המתן להתחלת המשחק על-ידי המארח")
	translation.put("click-to-homepage", "חזור לדף הבית")

	translation.put("now-spectating-title", "הינך במצב צפייה")
	translation.put("now-spectating-text", "ניתן לעזוב את מצב הצפייה על-ידי לחיצה על כפתור העין למעלה")
	translation.put("now-participating-title", "הינך משתתף במשחק")
	translation.put("now-participating-text", "ניתן לעבור למצב צפייה על-ידי לחיצה על כפתור העין למעלה")

	translation.put("spectation-requested-title", "ביקשת לעבור למצב צפייה")
	translation.put("spectation-requested-text", "תעבור למצב צפייה לאחר תור זה")
	translation.put("participation-requested-title", "ביקשת להשתתף במשחק")
	translation.put("participation-requested-text", "תצורף למשחק לאחר תור זה")

	translation.put("spectation-request-cancelled-title", "בקשת מצב צפייה בוטלה")
	translation.put("spectation-request-cancelled-text", "בקשתך לעבור למצב צפייה בוטלה, תמשיך להיות משתתף במשחק")
	translation.put("participation-request-cancelled-title", "בקשת השתתפות במשחק בוטלה")
	translation.put("participation-request-cancelled-text", "בקשתך להשתתף במשחק בוטלה, תמשיך להיות צופה במשחק")

	translation.put("round", "סיבוב")
	translation.put("toggle-soundeffects", "הפעלת/כיבוי סאונד")
	translation.put("toggle-pen-pressure", "הפעלת/כיבוי שליטה בלחץ העט")
	translation.put("change-your-name", "כינוי")
	translation.put("randomize", "אקראי")
	translation.put("apply", "החל")
	translation.put("save", "שמור")
	translation.put("toggle-fullscreen", "מסך מלא")
	translation.put("toggle-spectate", "הפעלת/כיבוי מצב צפייה")
	translation.put("show-help", "הצג עזרה")
	translation.put("votekick-a-player", "הצבע להרחקת שחקן")

	translation.put("last-turn", "(תור אחרון: %s)")

	translation.put("drawer-kicked", "השחקן שהורחק צייר, לכן אף שחקן לא ייקבל ניקוד בסיבוב זה")
	translation.put("self-kicked", "הורחקת מהמשחק")
	translation.put("kick-vote", "שחקנים הצביעו להרחיק את %s (%s/%s)")
	translation.put("player-kicked", "שחקן הורחק מהמשחק")
	translation.put("owner-change", "הוא המנהל החדש של הלובי %s")

	translation.put("change-lobby-settings-tooltip", "שינוי הגדרות לובי")
	translation.put("change-lobby-settings-title", "הגדרות לובי")
	translation.put("lobby-settings-changed", "הגדרות לובי שונו")
	translation.put("advanced-settings", "הגדרות מתקדמות")
	translation.put("chill", "רגוע")
	translation.put("competitive", "תחרותי")
	translation.put("chill-alt", "למרות שמהירות מתוגמלת בניקוד גבוה יותר, זה לא נורא אם אתם קצת יותר איטיים.\nהציון הבסיסי גבוה יחסית, אז התמקדו בליהנות!")
	translation.put("competitive-alt", "ככל שתהיו מהירים יותר, כך תקבלו יותר נקודות.\nהציון הבסיסי נמוך")
	translation.put("score-calculation", "שיטת ניקוד")
	translation.put("word-language", "שפה")
	translation.put("drawing-time-setting", "זמן ציור")
	translation.put("rounds-setting", "סיבובים")
	translation.put("max-players-setting", "מספר שחקנים מקסימלי")
	translation.put("public-lobby-setting", "לובי ציבורי")
	translation.put("custom-words", "מילים נוספות")
	translation.put("custom-words-info", "מילים נוספות מופרדות בפסיק")
	translation.put("custom-words-placeholder", "מילים, מופרדות, בפסיק, כאן")
	translation.put("custom-words-per-turn-setting", "מילים נוספות בכל תור")
	translation.put("players-per-ip-limit-setting", "הגבלת שחקנים לכתובת IP")
	translation.put("words-per-turn-setting", "מילים בכל תור")
	translation.put("save-settings", "שמור הגדרות")
	translation.put("input-contains-invalid-data", "הזנת תוכן לא חוקי")
	translation.put("please-fix-invalid-input", "תקן את התוכן ונסה שנית")
	translation.put("create-lobby", "יצירת לובי")
	translation.put("create-public-lobby", "לובי ציבורי")
	translation.put("create-private-lobby", "לובי פרטי")
	translation.put("no-lobbies-yet", "אין לובים")
	translation.put("lobby-full", "הלובי מלא")
	translation.put("lobby-ip-limit-excceeded", "עברת את מקבלת החיבורים עבור כתובת IP")
	translation.put("lobby-open-tab-exists", "נראה שכבר פתחת את לובי זה בלשונית אחרת")
	translation.put("lobby-doesnt-exist", "הלובי המבוקש אינו קיים")

	translation.put("refresh", "רענון")
	translation.put("join-lobby", "הצטרף ללובי")

	translation.put("message-input-placeholder", "כתוב כאן את הניחוש שלך")

	translation.put("word-choice-warning", "המילה אם לא תבחר בזמן")
	translation.put("choose-a-word", "בחר מילה")
	translation.put("waiting-for-word-selection", "המתן לבחירת מילה")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "בוחר מילה")

	translation.put("close-guess", "הניחוש '%s' קרוב לתשובה!")
	translation.put("correct-guess", "ניחשת נכון")
	translation.put("correct-guess-other-player", "'%s' ניחש נכון")
	translation.put("round-over", "התור נגמר, לא נבחרה מילה")
	translation.put("round-over-no-word", "התור נגמר. המילה הייתה '%s'")
	translation.put("game-over-win", "כל הכבוד, ניצחת!")
	translation.put("game-over-tie", "תיקו!")
	translation.put("game-over", "סיימת במקום ה %s עם %s נקודות")
	translation.put("drawer-disconnected", "התור הסתיים, השחקן שצייר התנתק")
	translation.put("guessers-disconnected", "התור הסתיים, כל השחקנים המנחשים התנתקו")

	translation.put("change-active-color", "שינוי צבע")
	translation.put("use-pencil", "עיפרון")
	translation.put("use-eraser", "מחק")
	translation.put("use-fill-bucket", "דלי")
	translation.put("change-pencil-size-to", "שינוי גודל העיפרון/מחק ל %s")
	translation.put("clear-canvas", "ניקוי קאנבס")
	translation.put("undo", "ביטול שינוי אחרון (לא ניתן לבטל \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "החיבור עם השרת התנתק")
	translation.put("connection-lost-text", "מנסה להתחבר מחדש"+
		" ...\n\nוודא שהחיבור לאינטרנט תקין")
	translation.put("error-connecting", "שגיאה בהתחברות לשרת")
	translation.put("error-connecting-text",
		"Scribble.rs לא הצליח ליצור חיבור Socket.\n\nלמרות שנראה שחיבור האינטרנט שלך עובד, או שה\nהשרת או חומת האש שלך לא הוגדרו כראוי.\n\nכדי לנסות שוב, טען מחדש את הדף.")

	translation.put("message-too-long", "ההודעה ארוכה מידי")
	translation.put("server-shutting-down-title", "השרת בתהליך כיבוי")
	translation.put("server-shutting-down-text", "מתנצלים אך השרת בתהליך כיבוי. נסו שוב מאוחר יותר")

	// Help dialog
	translation.put("controls", "כלים")
	translation.put("pencil", "עיפרון")
	translation.put("eraser", "מחק")
	translation.put("fill-bucket", "דלי")
	translation.put("switch-tools-intro", "ניתן להחליך בין הכלים השונים על-ידי שימוש בקיצורים")
	translation.put("switch-pencil-sizes", "ניתן לעבור בין גדלי העיפרון/מחק השונים על-ידי שימוש בכפתורים %s עד %s")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "סגור")
	translation.put("no", "לא")
	translation.put("yes", "כן")
	translation.put("system", "מערכת")
	translation.put("confirm", "אישור")
	translation.put("ready", "מוכן")
	translation.put("join", "הצטרף")
	translation.put("ongoing", "פעיל")
	translation.put("game-over-lobby", "נגמר")

	translation.put("source-code", "קוד מקור")
	translation.put("help", "עזרה")
	translation.put("submit-feedback", "משוב")
	translation.put("stats", "סטטוס")

	translation.put("forbidden", "לא ניתן להציג דף זה")

	RegisterTranslation("he", translation)
}
