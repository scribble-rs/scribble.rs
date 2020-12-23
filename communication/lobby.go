package communication

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

var (
	noLobbyIdSuppliedError = errors.New("please supply a lobby id via the 'lobby_id' query parameter")
	lobbyNotExistentError  = errors.New("the requested lobby doesn't exist")
)

func getLobby(r *http.Request) (*game.Lobby, error) {
	lobbyID := r.URL.Query().Get("lobby_id")
	if lobbyID == "" {
		return nil, noLobbyIdSuppliedError
	}

	lobby := game.GetLobby(lobbyID)

	if lobby == nil {
		return nil, lobbyNotExistentError
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

func getPlayername(r *http.Request) string {
	usernameCookie, noCookieError := r.Cookie("username")
	if noCookieError == nil {
		username := html.EscapeString(strings.TrimSpace(usernameCookie.Value))
		if username != "" {
			return trimDownTo(username, 30)
		}
	}

	parseError := r.ParseForm()
	if parseError == nil {
		username := r.Form.Get("username")
		if username != "" {
			return trimDownTo(username, 30)
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
	err = json.NewEncoder(w).Encode(lobby.Players)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//GetRounds returns the html structure and data for the current round info.
func GetRounds(w http.ResponseWriter, r *http.Request) {
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
	err = json.NewEncoder(w).Encode(game.Rounds{Round: lobby.Round, MaxRounds: lobby.MaxRounds})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetWordHint returns the html structure and data for the current word hint.
func GetWordHint(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	player := getPlayer(lobby, r)
	if player == nil {
		http.Error(w, "you aren't part of this lobby", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(lobby.GetAvailableWordHints(player))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const (
	DrawingBoardBaseWidth  = 1600
	DrawingBoardBaseHeight = 900
)

// LobbyData is the data necessary for initially displaying all data of
// the lobbies webpage.
type LobbyData struct {
	LobbyID                string `json:"lobbyId"`
	DrawingBoardBaseWidth  int    `json:"drawingBoardBaseWidth"`
	DrawingBoardBaseHeight int    `json:"drawingBoardBaseHeight"`
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

	pageData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	var templateError error

	if player == nil {
		if len(lobby.Players) >= lobby.MaxPlayers {
			userFacingError(w, "Sorry, but the lobby is full.")
			return
		}

		var clientsWithSameIP int
		requestAddress := getIPAddressFromRequest(r)
		for _, otherPlayer := range lobby.Players {
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

	templateError = lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
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
