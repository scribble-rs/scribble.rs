package communication

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/state"
)

//This file contains the API methods for the public API

// LobbyEntry is an API object for representing a join-able public lobby.
type LobbyEntry struct {
	ID              string `json:"id"`
	PlayerCount     int    `json:"playerCount"`
	MaxPlayers      int    `json:"maxPlayers"`
	Round           int    `json:"round"`
	MaxRounds       int    `json:"maxRounds"`
	DrawingTime     int    `json:"drawingTime"`
	CustomWords     bool   `json:"customWords"`
	Votekick        bool   `json:"votekick"`
	MaxClientsPerIP int    `json:"maxClientsPerIp"`
	Wordpack        string `json:"wordpack"`
}

func publicLobbies(w http.ResponseWriter, r *http.Request) {
	lobbies := state.GetPublicLobbies()
	lobbyEntries := make([]*LobbyEntry, 0, len(lobbies))
	for _, lobby := range lobbies {
		lobbyEntries = append(lobbyEntries, &LobbyEntry{
			ID:              lobby.ID,
			PlayerCount:     len(lobby.GetPlayers()),
			MaxPlayers:      lobby.MaxPlayers,
			Round:           lobby.Round,
			MaxRounds:       lobby.MaxRounds,
			DrawingTime:     lobby.DrawingTime,
			CustomWords:     len(lobby.CustomWords) > 0,
			Votekick:        lobby.EnableVotekick,
			MaxClientsPerIP: lobby.ClientsPerIPLimit,
			Wordpack:        lobby.Wordpack,
		})
	}
	encodingError := json.NewEncoder(w).Encode(lobbyEntries)
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
	publicLobby := r.Form.Get("public") == "true"

	var errors []string
	if languageInvalid != nil {
		errors = append(errors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		errors = append(errors, drawingTimeInvalid.Error())
	}
	if roundsInvalid != nil {
		errors = append(errors, roundsInvalid.Error())
	}
	if maxPlayersInvalid != nil {
		errors = append(errors, maxPlayersInvalid.Error())
	}
	if customWordsInvalid != nil {
		errors = append(errors, customWordsInvalid.Error())
	}
	if customWordChanceInvalid != nil {
		errors = append(errors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		errors = append(errors, clientsPerIPLimitInvalid.Error())
	}

	if len(errors) != 0 {
		http.Error(w, strings.Join(errors, ";"), http.StatusBadRequest)
		return
	}

	var playerName = getPlayername(r)
	player, lobby, createError := game.CreateLobby(playerName, language, publicLobby, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit, customWords, enableVotekick)
	if createError != nil {
		http.Error(w, createError.Error(), http.StatusBadRequest)
		return
	}

	player.SetLastKnownAddress(getIPAddressFromRequest(r))

	// Use the players generated usersession and pass it as a cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "usersession",
		Value:    player.GetUserSession(),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	lobbyData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}

	//We only add the lobby if everything else was successful.
	state.AddLobby(lobby)
}

func enterLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		if err == noLobbyIdSuppliedError {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err == lobbyNotExistentError {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	player := getPlayer(lobby, r)

	if player == nil {
		if len(lobby.GetPlayers()) >= lobby.MaxPlayers {
			http.Error(w, "lobby already full", http.StatusUnauthorized)
			return
		}

		var clientsWithSameIP int
		requestAddress := getIPAddressFromRequest(r)
		for _, otherPlayer := range lobby.GetPlayers() {
			if otherPlayer.GetLastKnownAddress() == requestAddress {
				clientsWithSameIP++
				if clientsWithSameIP >= lobby.ClientsPerIPLimit {
					http.Error(w, "maximum amount of newPlayer per IP reached", http.StatusUnauthorized)
					return
				}
			}
		}

		newPlayer := lobby.JoinPlayer(getPlayername(r))
		newPlayer.SetLastKnownAddress(getIPAddressFromRequest(r))

		// Use the players generated usersession and pass it as a cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "usersession",
			Value:    newPlayer.GetUserSession(),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		player.SetLastKnownAddress(getIPAddressFromRequest(r))
	}

	lobbyData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}
}

func lobbyEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		publicLobbies(w, r)
	} else {
		createLobby(w, r)
	}
}
