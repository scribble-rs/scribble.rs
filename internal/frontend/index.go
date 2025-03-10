package frontend

import (
	//nolint:gosec //We just use this for cache busting, so it's secure enough

	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	txtTemplate "text/template"

	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/state"
	"github.com/scribble-rs/scribble.rs/internal/translations"
	"github.com/scribble-rs/scribble.rs/internal/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	_ "embed"
)

//go:embed lobby.js
var lobbyJsRaw string

//go:embed lobby-discord.js
var lobbyDiscordJsRaw string

func init() {
	lobbyJsRaw = `
{{if .DiscordActivity}}
` + lobbyDiscordJsRaw + `
{{end}}
` + lobbyJsRaw
	lobbyDiscordJsRaw = ""
}

//go:embed index.js
var indexJsRaw string

type indexJsData struct {
	*BasePageConfig

	Translation translations.Translation
	Locale      string
}

// This file contains the API for the official web client.

type SSRHandler struct {
	cfg                *config.Config
	basePageConfig     *BasePageConfig
	lobbyJsRawTemplate *txtTemplate.Template
	indexJsRawTemplate *txtTemplate.Template
}

func NewHandler(cfg *config.Config) (*SSRHandler, error) {
	basePageConfig := &BasePageConfig{
		checksums:     make(map[string]string),
		hash:          md5.New(),
		Version:       version.Version,
		Commit:        version.Commit,
		RootURL:       cfg.RootURL,
		CanonicalURL:  cfg.CanonicalURL,
		AllowIndexing: cfg.AllowIndexing,
	}
	if cfg.RootPath != "" {
		basePageConfig.RootPath = "/" + cfg.RootPath
	}
	if basePageConfig.CanonicalURL == "" {
		basePageConfig.CanonicalURL = basePageConfig.RootURL
	}

	indexJsRawTemplate, err := txtTemplate.
		New("index-js").
		Parse(indexJsRaw)
	if err != nil {
		return nil, fmt.Errorf("error parsing index js template: %w", err)
	}

	lobbyJsRawTemplate, err := txtTemplate.
		New("lobby-js").
		Parse(lobbyJsRaw)
	if err != nil {
		return nil, fmt.Errorf("error parsing lobby js template: %w", err)
	}

	lobbyJsRawTemplate.AddParseTree("footer", pageTemplates.Tree)

	entries, err := frontendResourcesFS.ReadDir("resources")
	if err != nil {
		return nil, fmt.Errorf("error reading resource directory: %w", err)
	}

	//nolint:gosec //We just use this for cache busting, so it's secure enough
	for _, entry := range entries {
		bytes, err := frontendResourcesFS.ReadFile("resources/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("error reading resource %s: %w", entry.Name(), err)
		}

		if err := basePageConfig.Hash(entry.Name(), bytes); err != nil {
			return nil, fmt.Errorf("error hashing resource %s: %w", entry.Name(), err)
		}
	}
	if err := basePageConfig.Hash("index.js", []byte(indexJsRaw)); err != nil {
		return nil, fmt.Errorf("error hashing: %w", err)
	}
	if err := basePageConfig.Hash("lobby.js", []byte(lobbyJsRaw), []byte(lobbyDiscordJsRaw)); err != nil {
		return nil, fmt.Errorf("error hashing: %w", err)
	}

	handler := &SSRHandler{
		cfg:                cfg,
		basePageConfig:     basePageConfig,
		lobbyJsRawTemplate: lobbyJsRawTemplate,
		indexJsRawTemplate: indexJsRawTemplate,
	}
	return handler, nil
}

