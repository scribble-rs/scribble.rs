package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/state"
)

//This file contains the API methods for the public API

var (
	ErrNoLobbyIDSupplied = errors.New("please supply a lobby id via the 'lobby_id' query parameter")
	ErrLobbyNotExistent  = errors.New("the requested lobby doesn't exist")
)

// LobbyEntry is an API object for representing a join-able public lobby.
type LobbyEntry struct {
	LobbyID         string `json:"lobbyId"`
	PlayerCount     int    `json:"playerCount"`
	MaxPlayers      int    `json:"maxPlayers"`
	Round           int    `json:"round"`
	Rounds          int    `json:"rounds"`
	DrawingTime     int    `json:"drawingTime"`
	CustomWords     bool   `json:"customWords"`
	Votekick        bool   `json:"votekick"`
	MaxClientsPerIP int    `json:"maxClientsPerIp"`
	Wordpack        string `json:"wordpack"`
}

func publicLobbies(w http.ResponseWriter, r *http.Request) {
	//REMARK: If paging is ever implemented, we might want to maintain order
	//when deleting lobbies from state in the state package.

	lobbies := state.GetPublicLobbies()
	lobbyEntries := make([]*LobbyEntry, 0, len(lobbies))
	for _, lobby := range lobbies {
		//While one would expect locking the lobby here, it's not very
		//important to get 100% consistent results here.
		lobbyEntries = append(lobbyEntries, &LobbyEntry{
			LobbyID:         lobby.LobbyID,
			PlayerCount:     lobby.GetOccupiedPlayerSlots(),
			MaxPlayers:      lobby.MaxPlayers,
			Round:           lobby.Round,
			Rounds:          lobby.Rounds,
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

	language, languageInvalid := ParseLanguage(r.Form.Get("language"))
	drawingTime, drawingTimeInvalid := ParseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := ParseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := ParseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := ParseCustomWords(r.Form.Get("custom_words"))
	customWordChance, customWordChanceInvalid := ParseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := ParseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick, enableVotekickInvalid := ParseBoolean("enable votekick", r.Form.Get("enable_votekick"))
	publicLobby, publicLobbyInvalid := ParseBoolean("public", r.Form.Get("public"))

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

	var playerName = GetPlayername(r)
	player, lobby, createError := game.CreateLobby(playerName, language, publicLobby, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit, customWords, enableVotekick)
	if createError != nil {
		http.Error(w, createError.Error(), http.StatusBadRequest)
		return
	}

	lobby.WriteJSON = WriteJSON
	player.SetLastKnownAddress(GetIPAddressFromRequest(r))

	SetUsersessionCookie(w, player)

	lobbyData := CreateLobbyData(lobby)

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}

	//We only add the lobby if everything else was successful.
	state.AddLobby(lobby)
}

func enterLobbyEndpoint(w http.ResponseWriter, r *http.Request) {
	lobby, success := getLobbyWithErrorHandling(w, r)
	if !success {
		return
	}

	var lobbyData *LobbyData

	lobby.Synchronized(func() {
		player := GetPlayer(lobby, r)

		if player == nil {
			if !lobby.HasFreePlayerSlot() {
				http.Error(w, "lobby already full", http.StatusUnauthorized)
				return
			}

			requestAddress := GetIPAddressFromRequest(r)

			if !lobby.CanIPConnect(requestAddress) {
				http.Error(w, "maximum amount of players per IP reached", http.StatusUnauthorized)
				return
			}

			newPlayer := lobby.JoinPlayer(GetPlayername(r))
			newPlayer.SetLastKnownAddress(requestAddress)

			// Use the players generated usersession and pass it as a cookie.
			SetUsersessionCookie(w, newPlayer)
		} else {
			player.SetLastKnownAddress(GetIPAddressFromRequest(r))
		}

		lobbyData = CreateLobbyData(lobby)
	})

	if lobbyData != nil {
		encodingError := json.NewEncoder(w).Encode(lobbyData)
		if encodingError != nil {
			http.Error(w, encodingError.Error(), http.StatusInternalServerError)
		}
	}
}

// SetUsersessionCookie takes the players usersession and sets it as a cookie.
func SetUsersessionCookie(w http.ResponseWriter, player *game.Player) {
	http.SetCookie(w, &http.Cookie{
		Name:     "usersession",
		Value:    player.GetUserSession(),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}

func editLobby(w http.ResponseWriter, r *http.Request) {
	userSession := GetUserSession(r)
	if userSession == "" {
		http.Error(w, "no usersession supplied", http.StatusBadRequest)
		return
	}

	lobby, success := getLobbyWithErrorHandling(w, r)
	if !success {
		return
	}

	parseError := r.ParseForm()
	if parseError != nil {
		http.Error(w, fmt.Sprintf("error parsing request query into form (%s)", parseError), http.StatusBadRequest)
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

	//Editable properties
	maxPlayers, maxPlayersInvalid := ParseMaxPlayers(r.Form.Get("max_players"))
	drawingTime, drawingTimeInvalid := ParseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := ParseRounds(r.Form.Get("rounds"))
	customWordChance, customWordChanceInvalid := ParseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := ParseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick, enableVotekickInvalid := ParseBoolean("enable votekick", r.Form.Get("enable_votekick"))
	publicLobby, publicLobbyInvalid := ParseBoolean("public", r.Form.Get("public"))

	owner := lobby.Owner
	if owner == nil || owner.GetUserSession() != userSession {
		http.Error(w, "only the lobby owner can edit the lobby", http.StatusForbidden)
		return
	}

	if maxPlayersInvalid != nil {
		requestErrors = append(requestErrors, maxPlayersInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		requestErrors = append(requestErrors, drawingTimeInvalid.Error())
	}
	if roundsInvalid != nil {
		requestErrors = append(requestErrors, roundsInvalid.Error())
	} else {
		currentRound := lobby.Round
		if rounds < currentRound {
			requestErrors = append(requestErrors, fmt.Sprintf("rounds must be greater than or equal to the current round (%d)", currentRound))
		}
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

	//We synchronize as late as possible to avoid unnecessary lags.
	//The previous code here isn't really prone to bugs due to lack of sync.
	lobby.Synchronized(func() {
		//While changing maxClientsPerIP and maxPlayers to a value lower than
		//is currently being used makes little sense, we'll allow it, as it doesn't
		//really break anything.

		lobby.MaxPlayers = maxPlayers
		lobby.CustomWordsChance = customWordChance
		lobby.ClientsPerIPLimit = clientsPerIPLimit
		lobby.EnableVotekick = enableVotekick
		lobby.Public = publicLobby
		lobby.Rounds = rounds

		if lobby.State == game.Ongoing {
			lobby.DrawingTimeNew = drawingTime
		} else {
			lobby.DrawingTime = drawingTime
		}

		lobbySettingsCopy := *lobby.EditableLobbySettings
		lobbySettingsCopy.DrawingTime = drawingTime
		lobby.TriggerUpdateEvent("lobby-settings-changed", lobbySettingsCopy)
	})
}

func getLobbyWithErrorHandling(w http.ResponseWriter, r *http.Request) (*game.Lobby, bool) {
	lobby, err := GetLobby(r)
	if err != nil {
		if err == ErrNoLobbyIDSupplied {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err == ErrLobbyNotExistent {
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

func statsEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.Stats())
}

// GetLobby extracts the lobby_id field from an HTTP request and searches
// the corresponding lobby. If the loby doesn't exist, or no ID has been
// supplied, we return an error.
func GetLobby(r *http.Request) (*game.Lobby, error) {
	lobbyID := r.URL.Query().Get("lobby_id")
	if lobbyID == "" {
		lobbyID = r.FormValue("lobby_id")
		if lobbyID == "" {
			return nil, ErrNoLobbyIDSupplied
		}
	}

	lobby := state.GetLobby(lobbyID)

	if lobby == nil {
		return nil, ErrLobbyNotExistent
	}

	return lobby, nil
}

var (
	//CanvasColor is the initially / empty canvas colors value used for
	//Lobbydata objects.
	CanvasColor = game.RGBColor{R: 255, G: 255, B: 255}
	//SuggestedBrushSizes is suggested brush sizes value used for
	//Lobbydata objects. A unit test makes sure these values are ordered
	//and within the specified bounds.
	SuggestedBrushSizes = [4]uint8{8, 16, 24, 32}
)

// LobbyData is the data necessary for correctly configuring a lobby.
// While unofficial clients will probably need all of these values, the
// official webclient doesn't use all of them as of now.
type LobbyData struct {
	*game.SettingBounds
	*game.EditableLobbySettings

	LobbyID string `json:"lobbyId"`
	//DrawingBoardBaseWidth is the internal canvas width and is needed for
	//correctly up- / downscaling drawing instructions.
	DrawingBoardBaseWidth int `json:"drawingBoardBaseWidth"`
	//DrawingBoardBaseHeight is the internal canvas height and is needed for
	//correctly up- / downscaling drawing instructions.
	DrawingBoardBaseHeight int `json:"drawingBoardBaseHeight"`
	//MinBrushSize is the minimum amount of pixels the brush can draw in.
	MinBrushSize int `json:"minBrushSize"`
	//MaxBrushSize is the maximum amount of pixels the brush can draw in.
	MaxBrushSize int `json:"maxBrushSize"`
	//CanvasColor is the initially (empty) color of the canvas.
	CanvasColor game.RGBColor `json:"canvasColor"`
	//SuggestedBrushSizes are suggestions for the different brush sizes
	//that the user can choose between. These brushes are guaranteed to
	//be ordered from low to high and stay with the bounds.
	SuggestedBrushSizes [4]uint8 `json:"suggestedBrushSizes"`
}

// CreateLobbyData creates a ready to use LobbyData object containing data
// from the passed Lobby.
func CreateLobbyData(lobby *game.Lobby) *LobbyData {
	return &LobbyData{
		SettingBounds:          game.LobbySettingBounds,
		EditableLobbySettings:  lobby.EditableLobbySettings,
		LobbyID:                lobby.LobbyID,
		DrawingBoardBaseWidth:  game.DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: game.DrawingBoardBaseHeight,
		MinBrushSize:           game.MinBrushSize,
		MaxBrushSize:           game.MaxBrushSize,
		CanvasColor:            CanvasColor,
		SuggestedBrushSizes:    SuggestedBrushSizes,
	}
}

// GetUserSession accesses the usersession from an HTTP request and
// returns the session. The session can either be in the cookie or in
// the header. If no session can be found, an empty string is returned.
func GetUserSession(r *http.Request) string {
	sessionCookie, noCookieError := r.Cookie("usersession")
	if noCookieError == nil && sessionCookie.Value != "" {
		return sessionCookie.Value
	}

	session, ok := r.Header["Usersession"]
	if ok {
		return session[0]
	}

	return ""
}

// GetPlayer returns the player object that matches the usersession in the
// supplied HTTP request and lobby. If no user session is set, we return nil.
func GetPlayer(lobby *game.Lobby, r *http.Request) *game.Player {
	return lobby.GetPlayer(GetUserSession(r))
}

// GetPlayername either retrieves the playername from a cookie, the URL form.
// If no preferred name can be found, we return an empty string.
func GetPlayername(r *http.Request) string {
	parseError := r.ParseForm()
	if parseError == nil {
		username := r.Form.Get("username")
		if username != "" {
			return username
		}
	}

	usernameCookie, noCookieError := r.Cookie("username")
	if noCookieError == nil {
		username := usernameCookie.Value
		if username != "" {
			return username
		}
	}

	return ""
}
