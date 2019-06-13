package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var lobbyCreatePage *template.Template
var lobbySettingBounds = &LobbySettingBounds{
	MinDrawingTime: 60,
	MaxDrawingTime: 300,
	MinRounds:      1,
	MaxRounds:      20,
	MinMaxPlayers:  2,
	MaxMaxPlayers:  24,
}

// LobbySettingBounds defines the lower and upper bounds for the user-specified
// lobby creation input.
type LobbySettingBounds struct {
	MinDrawingTime int64
	MaxDrawingTime int64
	MinRounds      int64
	MaxRounds      int64
	MinMaxPlayers  int64
	MaxMaxPlayers  int64
}

// LobbyCreatePageData defines all non-static data for the lobby create page.
type LobbyCreatePageData struct {
	*LobbySettingBounds
	Errors      []string
	Name        string
	Password    string
	DrawingTime string
	Rounds      string
	MaxPlayers  string
	CustomWords string
}

func createDefaultLobbyCreatePageDat() *LobbyCreatePageData {
	return &LobbyCreatePageData{
		LobbySettingBounds: lobbySettingBounds,
		DrawingTime:        "120",
		Rounds:             "4",
		MaxPlayers:         "4",
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	err := lobbyCreatePage.ExecuteTemplate(w, "lobby.html", createDefaultLobbyCreatePageDat())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		panic(formParseError)
	}

	playerName, playerNameInvalid := parsePlayerName(r.Form.Get("player_name"))
	lobbyPassword, lobbyPasswordInvalid := parseLobbyPassword(r.Form.Get("lobby_password"))
	drawingTime, drawingTimeInvalid := parseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := parseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := parseCustomWords(r.Form.Get("custom_words"))

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := LobbyCreatePageData{
		LobbySettingBounds: lobbySettingBounds,
		Name:               r.Form.Get("player_name"),
		Password:           r.Form.Get("lobby_password"),
		DrawingTime:        r.Form.Get("drawing_time"),
		Rounds:             r.Form.Get("rounds"),
		MaxPlayers:         r.Form.Get("max_players"),
		CustomWords:        r.Form.Get("custom_words"),
	}

	if lobbyPasswordInvalid != nil {
		pageData.Errors = append(pageData.Errors, lobbyPasswordInvalid.Error())
	}
	if playerNameInvalid != nil {
		pageData.Errors = append(pageData.Errors, playerNameInvalid.Error())
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

	if len(pageData.Errors) != 0 {
		err := lobbyCreatePage.ExecuteTemplate(w, "lobby.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		//TODO CREATE LOBBY AND REDIRECT
		fmt.Println(playerName, lobbyPassword, drawingTime, rounds, maxPlayers, customWords)
	}
}

func parsePlayerName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed, errors.New("the player name must not be empty")
	}

	return trimmed, nil
}

func parseLobbyPassword(value string) (string, error) {
	return value, nil
}

func parseDrawingTime(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, parseErr
	}

	if result < lobbySettingBounds.MinDrawingTime {
		return 0, fmt.Errorf("drawing time must not be smaller than %d", lobbySettingBounds.MinDrawingTime)
	}

	if result > lobbySettingBounds.MaxDrawingTime {
		return 0, fmt.Errorf("drawing time must not be greater than %d", lobbySettingBounds.MaxDrawingTime)
	}

	return int(result), nil
}

func parseRounds(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, parseErr
	}

	if result < lobbySettingBounds.MinRounds {
		return 0, fmt.Errorf("rounds must not be smaller than %d", lobbySettingBounds.MinRounds)
	}

	if result > lobbySettingBounds.MaxRounds {
		return 0, fmt.Errorf("rounds must not be greater than %d", lobbySettingBounds.MaxRounds)
	}

	return int(result), nil
}

func parseMaxPlayers(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, parseErr
	}

	if result < lobbySettingBounds.MinMaxPlayers {
		return 0, fmt.Errorf("maximum players must not be smaller than %d", lobbySettingBounds.MinMaxPlayers)
	}

	if result > lobbySettingBounds.MaxMaxPlayers {
		return 0, fmt.Errorf("maximum players must not be greater than %d", lobbySettingBounds.MaxMaxPlayers)
	}

	return int(result), nil
}

func parseCustomWords(value string) ([]string, error) {
	result := strings.Split(value, ",")
	for index, item := range result {
		result[index] = strings.ToLower(strings.TrimSpace(item))
	}

	return result, nil
}
