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
var lobbySettingBounds = &SettingBounds{
	MinDrawingTime: 60,
	MaxDrawingTime: 300,
	MinRounds:      1,
	MaxRounds:      20,
	MinMaxPlayers:  2,
	MaxMaxPlayers:  24,
}

// SettingBounds defines the lower and upper bounds for the user-specified
// lobby creation input.
type SettingBounds struct {
	MinDrawingTime int64
	MaxDrawingTime int64
	MinRounds      int64
	MaxRounds      int64
	MinMaxPlayers  int64
	MaxMaxPlayers  int64
}

// CreatePageData defines all non-static data for the lobby create page.
type CreatePageData struct {
	*SettingBounds
	Errors      []string
	Name        string
	Password    string
	DrawingTime string
	Rounds      string
	MaxPlayers  string
	CustomWords string
}

func createDefaultLobbyCreatePageDat() *CreatePageData {
	return &CreatePageData{
		SettingBounds: lobbySettingBounds,
		DrawingTime:   "120",
		Rounds:        "4",
		MaxPlayers:    "4",
	}
}

func init() {
	var err error
	lobbyCreatePage, err = template.New("").ParseFiles("lobby_create.html", "footer.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", HomePage)
	http.HandleFunc("/lobby", ShowLobby)
	http.HandleFunc("/lobby/create", CreateLobby)
}

// HomePage servers the default page for scribble.rs, which is the page to
// create a new lobby.
func HomePage(w http.ResponseWriter, r *http.Request) {
	err := lobbyCreatePage.ExecuteTemplate(w, "lobby_create.html", createDefaultLobbyCreatePageDat())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateLobby allows creating a lobby, optionally returning errors that
// occured during creation.
func CreateLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		panic(formParseError)
	}

	playerName, playerNameInvalid := parsePlayerName(r.Form.Get("player_name"))
	password, passwordInvalid := parsePassword(r.Form.Get("lobby_password"))
	drawingTime, drawingTimeInvalid := parseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := parseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := parseCustomWords(r.Form.Get("custom_words"))

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := CreatePageData{
		SettingBounds: lobbySettingBounds,
		Name:          r.Form.Get("player_name"),
		Password:      r.Form.Get("lobby_password"),
		DrawingTime:   r.Form.Get("drawing_time"),
		Rounds:        r.Form.Get("rounds"),
		MaxPlayers:    r.Form.Get("max_players"),
		CustomWords:   r.Form.Get("custom_words"),
	}

	if passwordInvalid != nil {
		pageData.Errors = append(pageData.Errors, passwordInvalid.Error())
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
		err := lobbyCreatePage.ExecuteTemplate(w, "lobby_create.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		lobby := createLobby(password, drawingTime, rounds, maxPlayers, customWords)
		player := createPlayer(playerName)

		//FIXME Make a dedicated method that uses a mutex?
		lobby.Players = append(lobby.Players, player)

		// Use the players generated usersession and pass it as a cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "usersession",
			Value:    player.UserSession,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})

		http.Redirect(w, r, "/lobby?id="+lobby.ID, http.StatusFound)
	}
}

// ShowLobby opens a lobby, either opening it directly or asking for a username
// and or a lobby password.
func ShowLobby(w http.ResponseWriter, r *http.Request) {
	lobbyID := r.URL.Query().Get("id")
	if lobbyID == "" {
		errorPage.ExecuteTemplate(w, "error.html", "The entered URL is incorrect.")
		return
	}

	var lobby *Lobby
	for _, l := range lobbies {
		if l.ID == lobbyID {
			lobby = l
			break
		}
	}

	if lobby == nil {
		errorPage.ExecuteTemplate(w, "error.html", "The lobby does not exist.")
		return
	}

	sessionCookie, noCookieError := r.Cookie("usersession")
	var player *Player
	if noCookieError == nil {
		player = lobby.GetPlayer(sessionCookie.Value)
	}

	if player == nil {
		errorPage.ExecuteTemplate(w, "error.html", "Under construction - Come back later")
		//TODO Authenticate
	} else {
		errorPage.ExecuteTemplate(w, "error.html", "Under construction - Come back later")
		//TODO Serve lobby.
	}
}

func parsePlayerName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed, errors.New("the player name must not be empty")
	}

	return trimmed, nil
}

func parsePassword(value string) (string, error) {
	return value, nil
}

func parseDrawingTime(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the drawing time must be numeric")
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
		return 0, errors.New("the rounds amount must be numeric")
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
		return 0, errors.New("the max players amount must be numeric")
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
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	result := strings.Split(trimmedValue, ",")
	for index, item := range result {
		trimmedItem := strings.ToLower(strings.TrimSpace(item))
		if trimmedItem == "" {
			return nil, errors.New("custom words must not be empty")
		}
		result[index] = trimmedItem
	}

	return result, nil
}