func (handler *SSRHandler) indexJs(writer http.ResponseWriter, request *http.Request) {
	translation, locale := determineTranslation(request)
	pageData := &indexJsData{
		BasePageConfig: handler.basePageConfig,
		Translation:    translation,
		Locale:         locale,
	}

	writer.Header().Set("Content-Type", "text/javascript")
	// Duration of 1 year, since we use cachebusting anyway.
	writer.Header().Set("Cache-Control", "public, max-age=31536000")
	writer.WriteHeader(http.StatusOK)
	if err := handler.indexJsRawTemplate.ExecuteTemplate(writer, "index-js", pageData); err != nil {
		log.Printf("error templating JS: %s\n", err)
	}
}

// indexPageHandler servers the default page for scribble.rs, which is the
// page to create or join a lobby.
func (handler *SSRHandler) indexPageHandler(writer http.ResponseWriter, request *http.Request) {
	translation, locale := determineTranslation(request)
	createPageData := handler.createDefaultIndexPageData()
	createPageData.Translation = translation
	createPageData.Locale = locale

	api.SetDiscordCookies(writer, request)
	discordInstanceId := api.GetDiscordInstanceId(request)
	if discordInstanceId != "" {
		lobby := state.GetLobby(discordInstanceId)
		if lobby != nil {
			handler.ssrEnterLobbyNoChecks(lobby, writer, request,
				func() *game.Player {
					return api.GetPlayer(lobby, request)
				})
			return
		}

		createPageData.DiscordActivity = true
	}

	err := pageTemplates.ExecuteTemplate(writer, "index", createPageData)
	if err != nil {
		log.Printf("Error templating home page: %s\n", err)
	}
}

func (handler *SSRHandler) createDefaultIndexPageData() *IndexPageData {
	return &IndexPageData{
		BasePageConfig:       handler.basePageConfig,
		SettingBounds:        handler.cfg.LobbySettingBounds,
		Languages:            game.SupportedLanguages,
		ScoreCalculations:    game.SupportedScoreCalculations,
		LobbySettingDefaults: handler.cfg.LobbySettingDefaults,
	}
}

// IndexPageData defines all non-static data for the lobby create page.
type IndexPageData struct {
	*BasePageConfig
	config.LobbySettingDefaults
	game.SettingBounds

	DiscordActivity   bool
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

	api.SetDiscordCookies(writer, request)
	// Prevent resetting the form, since that would be annoying as hell.
	pageData := IndexPageData{
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
		err := pageTemplates.ExecuteTemplate(writer, "index", pageData)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	playerName := api.GetPlayername(request)

	var lobbyId string
	discordInstanceId := api.GetDiscordInstanceId(request)
	if discordInstanceId != "" {
		lobbyId = discordInstanceId
		// Workaround, since the discord proxy potentially always has the same
		// IP address, preventing all players from connecting.
		clientsPerIPLimit = maxPlayers
	}

	player, lobby, err := game.CreateLobby(lobbyId, playerName, languageKey,
		publicLobby, drawingTime, rounds, maxPlayers, customWordsPerTurn,
		clientsPerIPLimit, customWords, scoreCalculation)
	if err != nil {
		pageData.Errors = append(pageData.Errors, err.Error())
		if err := pageTemplates.ExecuteTemplate(writer, "index", pageData); err != nil {
			handler.userFacingError(writer, err.Error())
		}

		return
	}

	lobby.WriteObject = api.WriteObject
	lobby.WritePreparedMessage = api.WritePreparedMessage
	player.SetLastKnownAddress(api.GetIPAddressFromRequest(request))
	api.SetGameplayCookies(writer, request, player, lobby)

	// We only add the lobby if we could do all necessary pre-steps successfully.
	state.AddLobby(lobby)

	// Workaround for discord activity case not correctly being able to read
	// user session, as the cookie isn't being passed.
	if discordInstanceId != "" {
		handler.ssrEnterLobbyNoChecks(lobby, writer, request,
			func() *game.Player {
				return player
			})
		return
	}

	http.Redirect(writer, request, handler.basePageConfig.RootPath+"/ssrEnterLobby/"+lobby.LobbyID, http.StatusFound)
}
