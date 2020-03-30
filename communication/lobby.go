package communication

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

func getLobby(w http.ResponseWriter, r *http.Request) *game.Lobby {
	//FIXME We might wanna check the usersession here.

	lobbyID := r.URL.Query().Get("id")
	if lobbyID == "" {
		returnError(w, "The entered URL is incorrect.")
		return nil
	}

	lobby := game.GetLobby(lobbyID)

	if lobby == nil {
		returnError(w, "The lobby does not exist.")
	}

	return lobby
}

// GetPlayers returns divs for all players in the lobby to the calling client.
func GetPlayers(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		templatingError := lobbyPage.ExecuteTemplate(w, "players", lobby)
		if templatingError != nil {
			returnError(w, templatingError.Error())
		}
	}
}

//GetRounds returns the html structure and data for the current round info.
func GetRounds(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		fmt.Fprintf(w, "Round %d of %d", lobby.Round, lobby.Rounds)
	}
}

// GetWordHint returns the html structure and data for the current word hint.
func GetWordHint(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		sessionCookie, noCookieError := r.Cookie("usersession")
		if noCookieError != nil {
			errorPage.ExecuteTemplate(w, "error.html", "You aren't part of this lobby.")
			return
		}

		player := lobby.GetPlayer(sessionCookie.Value)
		if player == nil {
			errorPage.ExecuteTemplate(w, "error.html", "You aren't part of this lobby.")
			return
		}

		wordHints := lobby.GetAvailableWordHints(player)

		templatingError := lobbyPage.ExecuteTemplate(w, "word", wordHints)
		if templatingError != nil {
			errorPage.ExecuteTemplate(w, "error.html", templatingError.Error())
		}
	}
}

// ShowLobby opens a lobby, either opening it directly or asking for a lobby.
func ShowLobby(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		// TODO Improve this. Return metadata or so instead.
		userAgent := strings.ToLower(r.UserAgent())
		if !(strings.Contains(userAgent, "gecko") || strings.Contains(userAgent, "chrom") || strings.Contains(userAgent, "opera") || strings.Contains(userAgent, "safari")) {
			returnError(w, "Sorry, no robots allowed.")
			return
		}

		//FIXME Temporary
		if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "android") {
			returnError(w, "Sorry, mobile is currently not supported.")
			return
		}

		sessionCookie, noCookieError := r.Cookie("usersession")
		var player *game.Player
		if noCookieError == nil {
			player = lobby.GetPlayer(sessionCookie.Value)
		}

		//Potentially unused garbage, but we'll take it.
		pageData := &game.LobbyPageData{
			Players:        lobby.Players,
			LobbyID:        lobby.ID,
			Round:          lobby.Round,
			Rounds:         lobby.Rounds,
			EnableVotekick: lobby.EnableVotekick,
		}

		if player == nil {
			if len(lobby.Players) >= lobby.MaxPlayers {
				returnError(w, "Sorry, but the lobby is full.")
				return
			}

			matches := 0
			for _, otherPlayer := range lobby.Players {
				socket := otherPlayer.Ws
				if socket != nil && remoteAddressToSimpleIP(socket.RemoteAddr().String()) == remoteAddressToSimpleIP(r.RemoteAddr) {
					matches++
				}
			}

			if matches >= lobby.ClientsPerIPLimit {
				errorPage.ExecuteTemplate(w, "error.html", "Sorry, but you have exceeded the maximum number of clients per IP.")
				return
			}

			var playerName string
			usernameCookie, noCookieError := r.Cookie("username")
			if noCookieError == nil {
				playerName = usernameCookie.Value
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

			lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		} else {
			lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		}
	}
}
