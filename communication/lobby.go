package communication

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/state"
)

var (
	errNoLobbyIDSupplied = errors.New("please supply a lobby id via the 'lobby_id' query parameter")
	errLobbyNotExistent  = errors.New("the requested lobby doesn't exist")
)

func getLobby(r *http.Request) (*game.Lobby, error) {
	lobbyID := r.URL.Query().Get("lobby_id")
	if lobbyID == "" {
		return nil, errNoLobbyIDSupplied
	}

	lobby := state.GetLobby(lobbyID)

	if lobby == nil {
		return nil, errLobbyNotExistent
	}

	return lobby, nil
}

func getUserSession(r *http.Request) string {
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

func getPlayer(lobby *game.Lobby, r *http.Request) *game.Player {
	return lobby.GetPlayer(getUserSession(r))
}

// getPlayername either retrieves the playername from a cookie, the URL form
// or generates a new random name if no name can be found.
func getPlayername(r *http.Request) string {
	parseError := r.ParseForm()
	if parseError == nil {
		username := html.EscapeString(strings.TrimSpace(r.Form.Get("username")))
		if username != "" {
			return trimDownTo(username, game.MaxPlayerNameLength)
		}
	}

	usernameCookie, noCookieError := r.Cookie("username")
	if noCookieError == nil {
		username := html.EscapeString(strings.TrimSpace(usernameCookie.Value))
		if username != "" {
			return trimDownTo(username, game.MaxPlayerNameLength)
		}
	}

	return game.GeneratePlayerName()
}

func trimDownTo(text string, size int) string {
	if len(text) <= size {
		return text
	}

	return text[:size]
}

// GetPlayers returns divs for all players in the lobby to the calling client.
func GetPlayers(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if getPlayer(lobby, r) == nil {
		http.Error(w, "you aren't part of this lobby", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(lobby.GetPlayers())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var (
	//CanvasColor is the initialy / empty canvas colors value used for
	//Lobbydata objects.
	CanvasColor = [3]uint8{255, 255, 255}
	//SuggestedBrushSizes is suggested brush sizes value used for
	//Lobbydata objects. A unit test makes sure these values are ordered
	//and within the specified bounds.
	SuggestedBrushSizes = [4]uint8{8, 16, 24, 32}
)

// LobbyData is the data necessary for correctly configuring a lobby.
// While unofficial clients will probably need all of these values, the
// official webclient doesn't use all of them as of now.
type LobbyData struct {
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
	//CanvasColor is the initial (empty) color of the canvas.
	//It's an array containing [R,G,B]
	CanvasColor [3]uint8 `json:"canvasColor"`
	//SuggestedBrushSizes are suggestions for the different brush sizes
	//that the user can choose between. These brushes are guaranted to
	//be ordered from low to high and stay with the bounds.
	SuggestedBrushSizes [4]uint8 `json:"suggestedBrushSizes"`
}

func createLobbyData(lobbyID string) *LobbyData {
	return &LobbyData{
		LobbyID:                lobbyID,
		DrawingBoardBaseWidth:  game.DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: game.DrawingBoardBaseHeight,
		MinBrushSize:           game.MinBrushSize,
		MaxBrushSize:           game.MaxBrushSize,
		CanvasColor:            CanvasColor,
		SuggestedBrushSizes:    SuggestedBrushSizes,
	}
}

// ssrEnterLobby opens a lobby, either opening it directly or asking for a lobby.
func ssrEnterLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		userFacingError(w, err.Error())
		return
	}

	// TODO Improve this. Return metadata or so instead.
	userAgent := strings.ToLower(r.UserAgent())
	if !(strings.Contains(userAgent, "gecko") || strings.Contains(userAgent, "chrome") || strings.Contains(userAgent, "opera") || strings.Contains(userAgent, "safari")) {
		userFacingError(w, "Sorry, no robots allowed.")
		return
	}

	//FIXME Temporary
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "android") {
		userFacingError(w, "Sorry, mobile is currently not supported.")
		return
	}

	player := getPlayer(lobby, r)

	pageData := createLobbyData(lobby.ID)

	if player == nil {
		if !lobby.HasFreePlayerSlot() {
			userFacingError(w, "Sorry, but the lobby is full.")
			return
		}

		var clientsWithSameIP int
		requestAddress := getIPAddressFromRequest(r)
		for _, otherPlayer := range lobby.GetPlayers() {
			if otherPlayer.GetLastKnownAddress() == requestAddress {
				clientsWithSameIP++
				if clientsWithSameIP >= lobby.ClientsPerIPLimit {
					userFacingError(w, "Sorry, but you have exceeded the maximum number of clients per IP.")
					return
				}
			}
		}

		newPlayer := lobby.JoinPlayer(getPlayername(r))

		// Use the players generated usersession and pass it as a cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "usersession",
			Value:    newPlayer.GetUserSession(),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		if player.Connected && player.GetWebsocket() != nil {
			userFacingError(w, "It appears you already have an open tab for this lobby.")
			return
		}
		player.SetLastKnownAddress(getIPAddressFromRequest(r))
	}

	templateError := pageTemplates.ExecuteTemplate(w, "lobby-page", pageData)
	if templateError != nil {
		panic(templateError)
	}
}

func getIPAddressFromRequest(r *http.Request) string {
	//See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	//See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded

	//The following logic has been implemented according to the spec, therefore please
	//refer to the spec if you have a question.

	forwardedAddress := r.Header.Get("X-Forwarded-For")
	if forwardedAddress != "" {
		//Since the field may contain multiple addresses separated by commas, we use the first
		//one, which according to the docs is supposed to be the client address.
		clientAddress := strings.TrimSpace(strings.Split(forwardedAddress, ",")[0])
		return remoteAddressToSimpleIP(clientAddress)
	}

	standardForwardedHeader := r.Header.Get("Forwarded")
	if standardForwardedHeader != "" {
		targetPrefix := "for="
		//Since forwarded can contain more than one field, we search for one specific field.
		for _, part := range strings.Split(standardForwardedHeader, ";") {
			trimmed := strings.TrimSpace(part)
			if strings.HasPrefix(trimmed, targetPrefix) {
				//FIXME Maybe checking for a valid IP-Address would make sense here, not sure tho.
				address := remoteAddressToSimpleIP(strings.TrimPrefix(trimmed, targetPrefix))
				//Since the documentation doesn't mention which quotes are used, I just remove all ;)
				return strings.NewReplacer("`", "", "'", "", "\"", "", "[", "", "]", "").Replace(address)
			}
		}
	}

	return remoteAddressToSimpleIP(r.RemoteAddr)
}
