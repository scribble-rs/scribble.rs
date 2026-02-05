package frontend

import (
	"log"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/internal/translations"
)

type galleryPageData struct {
	*BasePageConfig

	LobbyID     string
	Translation *translations.Translation
	Locale      string
}

func (handler *SSRHandler) ssrGallery(writer http.ResponseWriter, request *http.Request) {
	userAgent := strings.ToLower(request.UserAgent())
	if !isHumanAgent(userAgent) {
		// FIXME Handle robots
		return
	}

	lobbyId := request.PathValue("lobby_id")
	translation, locale := determineTranslation(request)
	pageData := &galleryPageData{
		BasePageConfig: handler.basePageConfig,
		LobbyID:        lobbyId,
		Translation:    translation,
		Locale:         locale,
	}

	// If the pagedata isn't initialized, it means the synchronized block has exited.
	// In this case we don't want to template the lobby, since an error has occurred
	// and probably already has been handled.
	if err := pageTemplates.ExecuteTemplate(writer, "gallery-page", pageData); err != nil {
		log.Printf("Error templating lobby: %s\n", err)
	}
}
