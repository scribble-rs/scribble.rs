package frontend

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/api"
	"github.com/scribble-rs/scribble.rs/translations"
	"golang.org/x/text/language"
)

// GetPlayers returns divs for all players in the lobby to the calling client.
func GetPlayers(w http.ResponseWriter, r *http.Request) {
	lobby, err := api.GetLobby(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if api.GetPlayer(lobby, r) == nil {
		http.Error(w, "you aren't part of this lobby", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(lobby.GetPlayers())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type LobbyPageData struct {
	*api.LobbyData

	Translation translations.Translation
	Locale      string
}

// ssrEnterLobby opens a lobby, either opening it directly or asking for a lobby.
func ssrEnterLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := api.GetLobby(r)
	if err != nil {
		userFacingError(w, err.Error())
		return
	}

	userAgent := strings.ToLower(r.UserAgent())
	if !(strings.Contains(userAgent, "gecko") || strings.Contains(userAgent, "chrome") || strings.Contains(userAgent, "opera") || strings.Contains(userAgent, "safari")) {
		err := pageTemplates.ExecuteTemplate(w, "robot-page", api.CreateLobbyData(lobby))
		if err != nil {
			panic(err)
		}
		return
	}

	player := api.GetPlayer(lobby, r)

	if player == nil {
		if !lobby.HasFreePlayerSlot() {
			userFacingError(w, "Sorry, but the lobby is full.")
			return
		}

		var clientsWithSameIP int
		requestAddress := api.GetIPAddressFromRequest(r)
		for _, otherPlayer := range lobby.GetPlayers() {
			if otherPlayer.GetLastKnownAddress() == requestAddress {
				clientsWithSameIP++
				if clientsWithSameIP >= lobby.ClientsPerIPLimit {
					userFacingError(w, "Sorry, but you have exceeded the maximum number of clients per IP.")
					return
				}
			}
		}

		newPlayer := lobby.JoinPlayer(api.GetPlayername(r))

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
		player.SetLastKnownAddress(api.GetIPAddressFromRequest(r))
	}

	translation, locale := determineTranslation(r)

	pageData := &LobbyPageData{
		LobbyData:   api.CreateLobbyData(lobby),
		Translation: translation,
		Locale:      locale,
	}
	templateError := pageTemplates.ExecuteTemplate(w, "lobby-page", pageData)
	if templateError != nil {
		panic(templateError)
	}
}

func determineTranslation(r *http.Request) (translations.Translation, string) {
	var translation translations.Translation

	languageTags, _, languageParseError := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if languageParseError == nil && len(languageTags) > 0 {
		locale := languageTags[0]
		fullLocaleIdentifier := locale.String()
		fullLanguageIdentifier := strings.ToLower(fullLocaleIdentifier)
		translation = translations.GetLanguage(fullLanguageIdentifier)
		if translation != nil {
			return translation, fullLocaleIdentifier
		}

		baseLanguageIdentifier, _ := locale.Base()
		translation = translations.GetLanguage(strings.ToLower(baseLanguageIdentifier.String()))
		if translation != nil {
			return translation, baseLanguageIdentifier.String()
		}
	}

	return translations.DefaultTranslation, "en-US"
}
