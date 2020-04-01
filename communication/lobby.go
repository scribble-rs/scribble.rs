package communication

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

func getLobby(r *http.Request) (*game.Lobby, error) {
	lobbyID := r.URL.Query().Get("id")
	if lobbyID == "" {
		return nil, errors.New("the requested lobby doesn't exist")
	}

	lobby := game.GetLobby(lobbyID)

	if lobby == nil {
		return nil, errors.New("the requested lobby doesn't exist")
	}

	return lobby, nil
}

func getPlayer(lobby *game.Lobby, r *http.Request) *game.Player {
	sessionCookie, noCookieError := r.Cookie("usersession")
	var player *game.Player
	if noCookieError == nil {
		player = lobby.GetPlayer(sessionCookie.Value)
	}

	return player
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

// ShowLobby opens a lobby, either opening it directly or asking for a lobby.
func ShowLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		userFacingError(w, err.Error())
	} else {
		// TODO Improve this. Return metadata or so instead.
		userAgent := strings.ToLower(r.UserAgent())
		if !(strings.Contains(userAgent, "gecko") || strings.Contains(userAgent, "chrom") || strings.Contains(userAgent, "opera") || strings.Contains(userAgent, "safari")) {
			userFacingError(w, "Sorry, no robots allowed.")
			return
		}

		//FIXME Temporary
		if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "android") {
			userFacingError(w, "Sorry, mobile is currently not supported.")
			return
		}

		player := getPlayer(lobby, r)

		//Potentially unused garbage, but we'll take it.
		pageData := &game.LobbyPageData{
			Players:        lobby.Players,
			LobbyID:        lobby.ID,
			Round:          lobby.Round,
			Rounds:         lobby.MaxRounds,
			EnableVotekick: lobby.EnableVotekick,
		}

		var templateError error

		if player == nil {
			if len(lobby.Players) >= lobby.MaxPlayers {
				userFacingError(w, "Sorry, but the lobby is full.")
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
				userFacingError(w, "Sorry, but you have exceeded the maximum number of clients per IP.")
				return
			}

			var playerName string
			usernameCookie, noCookieError := r.Cookie("username")
			if noCookieError == nil {
				playerName = html.EscapeString(usernameCookie.Value)
			} else {
				playerName = game.GeneratePlayerName()
			}

			userSession := lobby.JoinPlayer(playerName)

			// Use the players generated usersession and pass it as a cookie.
			http.SetCookie(w, &http.Cookie{
				Name:     "usersession",
				Value:    userSession,
				Path:     "/",
				SameSite: http.SameSiteStrictMode,
			})

			templateError = lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		} else {
			templateError = lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		}

		if templateError != nil {
			panic(templateError)
		}
	}
}
