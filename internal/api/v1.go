//go:generate easyjson -all ${GOFILE}

// This file contains the API methods for the public API

package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/mailru/easyjson"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/state"
)

var ErrLobbyNotExistent = errors.New("the requested lobby doesn't exist")

//easyjson:skip
type V1Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *V1Handler {
	return &V1Handler{
		cfg: cfg,
	}
}

//easyjson:json
type LobbyEntries []*LobbyEntry

// LobbyEntry is an API object for representing a join-able public lobby.
type LobbyEntry struct {
	LobbyID             string     `json:"lobbyId"`
	Wordpack            string     `json:"wordpack"`
	State               game.State `json:"state"`
	PlayerCount         int        `json:"playerCount"`
	MaxPlayers          int        `json:"maxPlayers"`
	Round               int        `json:"round"`
	Rounds              int        `json:"rounds"`
	DrawingTime         int        `json:"drawingTime"`
	MaxClientsPerIP     int        `json:"maxClientsPerIp"`
	CustomWords         bool       `json:"customWords"`
	WordSelectCount     int        `json:"wordSelectCount"`
}

func (handler *V1Handler) getLobbies(writer http.ResponseWriter, _ *http.Request) {
	// REMARK: If paging is ever implemented, we might want to maintain order
	// when deleting lobbies from state in the state package.

	lobbies := state.GetPublicLobbies()
	lobbyEntries := make(LobbyEntries, 0, len(lobbies))
	for _, lobby := range lobbies {
		// While one would expect locking the lobby here, it's not very
		// important to get 100% consistent results here.
		lobbyEntries = append(lobbyEntries, &LobbyEntry{
			LobbyID:         lobby.LobbyID,
			PlayerCount:     lobby.GetOccupiedPlayerSlots(),
			MaxPlayers:      lobby.MaxPlayers,
			Round:           lobby.Round,
			Rounds:          lobby.Rounds,
			DrawingTime:     lobby.DrawingTime,
			WordSelectCount: lobby.WordSelectCount,
			CustomWords:     len(lobby.CustomWords) > 0,
			MaxClientsPerIP: lobby.ClientsPerIPLimit,
			Wordpack:        lobby.Wordpack,
			State:           lobby.State,
		})
	}

	if started, _, err := easyjson.MarshalToHTTPResponseWriter(lobbyEntries, writer); err != nil {
		if !started {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func (handler *V1Handler) postLobby(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	language, languageInvalid := ParseLanguage(request.Form.Get("language"))
	drawingTime, drawingTimeInvalid := ParseDrawingTime(request.Form.Get("drawing_time"))
	wordSelectCount, wordSelectCountInvalid := ParseWordSelectCount(request.Form.Get("word_select_count"))
	rounds, roundsInvalid := ParseRounds(request.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := ParseMaxPlayers(request.Form.Get("max_players"))
	customWords, customWordsInvalid := ParseCustomWords(request.Form.Get("custom_words"))
	customWordsPerTurn, customWordsPerTurnInvalid := ParseCustomWordsPerTurn(request.Form.Get("custom_words_per_turn"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := ParseClientsPerIPLimit(request.Form.Get("clients_per_ip_limit"))
	publicLobby, publicLobbyInvalid := ParseBoolean("public", request.Form.Get("public"))
	timerStart, timerStartInvalid := ParseBoolean("timerStart", request.Form.Get("timer_start"))

	var requestErrors []string
	if languageInvalid != nil {
		requestErrors = append(requestErrors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		requestErrors = append(requestErrors, drawingTimeInvalid.Error())
	}
	if wordSelectCountInvalid != nil {
		requestErrors = append(requestErrors, wordSelectCountInvalid.Error())
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
	if customWordsPerTurnInvalid != nil {
		requestErrors = append(requestErrors, customWordsPerTurnInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		requestErrors = append(requestErrors, clientsPerIPLimitInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		requestErrors = append(requestErrors, publicLobbyInvalid.Error())
	}
	if timerStartInvalid != nil {
		requestErrors = append(requestErrors, timerStartInvalid.Error())
	}

	if len(requestErrors) != 0 {
		http.Error(writer, strings.Join(requestErrors, ";"), http.StatusBadRequest)
		return
	}

	playerName := GetPlayername(request)
	player, lobby, err := game.CreateLobby(handler.cfg, playerName, language,
		publicLobby, timerStart, drawingTime, wordSelectCount, rounds, maxPlayers, customWordsPerTurn,
		clientsPerIPLimit, customWords)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	lobby.WriteObject = WriteObject
	lobby.WritePreparedMessage = WritePreparedMessage
	player.SetLastKnownAddress(GetIPAddressFromRequest(request))

	SetUsersessionCookie(writer, player)

	lobbyData := CreateLobbyData(lobby)

	if started, _, err := easyjson.MarshalToHTTPResponseWriter(lobbyData, writer); err != nil {
		if !started {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// We only add the lobby if everything else was successful.
	state.AddLobby(lobby)
}

func (handler *V1Handler) postPlayer(writer http.ResponseWriter, request *http.Request) {
	lobby := state.GetLobby(chi.URLParam(request, "lobby_id"))
	if lobby == nil {
		http.Error(writer, ErrLobbyNotExistent.Error(), http.StatusNotFound)
		return
	}

	var lobbyData *LobbyData

	lobby.Synchronized(func() {
		player := GetPlayer(lobby, request)

		if player == nil {
			if !lobby.HasFreePlayerSlot() {
				http.Error(writer, "lobby already full", http.StatusUnauthorized)
				return
			}

			requestAddress := GetIPAddressFromRequest(request)

			if !lobby.CanIPConnect(requestAddress) {
				http.Error(writer, "maximum amount of players per IP reached", http.StatusUnauthorized)
				return
			}

			newPlayer := lobby.JoinPlayer(GetPlayername(request))
			newPlayer.SetLastKnownAddress(requestAddress)

			// Use the players generated usersession and pass it as a cookie.
			SetUsersessionCookie(writer, newPlayer)
		} else {
			player.SetLastKnownAddress(GetIPAddressFromRequest(request))
		}

		lobbyData = CreateLobbyData(lobby)
	})

	if lobbyData != nil {
		if started, _, err := easyjson.MarshalToHTTPResponseWriter(lobbyData, writer); err != nil {
			if !started {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
}

// SetUsersessionCookie takes the players usersession and sets it as a cookie.
func SetUsersessionCookie(w http.ResponseWriter, player *game.Player) {
	session := player.GetUserSession().String()
	http.SetCookie(w, &http.Cookie{
		Name:     "usersession",
		Value:    session,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}

func (handler *V1Handler) patchLobby(writer http.ResponseWriter, request *http.Request) {
	userSession, err := GetUserSession(request)
	if err != nil {
		log.Printf("error getting user session: %v", err)
		http.Error(writer, "no valid usersession supplied", http.StatusBadRequest)
		return
	}

	if userSession == uuid.Nil {
		http.Error(writer, "no usersession supplied", http.StatusBadRequest)
		return
	}

	lobby := state.GetLobby(chi.URLParam(request, "lobby_id"))
	if lobby == nil {
		http.Error(writer, ErrLobbyNotExistent.Error(), http.StatusNotFound)
		return
	}

	if err := request.ParseForm(); err != nil {
		http.Error(writer, fmt.Sprintf("error parsing request query into form (%s)", err), http.StatusBadRequest)
		return
	}

	var requestErrors []string

	// Uneditable properties
	if request.Form.Get("custom_words") != "" {
		requestErrors = append(requestErrors, "can't modify custom_words in existing lobby")
	}
	if request.Form.Get("language") != "" {
		requestErrors = append(requestErrors, "can't modify language in existing lobby")
	}

	// Editable properties
	maxPlayers, maxPlayersInvalid := ParseMaxPlayers(request.Form.Get("max_players"))
	drawingTime, drawingTimeInvalid := ParseDrawingTime(request.Form.Get("drawing_time"))
	wordSelectCount, wordSelectCountInvalid := ParseWordSelectCount(request.Form.Get("word_select_count"))
	rounds, roundsInvalid := ParseRounds(request.Form.Get("rounds"))
	customWordsPerTurn, customWordsPerTurnInvalid := ParseCustomWordsPerTurn(request.Form.Get("custom_words_per_turn"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := ParseClientsPerIPLimit(request.Form.Get("clients_per_ip_limit"))
	publicLobby, publicLobbyInvalid := ParseBoolean("public", request.Form.Get("public"))
	timerStart, timerStartInvalid := ParseBoolean("timerStart", request.Form.Get("timer_start"))

	owner := lobby.Owner
	if owner == nil || owner.GetUserSession() != userSession {
		http.Error(writer, "only the lobby owner can edit the lobby", http.StatusForbidden)
		return
	}

	if maxPlayersInvalid != nil {
		requestErrors = append(requestErrors, maxPlayersInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		requestErrors = append(requestErrors, drawingTimeInvalid.Error())
	}
	if wordSelectCountInvalid != nil {
		requestErrors = append(requestErrors, wordSelectCountInvalid.Error())
	}
	if roundsInvalid != nil {
		requestErrors = append(requestErrors, roundsInvalid.Error())
	} else {
		currentRound := lobby.Round
		if rounds < currentRound {
			requestErrors = append(requestErrors, fmt.Sprintf("rounds must be greater than or equal to the current round (%d)", currentRound))
		}
	}
	if customWordsPerTurnInvalid != nil {
		requestErrors = append(requestErrors, customWordsPerTurnInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		requestErrors = append(requestErrors, clientsPerIPLimitInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		requestErrors = append(requestErrors, publicLobbyInvalid.Error())
	}
	if timerStartInvalid != nil {
		requestErrors = append(requestErrors, timerStartInvalid.Error())
	}

	if len(requestErrors) != 0 {
		http.Error(writer, strings.Join(requestErrors, ";"), http.StatusBadRequest)
		return
	}

	// We synchronize as late as possible to avoid unnecessary lags.
	// The previous code here isn't really prone to bugs due to lack of sync.
	lobby.Synchronized(func() {
		// While changing maxClientsPerIP and maxPlayers to a value lower than
		// is currently being used makes little sense, we'll allow it, as it doesn't
		// really break anything.

		lobby.MaxPlayers = maxPlayers
		lobby.CustomWordsPerTurn = customWordsPerTurn
		lobby.ClientsPerIPLimit = clientsPerIPLimit
		lobby.Public = publicLobby
		lobby.TimerStart = timerStart
		lobby.Rounds = rounds
		lobby.WordSelectCount = wordSelectCount

		if lobby.State == game.Ongoing {
			lobby.DrawingTimeNew = drawingTime
		} else {
			lobby.DrawingTime = drawingTime
		}
		lobbySettingsCopy := lobby.EditableLobbySettings
		lobbySettingsCopy.WordSelectCount = wordSelectCount
		lobbySettingsCopy.DrawingTime = drawingTime
		lobby.Broadcast(&game.Event{Type: game.EventTypeLobbySettingsChanged, Data: lobbySettingsCopy})
	})
}

func (handler *V1Handler) getStats(writer http.ResponseWriter, _ *http.Request) {
	if started, _, err := easyjson.MarshalToHTTPResponseWriter(state.Stats(), writer); err != nil {
		if !started {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

var (
	// CanvasColor is the initially / empty canvas colors value used for
	// Lobbydata objects.
	CanvasColor = game.RGBColor{R: 255, G: 255, B: 255}
	// SuggestedBrushSizes is suggested brush sizes value used for
	// Lobbydata objects. A unit test makes sure these values are ordered
	// and within the specified bounds.
	SuggestedBrushSizes = [4]uint8{8, 16, 24, 32}
)

// LobbyData is the data necessary for correctly configuring a lobby.
// While unofficial clients will probably need all of these values, the
// official webclient doesn't use all of them as of now.
type LobbyData struct {
	game.SettingBounds
	game.EditableLobbySettings

	LobbyID string `json:"lobbyId"`
	// DrawingBoardBaseWidth is the internal canvas width and is needed for
	// correctly up- / downscaling drawing instructions.
	DrawingBoardBaseWidth int `json:"drawingBoardBaseWidth"`
	// DrawingBoardBaseHeight is the internal canvas height and is needed for
	// correctly up- / downscaling drawing instructions.
	DrawingBoardBaseHeight int `json:"drawingBoardBaseHeight"`
	// MinBrushSize is the minimum amount of pixels the brush can draw in.
	MinBrushSize int `json:"minBrushSize"`
	// MaxBrushSize is the maximum amount of pixels the brush can draw in.
	MaxBrushSize int `json:"maxBrushSize"`
	// CanvasColor is the initially (empty) color of the canvas.
	CanvasColor game.RGBColor `json:"canvasColor"`
	// SuggestedBrushSizes are suggestions for the different brush sizes
	// that the user can choose between. These brushes are guaranteed to
	// be ordered from low to high and stay with the bounds.
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
func GetUserSession(request *http.Request) (uuid.UUID, error) {
	var userSession string
	if sessionCookie, err := request.Cookie("usersession"); err == nil && sessionCookie.Value != "" {
		userSession = sessionCookie.Value
	} else {
		userSession = request.Header.Get("Usersession")
	}

	if userSession == "" {
		return uuid.Nil, nil
	}

	id, err := uuid.FromString(userSession)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing user session: %w", err)
	}

	return id, nil
}

// GetPlayer returns the player object that matches the usersession in the
// supplied HTTP request and lobby. If no user session is set, we return nil.
func GetPlayer(lobby *game.Lobby, r *http.Request) *game.Player {
	userSession, err := GetUserSession(r)
	if err != nil {
		log.Printf("error getting user session: %v", err)
		return nil
	}

	if userSession == uuid.Nil {
		return nil
	}

	return lobby.GetPlayer(userSession)
}

// GetPlayername either retrieves the playername from a cookie, the URL form.
// If no preferred name can be found, we return an empty string.
func GetPlayername(request *http.Request) string {
	if err := request.ParseForm(); err == nil {
		username := request.Form.Get("username")
		if username != "" {
			return username
		}
	}

	if usernameCookie, err := request.Cookie("username"); err == nil {
		if usernameCookie.Value != "" {
			return usernameCookie.Value
		}
	}

	return ""
}
