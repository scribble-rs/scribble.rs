package frontend

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/state"
	"github.com/scribble-rs/scribble.rs/internal/translations"
)

// This file contains the API for the official web client.

type SSRHandler struct {
	cfg            *config.Config
	basePageConfig *BasePageConfig
}

func NewHandler(cfg *config.Config) (*SSRHandler, error) {
	basePageConfig := &BasePageConfig{}
	if cfg.RootPath != "" {
		basePageConfig.RootPath = "/" + cfg.RootPath
	}

	var err error
	pageTemplates, err = template.ParseFS(templateFS, "templates/*")
	if err != nil {
		return nil, fmt.Errorf("error loading templates: %w", err)
	}

	entries, err := frontendResourcesFS.ReadDir("resources")
	if err != nil {
		return nil, fmt.Errorf("error reading resource directory: %w", err)
	}

	hash := md5.New()
	for _, entry := range entries {
		bytes, err := frontendResourcesFS.ReadFile("resources/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("error reading resource %s: %w", entry.Name(), err)
		}

		if _, err := hash.Write(bytes); err != nil {
			return nil, fmt.Errorf("error hashing resource %s: %w", entry.Name(), err)
		}
	}

	basePageConfig.CacheBust = fmt.Sprintf("%x", hash.Sum(nil))

	handler := &SSRHandler{
		cfg:            cfg,
		basePageConfig: basePageConfig,
	}
	return handler, nil
}

// homePage servers the default page for scribble.rs, which is the page to
// create a new lobby.
func (handler *SSRHandler) homePageHandler(writer http.ResponseWriter, request *http.Request) {
	translation, locale := determineTranslation(request)
	createPageData := handler.createDefaultLobbyCreatePageData()
	createPageData.Translation = translation
	createPageData.Locale = locale

	err := pageTemplates.ExecuteTemplate(writer, "lobby-create-page", createPageData)
	if err != nil {
		log.Printf("Error templating home page: %s\n", err)
	}
}

func (handler *SSRHandler) createDefaultLobbyCreatePageData() *LobbyCreatePageData {
	return &LobbyCreatePageData{
		BasePageConfig:       handler.basePageConfig,
		SettingBounds:        game.LobbySettingBounds,
		Languages:            game.SupportedLanguages,
		LobbySettingDefaults: handler.cfg.LobbySettingDefaults,
	}
}

// LobbyCreatePageData defines all non-static data for the lobby create page.
type LobbyCreatePageData struct {
	*BasePageConfig
	config.LobbySettingDefaults
	game.SettingBounds

	Translation translations.Translation
	Locale      string
	Errors      []string
	Languages   map[string]string
}

// ssrCreateLobby allows creating a lobby, optionally returning errors that
// occurred during creation.
func (handler *SSRHandler) ssrCreateLobby(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	language, languageInvalid := api.ParseLanguage(request.Form.Get("language"))
	drawingTime, drawingTimeInvalid := api.ParseDrawingTime(request.Form.Get("drawing_time"))
	wordSelectCount, wordSelectCountInvalid := api.ParseWordSelectCount(request.Form.Get("word_select_count"))
	rounds, roundsInvalid := api.ParseRounds(request.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := api.ParseMaxPlayers(request.Form.Get("max_players"))
	customWords, customWordsInvalid := api.ParseCustomWords(request.Form.Get("custom_words"))
	customWordsPerTurn, customWordsPerTurnInvalid := api.ParseCustomWordsPerTurn(request.Form.Get("custom_words_per_turn"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := api.ParseClientsPerIPLimit(request.Form.Get("clients_per_ip_limit"))
	publicLobby, publicLobbyInvalid := api.ParseBoolean("public", request.Form.Get("public"))
	timerStart, timerStartInvalid := api.ParseBoolean("timerStart", request.Form.Get("timer_start"))

	// Prevent resetting the form, since that would be annoying as hell.
	pageData := LobbyCreatePageData{
		BasePageConfig: handler.basePageConfig,
		SettingBounds:  game.LobbySettingBounds,
		LobbySettingDefaults: config.LobbySettingDefaults{
			Public:             request.Form.Get("public"),
			TimerStart:         request.Form.Get("timer_start"),
			DrawingTime:        request.Form.Get("drawing_time"),
			WordSelectCount:    request.Form.Get("word_select_count"),
			Rounds:             request.Form.Get("rounds"),
			MaxPlayers:         request.Form.Get("max_players"),
			CustomWords:        request.Form.Get("custom_words"),
			CustomWordsPerTurn: request.Form.Get("custom_words_per_turn"),
			ClientsPerIPLimit:  request.Form.Get("clients_per_ip_limit"),
			Language:           request.Form.Get("language"),
		},
	}

	if languageInvalid != nil {
		pageData.Errors = append(pageData.Errors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		pageData.Errors = append(pageData.Errors, drawingTimeInvalid.Error())
	}
	if wordSelectCountInvalid != nil {
		pageData.Errors = append(pageData.Errors, wordSelectCountInvalid.Error())
	}
	if roundsInvalid != nil {
		pageData.Errors = append(pageData.Errors, roundsInvalid.Error())
	}
	if maxPlayersInvalid != nil {
		pageData.Errors = append(pageData.Errors, maxPlayersInvalid.Error())
	}
	if customWordsInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordsInvalid.Error())
	}
	if customWordsPerTurnInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordsPerTurnInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		pageData.Errors = append(pageData.Errors, clientsPerIPLimitInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		pageData.Errors = append(pageData.Errors, publicLobbyInvalid.Error())
	}
	if timerStartInvalid != nil {
		pageData.Errors = append(pageData.Errors, timerStartInvalid.Error())
	}

	translation, locale := determineTranslation(request)
	pageData.Translation = translation
	pageData.Locale = locale

	if len(pageData.Errors) != 0 {
		err := pageTemplates.ExecuteTemplate(writer, "lobby-create-page", pageData)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	playerName := api.GetPlayername(request)

	player, lobby, err := game.CreateLobby(handler.cfg, playerName, language,
		publicLobby, timerStart, drawingTime, wordSelectCount, rounds, maxPlayers, customWordsPerTurn,
		clientsPerIPLimit, customWords)
	if err != nil {
		pageData.Errors = append(pageData.Errors, err.Error())
		if err := pageTemplates.ExecuteTemplate(writer, "lobby-create-page", pageData); err != nil {
			handler.userFacingError(writer, err.Error())
		}

		return
	}

	lobby.WriteObject = api.WriteObject
	lobby.WritePreparedMessage = api.WritePreparedMessage
	player.SetLastKnownAddress(api.GetIPAddressFromRequest(request))

	api.SetUsersessionCookie(writer, player)

	// We only add the lobby if we could do all necessary pre-steps successfully.
	state.AddLobby(lobby)

	http.Redirect(writer, request, handler.basePageConfig.RootPath+"/ssrEnterLobby/"+lobby.LobbyID, http.StatusFound)
}
