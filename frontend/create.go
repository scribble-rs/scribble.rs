package frontend

import (
	"log"
	"net/http"

	"github.com/scribble-rs/scribble.rs/api"
	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/state"
	"github.com/scribble-rs/scribble.rs/translations"
)

//This file contains the API for the official web client.

// homePage servers the default page for scribble.rs, which is the page to
// create a new lobby.
func homePage(w http.ResponseWriter, r *http.Request) {
	translation, locale := determineTranslation(r)
	createPageData := createDefaultLobbyCreatePageData()
	createPageData.Translation = translation
	createPageData.Locale = locale

	err := pageTemplates.ExecuteTemplate(w, "lobby-create-page", createPageData)
	if err != nil {
		log.Printf("Error templating home page: %s\n", err)
	}
}

func createDefaultLobbyCreatePageData() *LobbyCreatePageData {
	return &LobbyCreatePageData{
		BasePageConfig:    currentBasePageConfig,
		SettingBounds:     game.LobbySettingBounds,
		Languages:         game.SupportedLanguages,
		Public:            "false",
		DrawingTime:       "120",
		Rounds:            "4",
		MaxPlayers:        "12",
		CustomWordsChance: "50",
		ClientsPerIPLimit: "1",
		EnableVotekick:    "true",
		Language:          "english",
	}
}

// LobbyCreatePageData defines all non-static data for the lobby create page.
type LobbyCreatePageData struct {
	*BasePageConfig
	*game.SettingBounds
	Translation       translations.Translation
	Locale            string
	Errors            []string
	Languages         map[string]string
	Public            string
	DrawingTime       string
	Rounds            string
	MaxPlayers        string
	CustomWords       string
	CustomWordsChance string
	ClientsPerIPLimit string
	EnableVotekick    string
	Language          string
}

// ssrCreateLobby allows creating a lobby, optionally returning errors that
// occurred during creation.
func ssrCreateLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		http.Error(w, formParseError.Error(), http.StatusBadRequest)
		return
	}

	language, languageInvalid := api.ParseLanguage(r.Form.Get("language"))
	drawingTime, drawingTimeInvalid := api.ParseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := api.ParseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := api.ParseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := api.ParseCustomWords(r.Form.Get("custom_words"))
	customWordChance, customWordChanceInvalid := api.ParseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := api.ParseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick, enableVotekickInvalid := api.ParseBoolean("enable votekick", r.Form.Get("enable_votekick"))
	publicLobby, publicLobbyInvalid := api.ParseBoolean("public", r.Form.Get("public"))

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := LobbyCreatePageData{
		BasePageConfig:    currentBasePageConfig,
		SettingBounds:     game.LobbySettingBounds,
		Languages:         game.SupportedLanguages,
		Public:            r.Form.Get("public"),
		DrawingTime:       r.Form.Get("drawing_time"),
		Rounds:            r.Form.Get("rounds"),
		MaxPlayers:        r.Form.Get("max_players"),
		CustomWords:       r.Form.Get("custom_words"),
		CustomWordsChance: r.Form.Get("custom_words_chance"),
		ClientsPerIPLimit: r.Form.Get("clients_per_ip_limit"),
		EnableVotekick:    r.Form.Get("enable_votekick"),
		Language:          r.Form.Get("language"),
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
	if customWordChanceInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		pageData.Errors = append(pageData.Errors, clientsPerIPLimitInvalid.Error())
	}
	if enableVotekickInvalid != nil {
		pageData.Errors = append(pageData.Errors, enableVotekickInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		pageData.Errors = append(pageData.Errors, publicLobbyInvalid.Error())
	}

	translation, locale := determineTranslation(r)
	pageData.Translation = translation
	pageData.Locale = locale

	if len(pageData.Errors) != 0 {
		err := pageTemplates.ExecuteTemplate(w, "lobby-create-page", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var playerName = api.GetPlayername(r)

	player, lobby, createError := game.CreateLobby(playerName, language, publicLobby, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit, customWords, enableVotekick)
	if createError != nil {
		pageData.Errors = append(pageData.Errors, createError.Error())
		templateError := pageTemplates.ExecuteTemplate(w, "lobby-create-page", pageData)
		if templateError != nil {
			userFacingError(w, templateError.Error())
		}

		return
	}

	lobby.WriteJSON = api.WriteJSON
	player.SetLastKnownAddress(api.GetIPAddressFromRequest(r))

	api.SetUsersessionCookie(w, player)

	//We only add the lobby if we could do all necessary pre-steps successfully.
	state.AddLobby(lobby)

	http.Redirect(w, r, currentBasePageConfig.RootPath+"/ssrEnterLobby?lobby_id="+lobby.LobbyID, http.StatusFound)
}
