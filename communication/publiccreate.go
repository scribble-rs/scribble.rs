package communication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

//This file contains the API methods for the public API

func enterLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	player := getPlayer(lobby, r)

	lobbyData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	if player == nil {
		if len(lobby.Players) >= lobby.MaxPlayers {
			http.Error(w, "lobby already full", http.StatusUnauthorized)
			return
		}

		matches := 0
		for _, otherPlayer := range lobby.Players {
			socket := otherPlayer.GetWebsocket()
			if socket != nil && remoteAddressToSimpleIP(socket.RemoteAddr().String()) == remoteAddressToSimpleIP(r.RemoteAddr) {
				matches++
			}
		}

		if matches >= lobby.ClientsPerIPLimit {
			http.Error(w, "maximum amount of player per IP reached", http.StatusUnauthorized)
			return
		}

		var playerName = getPlayername(r)
		userSession := lobby.JoinPlayer(playerName)

		// Use the players generated usersession and pass it as a cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "usersession",
			Value:    userSession,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	}

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}
}

func createLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		http.Error(w, formParseError.Error(), http.StatusBadRequest)
		return
	}

	language, languageInvalid := parseLanguage(r.Form.Get("language"))
	drawingTime, drawingTimeInvalid := parseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := parseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := parseCustomWords(r.Form.Get("custom_words"))
	customWordChance, customWordChanceInvalid := parseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := parseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick := r.Form.Get("enable_votekick") == "true"

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := CreatePageData{
		SettingBounds:     game.LobbySettingBounds,
		Languages:         game.SupportedLanguages,
		DrawingTime:       r.Form.Get("drawing_time"),
		Rounds:            r.Form.Get("rounds"),
		MaxPlayers:        r.Form.Get("max_players"),
		CustomWords:       r.Form.Get("custom_words"),
		CustomWordsChance: r.Form.Get("custom_words_chance"),
		ClientsPerIPLimit: r.Form.Get("clients_per_ip_limit"),
		EnableVotekick:    r.Form.Get("enable_votekick"),
		Language:          r.Form.Get("language"),
	}

	if languageInvalid != nil {
		pageData.Errors = append(pageData.Errors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		pageData.Errors = append(pageData.Errors, drawingTimeInvalid.Error())
	}
	if roundsInvalid != nil {
		pageData.Errors = append(pageData.Errors, roundsInvalid.Error())
	}
	if maxPlayersInvalid != nil {
		pageData.Errors = append(pageData.Errors, maxPlayersInvalid.Error())
	}
	if customWordsInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordsInvalid.Error())
	}
	if customWordChanceInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		pageData.Errors = append(pageData.Errors, clientsPerIPLimitInvalid.Error())
	}

	if len(pageData.Errors) != 0 {
		http.Error(w, strings.Join(pageData.Errors, ";"), http.StatusBadRequest)
		return
	}

	var playerName = getPlayername(r)
	session, lobby, createError := game.CreateLobby(playerName, language, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit, customWords, enableVotekick)
	if createError != nil {
		http.Error(w, createError.Error(), http.StatusBadRequest)
		return
	}

	// Use the players generated usersession and pass it as a cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "usersession",
		Value:    session,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	_, encodingError := fmt.Fprint(w, lobby.ID)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}
}
