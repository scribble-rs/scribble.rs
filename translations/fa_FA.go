package translations

func initPersianTranslation() Translation {
	translation := createTranslation()

	translation.put("requires-js", "این وبسایت به جاوا اسکریپت نیاز دارن")

	translation.put("start-the-game", "شروع بازی")
	translation.put("start", "شروع")
	translation.put("game-not-started-title", "بازی شروع نشده")
	translation.put("waiting-for-host-to-start", "لطفا منتظر بمانید بازی شروع شود")

	translation.put("round", "دور")
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

	translation.put("change-lobby-settings", "تغییر تنظیمات اتاق")
	translation.put("lobby-settings-changed", "تنظیمات اتاق تغییر کرد")
	translation.put("advanced-settings", "تنظیمات پیشرفته")
	translation.put("word-language", "زبان کلمات")
	translation.put("game-host", "مدیر")
	translation.put("drawing-time-setting", "زمان هر نقاشی")
	translation.put("rounds-setting", "تعداد نوبت ها")
	translation.put("max-players-setting", "حداکثر تعداد بازیکنان")
	translation.put("public-lobby-setting", "اتاق عمومی")
	translation.put("custom-words", "کلمات خودی")
	translation.put("custom-words-info", "کلمات خود را با - (خط فاصله) از هم جدا کنید")
	translation.put("custom-words-chance-setting", "شانس استفاده از کلمات خودی")
	translation.put("players-per-ip-limit-setting", "تعداد بازیکن از آیپی یکسان")
	translation.put("enable-votekick-setting", "قابلیت کیک کردن بازیکنان")
	translation.put("save-settings", "ذخیره تنظیمات")
	translation.put("input-contains-invalid-data", "ورودی های شما مشکلات زیر را دارد:")
	translation.put("please-fix-invalid-input", "مشکلات ذکر شده را حل کنید و دوباره تلاش کنید.")
	translation.put("create-lobby", "ساخت اتاق")

	translation.put("players", "بازیکنان")
	translation.put("refresh", "ریفرش")
	translation.put("join-lobby", "ورود به اتاق")

	translation.put("message-input-placeholder", "حدس ها و پیام هاتونو اینجا تایپ کنید")

	translation.put("choose-a-word", "یک کلمه انتخاب کنید")
	translation.put("waiting-for-word-selection", "نقاش در حال انتخاب کلمه میباشد")
	//This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "در حال انتخاب کلمه است")

	translation.put("close-guess", "کلمه '%s' خیلی نزدی بود")
	translation.put("correct-guess", "شما کلمه را حدس زدید.")
	translation.put("correct-guess-other-player", "%s کلمه را حدس زد")
	translation.put("round-over", "نوبت تموم شد، کلمه ای انتخاب نشد")
	translation.put("round-over-no-word", "نوبت تموم شد، کلمه '%s' بود")
	translation.put("game-over-win", "تبریک! شمار برنده شدید")
	translation.put("game-over", "شما نفر %s م شدید با %s امتیاز.")

	translation.put("change-active-color", "تغییر رنگ")
	translation.put("use-pencil", "مداد")
	translation.put("use-eraser", "پاککن")
	translation.put("use-fill-bucket", "سطل رنگ")
	translation.put("change-pencil-size-to", "تغییر سایز مداد به %s")
	translation.put("clear-canvas", "پاک کردن نقاشی")

	translation.put("connection-lost", "ارتباط قطع شد!")
	translation.put("connection-lost-text", "در حال تلاش برای اتصال مجدد...")
	translation.put("error-connecting", "اتصال به سرور موفقیت آمیز نبود")
	translation.put("error-connecting-text",
		"سایت بکش بکش نمیتواند یک ارتباط سوکتی برقرار کند\nاین مشکل میتواند از فایروال شما یا از سمت سرور باشد.")

	//Generic words
	//As "close" in "closing the window"
	translation.put("close", "بستن")
	translation.put("no", "خیر")
	translation.put("yes", "بله")
	translation.put("system", "سیستم")

	translation.put("source-code", "سورس")
	translation.put("help", "راهنما")
	translation.put("contact", "تماس با ما")
	translation.put("submit-feedback", "پیشنهاد یا گزارش باگ")
	translation.put("stats", "استاتوس")

	RegisterTranslation("fa", translation)

	return translation
}
