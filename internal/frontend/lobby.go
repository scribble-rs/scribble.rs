package frontend

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/state"
	"github.com/scribble-rs/scribble.rs/internal/translations"
)

type lobbyPageData struct {
	*BasePageConfig
	*api.LobbyData

	Translation translations.Translation
	Locale      string
}

type robotPageData struct {
	*BasePageConfig
	*api.LobbyData
}

// ssrEnterLobby opens a lobby, either opening it directly or asking for a lobby.
func (handler *SSRHandler) ssrEnterLobby(writer http.ResponseWriter, request *http.Request) {
	lobby := state.GetLobby(chi.URLParam(request, "lobby_id"))
	if lobby == nil {
		handler.userFacingError(writer, api.ErrLobbyNotExistent.Error())
		return
	}

	userAgent := strings.ToLower(request.UserAgent())
	if !isHumanAgent(userAgent) {
		err := pageTemplates.ExecuteTemplate(writer, "robot-page", &robotPageData{
			BasePageConfig: handler.basePageConfig,
			LobbyData:      api.CreateLobbyData(handler.cfg, lobby),
		})
		if err != nil {
			log.Printf("error templating robot page: %d\n", err)
		}
		return
	}

	translation, locale := determineTranslation(request)
	requestAddress := api.GetIPAddressFromRequest(request)

	var pageData *lobbyPageData
	lobby.Synchronized(func() {
		player := api.GetPlayer(lobby, request)

		if player == nil {
			if !lobby.HasFreePlayerSlot() {
				handler.userFacingError(writer, "Sorry, but the lobby is full.")
				return
			}

			if !lobby.CanIPConnect(requestAddress) {
				handler.userFacingError(writer, "Sorry, but you have exceeded the maximum number of clients per IP.")
				return
			}

			newPlayer := lobby.JoinPlayer(api.GetPlayername(request))

			api.SetUsersessionCookie(writer, newPlayer)
		} else {
			if player.Connected && player.GetWebsocket() != nil {
				handler.userFacingError(writer, "It appears you already have an open tab for this lobby.")
				return
			}
			player.SetLastKnownAddress(requestAddress)
		}

		pageData = &lobbyPageData{
			BasePageConfig: handler.basePageConfig,
			LobbyData:      api.CreateLobbyData(handler.cfg, lobby),
			Translation:    translation,
			Locale:         locale,
		}
	})

	// If the pagedata isn't initialized, it means the synchronized block has exited.
	// In this case we don't want to template the lobby, since an error has occurred
	// and probably already has been handled.
	if pageData != nil {
		if err := pageTemplates.ExecuteTemplate(writer, "lobby-page", pageData); err != nil {
			log.Printf("Error templating lobby: %s\n", err)
		}
	}
}
