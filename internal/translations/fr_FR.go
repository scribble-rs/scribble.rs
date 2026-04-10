package translations

func initFrenchTranslation() *Translation {
	translation := createTranslation()

	translation.put("requires-js", "Ce site nécessite JavaScript pour fonctionner correctement.")

	translation.put("start-the-game", "Prêt !")
	translation.put("force-start", "Forcer le démarrage")
	translation.put("force-restart", "Forcer le redémarrage")
	translation.put("game-not-started-title", "La partie n'a pas commencé")
	translation.put("waiting-for-host-to-start", "Veuillez attendre que l'hôte du salon démarre la partie.")
	translation.put("click-to-homepage", "Cliquez ici pour revenir à la page d'accueil")

	translation.put("now-spectating-title", "Vous êtes maintenant spectateur")
	translation.put("now-spectating-text", "Vous pouvez quitter le mode spectateur en appuyant sur le bouton œil en haut.")
	translation.put("now-participating-title", "Vous participez maintenant")
	translation.put("now-participating-text", "Vous pouvez activer le mode spectateur en appuyant sur le bouton œil en haut.")

	translation.put("spectation-requested-title", "Mode spectateur demandé")
	translation.put("spectation-requested-text", "Vous serez spectateur après ce tour.")
	translation.put("participation-requested-title", "Participation demandée")
	translation.put("participation-requested-text", "Vous participerez après ce tour.")

	translation.put("spectation-request-cancelled-title", "Demande de mode spectateur annulée")
	translation.put("spectation-request-cancelled-text", "Votre demande de spectateur a été annulée, vous resterez participant.")
	translation.put("participation-request-cancelled-title", "Demande de participation annulée")
	translation.put("participation-request-cancelled-text", "Votre demande de participation a été annulée, vous resterez spectateur.")

	translation.put("round", "Manche")
	translation.put("toggle-soundeffects", "Activer/désactiver les effets sonores")
	translation.put("toggle-pen-pressure", "Activer/désactiver la pression du stylet")
	translation.put("change-your-name", "Pseudo")
	translation.put("randomize", "Aléatoire")
	translation.put("apply", "Appliquer")
	translation.put("save", "Enregistrer")
	translation.put("toggle-fullscreen", "Activer/désactiver le plein écran")
	translation.put("toggle-spectate", "Activer/désactiver le mode spectateur")
	translation.put("show-help", "Afficher l'aide")
	translation.put("votekick-a-player", "Voter pour expulser un joueur")

	translation.put("last-turn", "(Dernier tour : %s)")

	translation.put("drawer-kicked", "Comme le joueur expulsé dessinait, personne ne gagnera de points ce tour.")
	translation.put("self-kicked", "Vous avez été expulsé")
	translation.put("kick-vote", "(%s/%s) joueurs ont voté pour expulser %s.")
	translation.put("player-kicked", "Le joueur a été expulsé.")
	translation.put("owner-change", "%s est le nouveau propriétaire du salon.")

	translation.put("change-lobby-settings-tooltip", "Modifier les paramètres du salon")
	translation.put("change-lobby-settings-title", "Paramètres du salon")
	translation.put("lobby-settings-changed", "Paramètres du salon modifiés")
	translation.put("advanced-settings", "Paramètres avancés")
	translation.put("chill", "Détente")
	translation.put("competitive", "Compétitif")
	translation.put("chill-alt", "La rapidité est récompensée, mais ce n'est pas grave si vous êtes un peu plus lent.\nLe score de base est assez élevé, l'essentiel est de s'amuser !")
	translation.put("competitive-alt", "Plus vous êtes rapide, plus vous gagnerez de points.\nLe score de base est bien plus bas et la baisse est plus rapide.")
	translation.put("score-calculation", "Score")
	translation.put("word-language", "Langue")
	translation.put("drawing-time-setting", "Temps de dessin")
	translation.put("rounds-setting", "Manches")
	translation.put("max-players-setting", "Joueurs maximum")
	translation.put("public-lobby-setting", "Salon public")
	translation.put("custom-words", "Mots personnalisés")
	translation.put("custom-words-info", "Saisissez vos mots supplémentaires en les séparant par des virgules")
	translation.put("custom-words-placeholder", "Liste, de, mots, séparés, par, des, virgules")
	translation.put("custom-words-per-turn-setting", "Mots personnalisés par tour")
	translation.put("players-per-ip-limit-setting", "Limite de joueurs par IP")
	translation.put("words-per-turn-setting", "Mots par tour")
	translation.put("save-settings", "Enregistrer les paramètres")
	translation.put("input-contains-invalid-data", "Votre saisie contient des données invalides :")
	translation.put("please-fix-invalid-input", "Corrigez la saisie invalide et réessayez.")
	translation.put("create-lobby", "Créer un salon")
	translation.put("create-public-lobby", "Créer un salon public")
	translation.put("create-private-lobby", "Créer un salon privé")
	translation.put("no-lobbies-yet", "Il n'y a encore aucun salon.")
	translation.put("lobby-full", "Désolé, le salon est complet.")
	translation.put("lobby-ip-limit-excceeded", "Désolé, vous avez dépassé le nombre maximal de clients par IP.")
	translation.put("lobby-open-tab-exists", "Il semble qu'un onglet pour ce salon soit déjà ouvert.")
	translation.put("lobby-doesnt-exist", "Le salon demandé n'existe pas")

	translation.put("refresh", "Actualiser")
	translation.put("join-lobby", "Rejoindre le salon")

	translation.put("message-input-placeholder", "Tapez ici vos propositions et messages")

	translation.put("word-choice-warning", "Mot choisi si vous ne décidez pas à temps")
	translation.put("choose-a-word", "Choisissez un mot")
	translation.put("waiting-for-word-selection", "En attente du choix du mot")
	// This one doesn't use %s, since we want to make one part bold.
	translation.put("is-choosing-word", "choisit un mot.")

	translation.put("close-guess", "'%s' est très proche.")
	translation.put("correct-guess", "Vous avez correctement deviné le mot.")
	translation.put("correct-guess-other-player", "'%s' a correctement deviné le mot.")
	translation.put("round-over", "Tour terminé, aucun mot n'a été choisi.")
	translation.put("round-over-no-word", "Tour terminé, le mot était '%s'.")
	translation.put("game-over-win", "Félicitations, vous avez gagné !")
	translation.put("game-over-tie", "Égalité !")
	translation.put("game-over", "Vous avez terminé %s. avec %s points")
	translation.put("drawer-disconnected", "Tour terminé prématurément, le dessinateur s'est déconnecté.")
	translation.put("guessers-disconnected", "Tour terminé prématurément, les devineurs se sont déconnectés.")
	translation.put("word-hint-revealed", "Un indice du mot a été révélé !")

	translation.put("change-active-color", "Changer votre couleur active")
	translation.put("use-pencil", "Utiliser le crayon")
	translation.put("use-eraser", "Utiliser la gomme")
	translation.put("use-fill-bucket", "Utiliser le pot de peinture (Remplit la zone ciblée avec la couleur sélectionnée)")
	translation.put("change-pencil-size-to", "Changer la taille du crayon / de la gomme à %s")
	translation.put("clear-canvas", "Effacer le canevas")
	translation.put("undo", "Annuler la dernière modification effectuée (Ne fonctionne pas après \""+translation.Get("clear-canvas")+"\")")

	translation.put("connection-lost", "Connexion perdue !")
	translation.put("connection-lost-text", "Tentative de reconnexion"+
		" ...\n\nAssurez-vous que votre connexion internet fonctionne.\nSi le "+
		"problème persiste, contactez le webmaster.")
	translation.put("error-connecting", "Erreur de connexion au serveur")
	translation.put("error-connecting-text",
		"Scribble.rs n'a pas pu établir une connexion socket.\n\nBien que votre connexion internet "+
			"semble fonctionner, soit le\nserveur, soit votre pare-feu n'a pas "+
			"été configuré correctement.\n\nPour réessayer, rechargez la page.")
	translation.put("message-too-long", "Votre message est trop long.")
	translation.put("server-shutting-down-title", "Arrêt du serveur")
	translation.put("server-shutting-down-text", "Désolé, le serveur va s'arrêter. Merci de revenir plus tard.")

	// Help dialog
	translation.put("controls", "Contrôles")
	translation.put("pencil", "Crayon")
	translation.put("eraser", "Gomme")
	translation.put("fill-bucket", "Pot de peinture")
	translation.put("switch-pencil-sizes", "Vous pouvez aussi changer la taille du crayon avec les touches %s à %s.")

	// Generic words
	// "close" as in "closing the window"
	translation.put("close", "Fermer")
	translation.put("no", "Non")
	translation.put("yes", "Oui")
	translation.put("system", "Système")
	translation.put("confirm", "OK")
	translation.put("ready", "Prêt")
	translation.put("join", "Rejoindre")
	translation.put("ongoing", "En cours")
	translation.put("game-over-lobby", "Partie terminée")

	translation.put("source-code", "Code source")
	translation.put("help", "Aide")
	translation.put("submit-feedback", "Retour")
	translation.put("stats", "Statut")

	translation.put("forbidden", "Interdit")

	RegisterTranslation("fr", translation)

	return translation
}
