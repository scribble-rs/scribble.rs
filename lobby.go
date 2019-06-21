package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/gorilla/websocket"
)

var (
	lobbyCreatePage    *template.Template
	lobbyPage          *template.Template
	lobbySettingBounds = &SettingBounds{
		MinDrawingTime: 60,
		MaxDrawingTime: 300,
		MinRounds:      1,
		MaxRounds:      20,
		MinMaxPlayers:  2,
		MaxMaxPlayers:  24,
	}
)

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
	Password    string
	DrawingTime string
	Rounds      string
	MaxPlayers  string
	CustomWords string
}

func createDefaultLobbyCreatePageData() *CreatePageData {
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

	lobbyPage, err = template.New("").ParseFiles("lobby.html", "lobby_players.html", "lobby_word.html", "footer.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", HomePage)
	http.HandleFunc("/lobby", ShowLobby)
	http.HandleFunc("/lobby/create", CreateLobby)
	http.HandleFunc("/lobby/players", GetPlayers)
	http.HandleFunc("/lobby/wordhint", GetWordHint)
	http.HandleFunc("/lobby/rounds", GetRounds)
	http.HandleFunc("/ws", wsEndpoint)
}

// HomePage servers the default page for scribble.rs, which is the page to
// create a new lobby.
func HomePage(w http.ResponseWriter, r *http.Request) {
	err := lobbyCreatePage.ExecuteTemplate(w, "lobby_create.html", createDefaultLobbyCreatePageData())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	lobbyID := r.URL.Query().Get("id")
	if lobbyID == "" {
		errorPage.ExecuteTemplate(w, "error.html", "The entered URL is incorrect.")
		return
	}

	lobby := GetLobby(lobbyID)

	if lobby == nil {
		errorPage.ExecuteTemplate(w, "error.html", "The lobby does not exist.")
		return
	}

	sessionCookie, noCookieError := r.Cookie("usersession")
	var player *Player
	if noCookieError == nil {
		player = lobby.GetPlayer(sessionCookie.Value)
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(player.Name + " has connected")

	player.ws = ws

	go func(player *Player, socket *websocket.Conn) {
		for {
			messageType, data, err := socket.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			} else if messageType == websocket.TextMessage {
				received := &JSEvent{}
				err := json.Unmarshal(data, received)
				if err != nil {
					//TODO NO PANICS!
					panic(err)
				}

				if received.Type == "message" {
					dataAsString := (received.Data).(string)
					if dataAsString[0] == '!' {
						command := dataAsString[1:]
						switch command {
						case "start":
							//TODO
						case "1":
							//TODO
						case "2":
							//TODO
						case "3":
							//TODO
						}
					} else {
						trimmed := strings.TrimSpace(dataAsString)
						if trimmed == "" {
							continue
						}

						escaped := html.EscapeString(trimmed)
						for _, target := range lobby.Players {
							if target.ws != nil {
								target.ws.WriteJSON(JSEvent{Type: "message", Data: Message{
									Author:  player.Name,
									Content: escaped,
								}})
							}
						}
					}
				}
			}
		}
	}(player, ws)
}

// LobbyPageData is the data necessary for initially displaying all data of
// the lobbies webpage.
type LobbyPageData struct {
	Players   []*Player
	Port      int
	LobbyID   string
	WordHints []*WordHint
	Round     int
	Rounds    int
}

func getLobby(w http.ResponseWriter, r *http.Request) *Lobby {
	lobbyID := r.URL.Query().Get("id")
	if lobbyID == "" {
		errorPage.ExecuteTemplate(w, "error.html", "The entered URL is incorrect.")
		return nil
	}

	lobby := GetLobby(lobbyID)

	if lobby == nil {
		errorPage.ExecuteTemplate(w, "error.html", "The lobby does not exist.")
	}

	return lobby
}

// GetPlayers returns divs for all players in the lobby to the calling client.
func GetPlayers(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		templatingError := lobbyPage.ExecuteTemplate(w, "players", lobby.Players)
		if templatingError != nil {
			errorPage.ExecuteTemplate(w, "error.html", templatingError.Error())
		}
	}
}

// GetWordHint returns the html structure and data for the current word hint.
func GetWordHint(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		templatingError := lobbyPage.ExecuteTemplate(w, "word", lobby.WordHints)
		if templatingError != nil {
			errorPage.ExecuteTemplate(w, "error.html", templatingError.Error())
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

// CreateLobby allows creating a lobby, optionally returning errors that
// occured during creation.
func CreateLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		panic(formParseError)
	}

	password, passwordInvalid := parsePassword(r.Form.Get("lobby_password"))
	drawingTime, drawingTimeInvalid := parseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := parseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := parseCustomWords(r.Form.Get("custom_words"))

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := CreatePageData{
		SettingBounds: lobbySettingBounds,
		Password:      r.Form.Get("lobby_password"),
		DrawingTime:   r.Form.Get("drawing_time"),
		Rounds:        r.Form.Get("rounds"),
		MaxPlayers:    r.Form.Get("max_players"),
		CustomWords:   r.Form.Get("custom_words"),
	}

	if passwordInvalid != nil {
		pageData.Errors = append(pageData.Errors, passwordInvalid.Error())
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

		adjective := strings.Title(petname.Adjective())
		adverb := strings.Title(petname.Adverb())
		name := strings.Title(petname.Name())
		player := createPlayer(adverb + adjective + name)

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

// JSEvent contains an eventtype and optionally any data.
type JSEvent struct {
	Type string
	Data interface{}
}

// Message represents a message in the chatroom.
type Message struct {
	// Author is the player / thing that wrote the message
	Author string
	// Content is the actual message text.
	Content string
}

// ShowLobby opens a lobby, either opening it directly or asking for a username
// and or a lobby password.
func ShowLobby(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		sessionCookie, noCookieError := r.Cookie("usersession")
		var player *Player
		if noCookieError == nil {
			player = lobby.GetPlayer(sessionCookie.Value)
		}

		fmt.Printf("Useragent of %s: %s", r.Host, r.UserAgent())

		if player == nil {
			if len(lobby.Players) >= lobby.MaxPlayers {
				errorPage.ExecuteTemplate(w, "error.html", "Sorry, but the lobby is full.")
				return
			}

			adjective := strings.Title(petname.Adjective())
			adverb := strings.Title(petname.Adverb())
			name := strings.Title(petname.Name())
			player := createPlayer(adverb + adjective + name)

			//FIXME Make a dedicated method that uses a mutex?
			lobby.Players = append(lobby.Players, player)

			pageData := &LobbyPageData{
				Port:    *portHTTP,
				Players: lobby.Players,
				LobbyID: lobby.ID,
				Round:   lobby.Round,
				Rounds:  lobby.Rounds,
			}

			for _, player := range lobby.Players {
				if player.ws != nil {
					player.ws.WriteJSON(JSEvent{Type: "update-players"})
				}
			}

			// Use the players generated usersession and pass it as a cookie.
			http.SetCookie(w, &http.Cookie{
				Name:     "usersession",
				Value:    player.UserSession,
				Path:     "/",
				SameSite: http.SameSiteStrictMode,
			})

			lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		} else {
			pageData := &LobbyPageData{
				Port:    *portHTTP,
				Players: lobby.Players,
				LobbyID: lobby.ID,
				Round:   lobby.Round,
				Rounds:  lobby.Rounds,
			}

			lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		}
	}
}

// WordHint describes a character of the word that is to be guessed, whether
// the character should be shown and whether it should be underlined on the
// UI.
type WordHint struct {
	Character string
	Show      bool
	Underline bool
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
