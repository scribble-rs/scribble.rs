package translations

func initPersianTranslation() *Translation {
	translation := createTranslation()
	translation.IsRtl = true

	translation.put("requires-js", "این وبسایت برای درست اجرا شدن نیاز به جاوااسکریپت داره.")

	translation.put("start-the-game", "بزن بریم!")
	translation.put("force-start", "شروع فوری")
	translation.put("force-restart", "راه‌اندازی دوباره فوری")
	translation.put("game-not-started-title", "بازی هنوز شروع نشده")
	translation.put("waiting-for-host-to-start", "لطفا منتظر میزبان لابی باشید تا بازی رو شروع کنه.")
	translation.put("click-to-homepage", "برای بازگشت به خونه اینجا رو کلیک کنید")

	translation.put("now-spectating-title", "الان شما تماشاچی هستید")
	translation.put("now-spectating-text", "برای خروج از حالت تماشاچی می‌تونید روی دکمه چشم بالای صفحه بزنید.")
	translation.put("now-participating-title", "الان شما شرکت‌کننده هستید")
	translation.put("now-participating-text", "برای ورود به حالت تماشاچی روی دکمه چشم بالای صفحه بزنید.")

	translation.put("spectation-requested-title", "حالت تماشاچی درخواست شد")
	translation.put("spectation-requested-text", "بعد از این دور بازی شما تماشاچی می‌شید.")
	translation.put("participation-requested-title", "درخواست شرکت تو بازی داده شد")
	translation.put("participation-requested-text", "بعد از این دور بازی شما هم می‌تونید بازی کنید.")

	translation.put("spectation-request-cancelled-title", "درخواست حالت تماشاچی لغو شد")
	translation.put("spectation-request-cancelled-text", "درخواست حالت تماشاچی لغو شد، می‌تونید به بازی ادامه بدید.")
	translation.put("participation-request-cancelled-title", "درخواست شرکت تو بازی لغو شد")
	translation.put("participation-request-cancelled-text", "درخواست شرکت تو بازی لغو شد، می‌تونید به تماشای بازی ادامه بدید.")

	translation.put("round", "دور")
	translation.put("toggle-soundeffects", "روشن/خاموش کردن افکت‌های صدا")
	translation.put("toggle-pen-pressure", "روشن/خاموش کردن فشار قلم")
	translation.put("change-your-name", "لقب")
	translation.put("randomize", "انتخاب تصادفی")
	translation.put("apply", "ثبت")
	translation.put("save", "ذخیره")
	translation.put("toggle-fullscreen", "روشن/خاموش کردن حالت تمام صفحه")
	translation.put("toggle-spectate", "روشن/خاموش کردن حالت تماشاچی")
	translation.put("show-help", "نمایش راهنما")
	translation.put("votekick-a-player", "رای به بیرون انداختن یه بازیکن")

	translation.put("last-turn", "(دور آخر: %s)")

	translation.put("drawer-kicked", "چون کسی که بیرون انداخته شد نقاش بود، هیچ کدوم از شما این دست امتیازی نمی‌گیرید.")
	translation.put("self-kicked", "شما بیرون انداخته شدید")
	translation.put("kick-vote", "(%s/%s) بازیکن رای به بیرون انداختن %s دادن.")
	translation.put("player-kicked", "بازیکن بیرون انداخته شد.")
	translation.put("owner-change", "%s میزبان جدید لابیه.")

	translation.put("change-lobby-settings-tooltip", "تغییر تنظیمات لابی")
	translation.put("change-lobby-settings-title", "تنظیمات لابی")
	translation.put("lobby-settings-changed", "تنظیمات لابی تغییر کرد")
	translation.put("advanced-settings", "تنظیمات پیشرفته")
	translation.put("chill", "دوستانه")
	translation.put("competitive", "رقابتی")
	translation.put("chill-alt", "درسته که سریع بودن جایزه داره، اما حالا یه کم یواش‌ترم باشی خیلی بد نیست.\nامتیاز پایه نسبتا بالاتره، روی عشق و حال تمرکز کن!")
	translation.put("competitive-alt", "هر چی سریع‌تر باشی، امتیاز بیشتری می‌گیری.\nامتیاز پایه خیلی پایین‌تره و سقوط سریع‌تره.")
	translation.put("score-calculation", "امتیازدهی")
	translation.put("word-language", "زبان واژه‌ها")
	translation.put("drawing-time-setting", "زمان نقاشی")
	translation.put("rounds-setting", "دورها")
	translation.put("max-players-setting", "بیشترین تعداد بازیکنان")
	translation.put("public-lobby-setting", "لابی همگانی")
	translation.put("custom-words", "واژه‌های سفارشی")
	translation.put("custom-words-info", "واژه‌های اضافی‌تونو وارد کنید،با کاما از هم جداشون کنید")
	translation.put("custom-words-placeholder", "فهرست, واژه‌های, جدا, شده, با, کاما")
	translation.put("custom-words-per-turn-setting", "واژه‌های سفارشی در هر نوبت")
	translation.put("players-per-ip-limit-setting", "حداکثر تعداد بازیکنن با یک IP")
	translation.put("save-settings", "ذخیره تنظیمات")
	translation.put("input-contains-invalid-data", "ورودی شامل داده اشتباهه:")
	translation.put("please-fix-invalid-input", "ورودی اشتباهو درست کنید و دوباره امتحان کنید.")
	translation.put("create-lobby", "ساخت لابی")
	translation.put("create-public-lobby", "ساخت لابی همگانی")
	translation.put("create-private-lobby", "ساخت لابی خصوصی")
	translation.put("no-lobbies-yet", "هنوز لابی همگانی نداریم.")
	translation.put("lobby-full", "ببخشید، ولی لابی پره.")
	translation.put("lobby-ip-limit-excceeded", "ببخشید، ولی شما تعداد دستگاه‌های مجاز با همین IP رو رد کردید.")
	translation.put("lobby-open-tab-exists", "به نظر میاد یه تب باز دیگه برای این لابی دارید.")
	translation.put("lobby-doesnt-exist", "لابی درخواستی وجود نداره")

	translation.put("refresh", "تازه کردن")
	translation.put("join-lobby", "ورود به لابی")

	translation.put("message-input-placeholder", "حدس‌ها و پیاماتو اینجا تایپ کن")

	translation.put("word-choice-warning", "اگه به موقع انتخاب نکنی این واژه انتخاب میشه")
	translation.put("choose-a-word", "یه واژه انتخاب کن")
	translation.put("waiting-for-word-selection", "در انتظار انتخاب واژه")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "داره واژه انتخاب می‌کنه.")

	translation.put("close-guess", "'%s' خیلی نزدیکه")
	translation.put("correct-guess", "شما واژه رو درست حدس زدید.")
	translation.put("correct-guess-other-player", "'%s' واژه رو درست حدس زد.")
	translation.put("round-over", "نوبت تموم شد، واژه‌ای انتخاب نشد.")
	translation.put("round-over-no-word", "نوبت تموم شد، واژه '%s' بود.")
	translation.put("game-over-win", "آفرین، تو بردی!")
	translation.put("game-over-tie", "بازی مساوی شد!")
	translation.put("game-over", "رتبه شما %s. با %s امتیاز")
	translation.put("drawer-disconnected", "دست زود‌تر تموم شد، ارتباط نقاش قطع شد.")
	translation.put("guessers-disconnected", "دست زودتر تموم شد، ارتباط بازیکنا قطع شد.")

	translation.put("change-active-color", "تغییر رنگ")
	translation.put("use-pencil", "استفاده از قلم")
	translation.put("use-eraser", "استفاده از پاک‌کن")
	translation.put("use-fill-bucket", "استفاده از سطل رنگ (ناحیه هدف رو با رنگ انتخاب شده رنگ می‌کنه)")
	translation.put("change-pencil-size-to", "تغییر اندازه قلم / پاک‌کن به %s")
	translation.put("clear-canvas", "پاک کردن بوم نقاشی")
	translation.put("undo", "برگردوندن آخرین تغییری که دادید (بعد از \""+translation.Get("clear-canvas")+"\" کار نمی‌کنه)")

	translation.put("connection-lost", "ارتباط قطع شد!")
	translation.put("connection-lost-text", "در تلاش برای برقراری ارتباط"+
		" ...\n\nمطمئن شید که اینترنتتون وصله.\nاگر "+
		"مشکل برقرار بود، با ادمین تماس بگیرید")
	translation.put("error-connecting", "خطا در برقراری ارتباط با سرور")
	translation.put("error-connecting-text",
		"Scribble.rs نتونست ارتباط رو برقرار کنه\n\nدر حالی که اینترنتتون "+
			"به نظر میاد که کار می‌کنه، یا شاید\nسرور یا دیواره‌آتشتون  "+
			"درست تنظیم نشده.\n\nبرای تلاش دوباره، صفحه رو دوباره باز کنید.")
	translation.put("message-too-long", "پیام خیلی طولانیه.")
	translation.put("server-shutting-down-title", "سرور در حال خاموش شدنه")
	translation.put("server-shutting-down-text", "ببخشید، اما سرور در حال خاموش شدنه. لطفا چند لحظه دیگه برگردید.")

	// Help dialog
	translation.put("controls", "کلیدا")
	translation.put("pencil", "قلم")
	translation.put("eraser", "پاک‌کن")
	translation.put("fill-bucket", "سطل رنگ")
	translation.put("switch-tools-intro", "شما می‌تونید ابزارها رو با استفاده از کلیدای میانبر عوض کنید")
	translation.put("switch-pencil-sizes", "همچنین می‌تونید اندازه قلم رو با کلیدای %s تا %s عوض کنید.")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "بستن")
	translation.put("no", "نه")
	translation.put("yes", "بله")
	translation.put("system", "سامانه")
	translation.put("confirm", "باشه")
	translation.put("ready", "آماده")
	translation.put("join", "ورود")
	translation.put("ongoing", "در حال برگزاری")
	translation.put("game-over-lobby", "بازی تموم شده")

	translation.put("source-code", "کد منبع")
	translation.put("help", "راهنمایی")
	translation.put("submit-feedback", "بازخورد")
	translation.put("stats", "وضعیت")

	translation.put("forbidden", "ممنوع")

	RegisterTranslation("fa", translation)

	return translation
}
