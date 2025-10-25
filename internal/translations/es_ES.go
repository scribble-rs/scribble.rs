package translations

func initSpainTranslation() Translation {
	translation := createTranslation()

	translation.put("requires-js", "Este sitio web requiere JavaScript para funcionar correctamente..")

	translation.put("start-the-game", "Prepárate!")
	translation.put("force-start", "Inicio forzado")
	translation.put("force-restart", "Force RestartReinicio forzado")
	translation.put("game-not-started-title", "El juego no ha comenzado")
	translation.put("waiting-for-host-to-start", "Por favor, espere a que el anfitrión de su lobby inicie el juego..")

	translation.put("now-spectating-title", "Ahora estás observando")
	translation.put("now-spectating-text", "Puedes salir del modo espectador presionando el botón del ojo en la parte superior.")
	translation.put("now-participating-title", "Ahora estás participando")
	translation.put("now-participating-text", "Puedes ingresar al modo espectador presionando el botón del ojo en la parte superior.")

	translation.put("spectation-requested-title", "Se solicita modo espectador")
	translation.put("spectation-requested-text", "Serás espectador después de este turno.")
	translation.put("participation-requested-title", "Se solicita participación")
	translation.put("participation-requested-text", "Participarás después de este turno..")

	translation.put("spectation-request-cancelled-title", "Se solicitó el modo espectador cancelado")
	translation.put("spectation-request-cancelled-text", "Tu solicitud de espectación ha sido cancelada, seguirás participando..")
	translation.put("participation-request-cancelled-title", "Participación solicitada cancelada")
	translation.put("participation-request-cancelled-text", "Tu solicitud de participación ha sido cancelada, continuarás viendo..")

	translation.put("round", "Redondo")
	translation.put("toggle-soundeffects", "Activar o desactivar efectos de sonido")
	translation.put("toggle-pen-pressure", "Alternar la presión del lápiz")
	translation.put("change-your-name", "Apodo")
	translation.put("randomize", "Aleatorizar")
	translation.put("apply", "Aplicar")
	translation.put("save", "Ahorrar")
	translation.put("toggle-fullscreen", "Cambiar a pantalla completa")
	translation.put("toggle-spectate", "Activar o desactivar el modo espectador")
	translation.put("show-help", "Mostrar ayuda")
	translation.put("votekick-a-player", "Votar para patear a un jugador")

	translation.put("last-turn", "(Último turno: %s)")

	translation.put("drawer-kicked", "Dado que el jugador expulsado ha estado robando, ninguno de ustedes obtendrá puntos en esta ronda.")
	translation.put("self-kicked", "Te han pateado")
	translation.put("kick-vote", "(%s/%s) Los jugadores votaron para patear %s.")
	translation.put("player-kicked", "El jugador ha sido expulsado.")
	translation.put("owner-change", "%s es el nuevo dueño del lobby.")

	translation.put("change-lobby-settings-tooltip", "Cambiar la configuración del lobby")
	translation.put("change-lobby-settings-title", "Configuración del lobby")
	translation.put("lobby-settings-changed", "Se cambiaron las configuraciones del lobby")
	translation.put("advanced-settings", "Configuración avanzada")
	translation.put("chill", "Enfriar")
	translation.put("competitive", "Competitivo")
	translation.put("chill-alt", "Aunque ser rápido tiene recompensa, no es tan malo si eres un poco más lento..\nLa puntuación base es bastante alta, concéntrate en divertirte.!")
	translation.put("competitive-alt", "Cuanto más rápido seas, más puntos obtendrás..\nLa puntuación base es mucho más baja y el descenso es más rápido..")
	translation.put("score-calculation", "Tanteo")
	translation.put("word-language", "Idioma")
	translation.put("drawing-time-setting", "Tiempo de dibujo")
	translation.put("rounds-setting", "Rondas")
	translation.put("max-players-setting", "Máximo de jugadores")
	translation.put("public-lobby-setting", "Vestíbulo público")
	translation.put("custom-words", "Palabras personalizadas")
	translation.put("custom-words-info", "Ingrese sus palabras adicionales, separándolas por comas")
	translation.put("custom-words-per-turn-setting", "Palabras personalizadas por turno")
	translation.put("players-per-ip-limit-setting", "Jugadores por límite de IP")
	translation.put("save-settings", "Guardar configuración")
	translation.put("input-contains-invalid-data", "Su entrada contiene datos no válidos:")
	translation.put("please-fix-invalid-input", "Corrija la entrada no válida y vuelva a intentarlo.")
	translation.put("create-lobby", "Crear lobby")
	translation.put("create-public-lobby", "Crear un lobby público")
	translation.put("create-private-lobby", "Crear un lobby privado")

	translation.put("players", "Jugadores")
	translation.put("refresh", "Refrescar")
	translation.put("join-lobby", "Únase al lobby")

	translation.put("message-input-placeholder", "Escribe tus conjeturas y mensajes aquí")

	translation.put("word-choice-warning", "Palabra si no eliges a tiempo")
	translation.put("choose-a-word", "Elige una palabra")
	translation.put("waiting-for-word-selection", "Esperando la selección de palabras")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "está eligiendo una palabra.")

	translation.put("close-guess", "'%s' está muy cerca.")
	translation.put("correct-guess", "Has adivinado correctamente la palabra..")
	translation.put("correct-guess-other-player", "'%s' adivinó correctamente la palabra.")
	translation.put("round-over", "Dé la vuelta, no se eligió ninguna palabra.")
	translation.put("round-over-no-word", "Al dar la vuelta, la palabra era '%s'.")
	translation.put("game-over-win", "¡Enhorabuena, has ganado!")
	translation.put("game-over-tie", "Es un empate!")
	translation.put("game-over", "Quedaste en %s lugar. Con %s puntos")

	translation.put("change-active-color", "Cambia tu color activo")
	translation.put("use-pencil", "Usa lápiz")
	translation.put("use-eraser", "Usar borrador")
	translation.put("use-fill-bucket", "UUtilice el cubo de llenado (Rellena el área objetivo con el color seleccionado)")
	translation.put("change-pencil-size-to", "Cambia el lápiz / tamaño del borrador a %s")
	translation.put("clear-canvas", "Limpiar el lienzo")
	translation.put("undo", "Revertir el último cambio realizado (No funciona después \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "Conexión perdida!")
	translation.put("connection-lost-text", "Intentando reconectarse"+
		" ...\n\nAsegúrese de que su conexión a Internet funcione.\n"+
		"Si el problema persiste, contacte con el webmaster..")
	translation.put("error-connecting", "Error al conectar con el servidor")
	translation.put("error-connecting-text",
		"Scribble.rs no se pudo establecer una conexión de socket.\n\nAunque su conexión "+
			"a Internet parece funcionar, es posible que el servidor\no su firewall no se hayan "+
			"configurado correctamente.\n\nPara volver a intentarlo, recarga la página.")

	translation.put("message-too-long", "Tu mensaje es demasiado largo.")

	// Help dialog
	translation.put("controls", "Controles")
	translation.put("pencil", "Lápiz")
	translation.put("eraser", "Borrador")
	translation.put("fill-bucket", "Llenar el cubo")
	translation.put("switch-tools-intro", "Puede cambiar entre herramientas mediante atajos")
	translation.put("switch-pencil-sizes", "También puedes cambiar entre tamaños de lápiz usando teclas %s para %s.")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "Cerca")
	translation.put("no", "No")
	translation.put("yes", "Sí")
	translation.put("system", "Sistema")

	translation.put("source-code", "Código fuente")
	translation.put("help", "Ayuda")
	translation.put("submit-feedback", "Comentario")
	translation.put("stats", "Estado")

	RegisterTranslation("es", translation)
	RegisterTranslation("es-es", translation)

	return translation
}
