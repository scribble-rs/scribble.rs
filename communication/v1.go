package communication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/state"
)

//This file contains the API methods for the public API

// LobbyEntry is an API object for representing a join-able public lobby.
type LobbyEntry struct {
	LobbyID         string `json:"lobbyId"`
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
			LobbyID:         lobby.LobbyID,
			PlayerCount:     lobby.GetOccupiedPlayerSlots(),
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
	enableVotekick, enableVotekickInvalid := parseBoolean("enable votekick", r.Form.Get("enable_votekick"))
	publicLobby, publicLobbyInvalid := parseBoolean("public", r.Form.Get("public"))

	var requestErrors []string
	if languageInvalid != nil {
		requestErrors = append(requestErrors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		requestErrors = append(requestErrors, drawingTimeInvalid.Error())
	}
	if roundsInvalid != nil {
		requestErrors = append(requestErrors, roundsInvalid.Error())
	}
	if maxPlayersInvalid != nil {
		requestErrors = append(requestErrors, maxPlayersInvalid.Error())
	}
	if customWordsInvalid != nil {
		requestErrors = append(requestErrors, customWordsInvalid.Error())
	}
	if customWordChanceInvalid != nil {
		requestErrors = append(requestErrors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		requestErrors = append(requestErrors, clientsPerIPLimitInvalid.Error())
	}
	if enableVotekickInvalid != nil {
		requestErrors = append(requestErrors, enableVotekickInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		requestErrors = append(requestErrors, publicLobbyInvalid.Error())
	}

	if len(requestErrors) != 0 {
		http.Error(w, strings.Join(requestErrors, ";"), http.StatusBadRequest)
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

	lobbyData := createLobbyData(lobby)

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}

	//We only add the lobby if everything else was successful.
	state.AddLobby(lobby)
}

func enterLobby(w http.ResponseWriter, r *http.Request) {
	lobby, success := getLobbyWithErrorHandling(w, r)
	if !success {
		return
	}

	player := getPlayer(lobby, r)

	if player == nil {
		if !lobby.HasFreePlayerSlot() {
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

	lobbyData := createLobbyData(lobby)

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}
}

func editLobby(w http.ResponseWriter, r *http.Request) {
	lobby, success := getLobbyWithErrorHandling(w, r)
	if !success {
		return
	}

	userSession := getUserSession(r)
	if userSession == "" {
		http.Error(w, "no usersession supplied", http.StatusBadRequest)
		return
	}

	owner := lobby.Owner
	if owner == nil || owner.GetUserSession() != userSession {
		http.Error(w, "only the lobby owner can edit the lobby", http.StatusForbidden)
		return
	}

	var requestErrors []string

	//Uneditable properties
	if r.Form.Get("custom_words") != "" {
		requestErrors = append(requestErrors, "can't modify custom_words in existing lobby")
	}
	if r.Form.Get("language") != "" {
		requestErrors = append(requestErrors, "can't modify language in existing lobby")
	}
	//FIXME Make editable. As of now, changing this would require an update event for clients.
	if r.Form.Get("rounds") != "" {
		requestErrors = append(requestErrors, "can't modify rounds in existing lobby")
	}
	//FIXME Make editable. As of now, making this editable mid-turn would break score calculation.
	if r.Form.Get("drawing_time") != "" {
		requestErrors = append(requestErrors, "can't modify drawing_time in existing lobby")
	}

	parseError := r.ParseForm()
	if parseError != nil {
		http.Error(w, fmt.Sprintf("error parsing from (%s)", parseError), http.StatusBadRequest)
	}

	//Editable properties
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWordChance, customWordChanceInvalid := parseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := parseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick, enableVotekickInvalid := parseBoolean("enable votekick", r.Form.Get("enable_votekick"))
	publicLobby, publicLobbyInvalid := parseBoolean("public", r.Form.Get("public"))

	if maxPlayersInvalid != nil {
		requestErrors = append(requestErrors, maxPlayersInvalid.Error())
	}
	if customWordChanceInvalid != nil {
		requestErrors = append(requestErrors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		requestErrors = append(requestErrors, clientsPerIPLimitInvalid.Error())
	}
	if enableVotekickInvalid != nil {
		requestErrors = append(requestErrors, enableVotekickInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		requestErrors = append(requestErrors, publicLobbyInvalid.Error())
	}

	if len(requestErrors) != 0 {
		http.Error(w, strings.Join(requestErrors, ";"), http.StatusBadRequest)
		return
	}

	//While changing maxClientsPerIP and maxPlayers to a value lower than
	//is currently being used makes little sense, we'll allow it, as it doesn't
	//really break anything.

	lobby.MaxPlayers = maxPlayers
	lobby.CustomWordsChance = customWordChance
	lobby.ClientsPerIPLimit = clientsPerIPLimit
	lobby.EnableVotekick = enableVotekick
	lobby.Public = publicLobby

	TriggerUpdateEvent("lobby-settings-changed", lobby.EditableLobbySettings, lobby)
}

func getLobbyWithErrorHandling(w http.ResponseWriter, r *http.Request) (*game.Lobby, bool) {
	lobby, err := getLobby(r)
	if err != nil {
		if err == errNoLobbyIDSupplied {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err == errLobbyNotExistent {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return nil, false
	}

	return lobby, true
}

func lobbyEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		publicLobbies(w, r)
	} else if r.Method == http.MethodPatch {
		editLobby(w, r)
	} else if r.Method == http.MethodPost || r.Method == http.MethodPut {
		createLobby(w, r)
	} else {
		http.Error(w, fmt.Sprintf("method %s not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.Stats())
}
