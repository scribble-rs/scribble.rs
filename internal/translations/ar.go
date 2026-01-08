package translations

func initArabicTranslation() Translation {
	translation := createTranslation()
	translation.IsRtl = true

	translation.put("requires-js", "يحتاج هذا الموقع لصلاحيات جافاسكريبت ليعمل.")

	translation.put("start-the-game", "إستعدوا!")
	translation.put("force-start", "البدء الإجباري")
	translation.put("force-restart", "الإعادة الإجبارية")
	translation.put("game-not-started-title", "لم تبدإ اللعبة بعد")
	translation.put("waiting-for-host-to-start", "الرجاء إنتظار المنظم لبدإ اللعبة")

	translation.put("now-spectating-title", "أنت الآن متفرج")
	translation.put("now-spectating-text", "يمكنك الخروج من وضع المتفرج بالضغط على زر العين أعلاه")
	translation.put("now-participating-title", "أنت الآن مشارك")
	translation.put("now-participating-text", "يمكنك دخول وضع المتفرج بالضغط على زر العين أعلاه")

	translation.put("spectation-requested-title", "تم طلب وضع المتفرج")
	translation.put("spectation-requested-text", "ستكون متفرجا بعد هذه الجولة.")
	translation.put("participation-requested-title", "تم طلب المشاركة")
	translation.put("participation-requested-text", "ستكون مشاركا بعد هذه الجولة.")

	translation.put("spectation-request-cancelled-title", "تم إلغاء طلب وضع المتفرج")
	translation.put("spectation-request-cancelled-text", "تم إلغاء وضع المتفرج, ستبقى مشاركا.")
	translation.put("participation-request-cancelled-title", "تم إلغاء طلب المشاركة")
	translation.put("participation-request-cancelled-text", "تم إلغاء طلب المشاركة ستبقى متفرجا.")

	translation.put("round", "الجولة")
	translation.put("toggle-soundeffects", "تشغيل/إيقاف مؤثرات الصوتية")
	translation.put("toggle-pen-pressure", "تشغيل/إيقاف ضغط القلم")
	translation.put("change-your-name", "إسم الشهرة")
	translation.put("randomize", "إختيار عشوائي")
	translation.put("apply", "تطبيق")
	translation.put("save", "حفظ")
	translation.put("toggle-fullscreen", "وضع ملء الشاشة")
	translation.put("toggle-spectate", "تشغيل/إيقاف وضع المتفرج")
	translation.put("show-help", "عرض المساعدة")
	translation.put("votekick-a-player", "تصويت لطرد اللاعب")

	translation.put("last-turn", "(آخر دور %s)")

	translation.put("drawer-kicked", "بما أن اللاعب المطرود كان يرسم, لن يأخذ أي منكم نقاط")
	translation.put("self-kicked", "تم طردك")
	translation.put("kick-vote", "(%s/%s) اللاعبين الذين صوتو لطرد %s.")
	translation.put("player-kicked", "اللاعب تم طرده.")
	translation.put("owner-change", "%s هو الآن مالك الردهة")

	translation.put("change-lobby-settings-tooltip", "تغيير إعدادات الردهة")
	translation.put("change-lobby-settings-title", "إعدادات الردهة")
	translation.put("lobby-settings-changed", "تم تغيير إعدادات الردهة")
	translation.put("advanced-settings", "إعدادات متقدمة")
	translation.put("chill", "هادئة")
	translation.put("competitive", "تنافسية")
	translation.put("chill-alt", "مع أن السرعة تكافأ, إلا أنه لا بأس إن كنت أبطأ قليلا.\nالنقاط الأساسية مرتفعة نسبيا, ركز على الإستمتاع!")
	translation.put("competitive-alt", "كلما كنت أسرع, حصلت على نقاط أكثر.\nالنقاط الأساسية أقل بكثير, و الإنخفاض أسرع.")
	translation.put("score-calculation", "النقاط")
	translation.put("word-language", "اللغة")
	translation.put("drawing-time-setting", "وقت الرسم")
	translation.put("rounds-setting", "الجولات")
	translation.put("max-players-setting", "الحد الأقصى للاعبين")
	translation.put("public-lobby-setting", "ردهة عامة")
	translation.put("custom-words", "كلمات مخصصة")
	translation.put("custom-words-info", "أضف كلماتك الخاصة, فرق بينها بالفاصلة")
	translation.put("custom-words-per-turn-setting", "كلمات مخصصة لكل دور")
	translation.put("players-per-ip-limit-setting", "حد اللاعبين في كل IP")
	translation.put("save-settings", "Save settings")
	translation.put("input-contains-invalid-data", "مدخلاتك تحتوي على معلومات خاطئة")
	translation.put("please-fix-invalid-input", "صحح المدخل الخاطئ و أعد المحاولة")
	translation.put("create-lobby", "إنشاء ردهة")
	translation.put("create-public-lobby", "إنشاء ردهة عامة")
	translation.put("create-private-lobby", "إنشاء ردهة خاصة")

	translation.put("refresh", "إعادة تحميل")
	translation.put("join-lobby", "دخول الردهة")

	translation.put("message-input-placeholder", "أدخل تخميناتك و رسائلك هنا!")

	translation.put("word-choice-warning", "كلمة إذا لم تختر في الوقت")
	translation.put("choose-a-word", "إختر كلمة")
	translation.put("waiting-for-word-selection", "في إنتظار إختيار كلمة")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "يختار كلمة.")

	translation.put("close-guess", "'%s' قريب جدا")
	translation.put("correct-guess", "لقد أصبت في تخمينك للكلمة.")
	translation.put("correct-guess-other-player", "'%s' اصاب في تخمين الكلمة")
	translation.put("round-over", "إنتهى دورك, لم يتم إختيار أي كلمة.")
	translation.put("round-over-no-word", "إنتهى دورك, الكلمة كانت '%s'")
	translation.put("game-over-win", "مبارك, لقد فزت!")
	translation.put("game-over-tie", "إنه تعادل")
	translation.put("game-over", "لقد وضعت %s. برصيد %s نقطة.")

	translation.put("change-active-color", "غير لونك الحالي")
	translation.put("use-pencil", "إستخدم القلم")
	translation.put("use-eraser", "إستخدم الممحاة")
	translation.put("use-fill-bucket", "إستعمل الملء بالدلو (يقوم بملء مساحة معينة باللون المعين)")
	translation.put("change-pencil-size-to", "غير حجم القلم/الممحاة %s")
	translation.put("clear-canvas", "مسح مساحة الرسم")
	translation.put("undo", "التراجع عن آخر تغيير قمت به  ( لا تعمل بعد \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "فقدت الإشارة")
	translation.put("connection-lost-text", "محاولة الإتصال مجددا"+
		" ...\n\nتأكد من إتصالك بالإنترنت\nإذا "+
		"بقي المشكل على حاله, تواصل مع المالك")
	translation.put("error-connecting", "خطأ في الإتصال بالخادم")
	translation.put("error-connecting-text",
		"Scribble.rs لا يمكنه الإتصال بالخادم\n\nبما أن إتصالك "+
			"بالأنترنت يبدو أنه يعمل, فإن\nالخادم أو الجدار الناري لم  "+
			"يتم ضبطه بشكل صحيح\n\nلإعادة المحاولة, قم بإعادة تحميل الصفحة.")

	translation.put("message-too-long", "رسالتك طويلة جدا!")

	// Help dialog
	translation.put("controls", "التحكم")
	translation.put("pencil", "القلم")
	translation.put("eraser", "الممحاة")
	translation.put("fill-bucket", "الملء بالدلو")
	translation.put("switch-tools-intro", "يمكنك التحويل بين الأدوات بواسطة الاختصارات")
	translation.put("switch-pencil-sizes", "يمكنك التغيير بين احجام القلم من %s إلى %s.")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "إغلاق")
	translation.put("no", "لا")
	translation.put("yes", "نعم")
	translation.put("system", "النظام")

	translation.put("source-code", "مصدر الكود")
	translation.put("help", "المساعدة")
	translation.put("submit-feedback", "الآراء")
	translation.put("stats", "الحالة")

	RegisterTranslation("ar", translation)

	return translation
}
