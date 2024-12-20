package frontend

import (
	//nolint:gosec //We just use this for cache busting, so it's secure enough
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/state"
	"github.com/scribble-rs/scribble.rs/internal/translations"
	"github.com/scribble-rs/scribble.rs/internal/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// This file contains the API for the official web client.

type SSRHandler struct {
	cfg            *config.Config
	basePageConfig *BasePageConfig
}

func NewHandler(cfg *config.Config) (*SSRHandler, error) {
	basePageConfig := &BasePageConfig{
		Version: version.Version,
		Commit:  version.Commit,
		RootURL: cfg.RootURL,
	}
	if cfg.RootPath != "" {
		basePageConfig.RootPath = "/" + cfg.RootPath
	}

	entries, err := frontendResourcesFS.ReadDir("resources")
	if err != nil {
		return nil, fmt.Errorf("error reading resource directory: %w", err)
	}

	//nolint:gosec //We just use this for cache busting, so it's secure enough
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

	basePageConfig.CacheBust = hex.EncodeToString(hash.Sum(nil))

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
		SettingBounds:        handler.cfg.LobbySettingBounds,
		Languages:            game.SupportedLanguages,
		ScoreCalculations:    game.SupportedScoreCalculations,
		LobbySettingDefaults: handler.cfg.LobbySettingDefaults,
	}
}

// LobbyCreatePageData defines all non-static data for the lobby create page.
type LobbyCreatePageData struct {
	*BasePageConfig
	config.LobbySettingDefaults
	game.SettingBounds

	Translation       translations.Translation
	Locale            string
	Errors            []string
	Languages         map[string]string
	ScoreCalculations []string
}

// ssrCreateLobby allows creating a lobby, optionally returning errors that
// occurred during creation.
func (handler *SSRHandler) ssrCreateLobby(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	scoreCalculation, scoreCalculationInvalid := api.ParseScoreCalculation(request.Form.Get("score_calculation"))
	languageData, languageKey, languageInvalid := api.ParseLanguage(request.Form.Get("language"))
	drawingTime, drawingTimeInvalid := api.ParseDrawingTime(handler.cfg, request.Form.Get("drawing_time"))
	rounds, roundsInvalid := api.ParseRounds(handler.cfg, request.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := api.ParseMaxPlayers(handler.cfg, request.Form.Get("max_players"))
	customWordsPerTurn, customWordsPerTurnInvalid := api.ParseCustomWordsPerTurn(request.Form.Get("custom_words_per_turn"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := api.ParseClientsPerIPLimit(handler.cfg, request.Form.Get("clients_per_ip_limit"))
	publicLobby, publicLobbyInvalid := api.ParseBoolean("public", request.Form.Get("public"))

	var lowercaser cases.Caser
	if languageInvalid != nil {
		lowercaser = cases.Lower(language.English)
	} else {
		lowercaser = languageData.Lowercaser()
	}
	customWords, customWordsInvalid := api.ParseCustomWords(lowercaser, request.Form.Get("custom_words"))

	// Prevent resetting the form, since that would be annoying as hell.
	pageData := LobbyCreatePageData{
		BasePageConfig: handler.basePageConfig,
		SettingBounds:  handler.cfg.LobbySettingBounds,
		LobbySettingDefaults: config.LobbySettingDefaults{
			Public:             request.Form.Get("public"),
			DrawingTime:        request.Form.Get("drawing_time"),
			Rounds:             request.Form.Get("rounds"),
			MaxPlayers:         request.Form.Get("max_players"),
			CustomWords:        request.Form.Get("custom_words"),
			CustomWordsPerTurn: request.Form.Get("custom_words_per_turn"),
			ClientsPerIPLimit:  request.Form.Get("clients_per_ip_limit"),
			Language:           request.Form.Get("language"),
			ScoreCalculation:   request.Form.Get("score_calculation"),
		},
		Languages: game.SupportedLanguages,
	}

	if scoreCalculationInvalid != nil {
		pageData.Errors = append(pageData.Errors, scoreCalculationInvalid.Error())
	}
	if languageInvalid != nil {
		pageData.Errors = append(pageData.Errors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		pageData.Errors = append(pageData.Errors, drawingTimeInvalid.Error())
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

	player, lobby, err := game.CreateLobby(uuid.Nil, playerName, languageKey,
		publicLobby, drawingTime, rounds, maxPlayers, customWordsPerTurn,
		clientsPerIPLimit, customWords, scoreCalculation)
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
