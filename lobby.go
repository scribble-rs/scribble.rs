package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	commands "github.com/Bios-Marcel/cmdp"
	"github.com/Bios-Marcel/discordemojimap"
	"github.com/agnivade/levenshtein"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/gorilla/websocket"
)

var (
	lobbyCreatePage    *template.Template
	lobbyPage          *template.Template
	lobbySettingBounds = &SettingBounds{
		MinDrawingTime:       60,
		MaxDrawingTime:       300,
		MinRounds:            1,
		MaxRounds:            20,
		MinMaxPlayers:        2,
		MaxMaxPlayers:        24,
		MinClientsPerIPLimit: 1,
		MaxClientsPerIPLimit: 24,
	}
	supportedLanguages = []string{"English", "Italian", "German"}
)

// SettingBounds defines the lower and upper bounds for the user-specified
// lobby creation input.
type SettingBounds struct {
	MinDrawingTime       int64
	MaxDrawingTime       int64
	MinRounds            int64
	MaxRounds            int64
	MinMaxPlayers        int64
	MaxMaxPlayers        int64
	MinClientsPerIPLimit int64
	MaxClientsPerIPLimit int64
}

// CreatePageData defines all non-static data for the lobby create page.
type CreatePageData struct {
	*SettingBounds
	Errors            []string
	Password          string
	Languages         []string
	DrawingTime       string
	Rounds            string
	MaxPlayers        string
	CustomWords       string
	CustomWordsChance string
	ClientsPerIPLimit string
	EnableVotekick    string
	Language          string
}

func createDefaultLobbyCreatePageData() *CreatePageData {

	return &CreatePageData{
		SettingBounds:     lobbySettingBounds,
		Languages:         supportedLanguages,
		DrawingTime:       "120",
		Rounds:            "4",
		MaxPlayers:        "12",
		CustomWordsChance: "50",
		ClientsPerIPLimit: "1",
		EnableVotekick:    "true",
		Language:          supportedLanguages[0],
	}
}

func init() {
	var err error

	lobbyCreatePage, err = template.New("lobby_create.html").Parse(readTemplateFile("lobby_create.html"))
	if err != nil {
		panic(err)
	}
	lobbyCreatePage, err = lobbyCreatePage.New("footer.html").Parse(readTemplateFile("footer.html"))
	if err != nil {
		panic(err)
	}

	lobbyPage, err = template.New("lobby.html").Parse(readTemplateFile("lobby.html"))
	if err != nil {
		panic(err)
	}
	lobbyPage, err = lobbyPage.New("lobby_players.html").Parse(readTemplateFile("lobby_players.html"))
	if err != nil {
		panic(err)
	}
	lobbyPage, err = lobbyPage.New("lobby_word.html").Parse(readTemplateFile("lobby_word.html"))
	if err != nil {
		panic(err)
	}
	lobbyPage, err = lobbyPage.New("footer.html").Parse(readTemplateFile("footer.html"))
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
	//This issue can happen if you illegally request a websocket connection without ever having had
	//a usersession or your client having deleted the usersession cookie.
	if noCookieError != nil {
		errorPage.ExecuteTemplate(w, "error.html", "You are not a player of this lobby.")
		return
	}
	player := lobby.GetPlayer(sessionCookie.Value)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(player.Name + " has connected")

	player.ws = ws
	player.State = Guessing
	triggerPlayersUpdate(lobby)
	ws.SetCloseHandler(func(code int, text string) error {
		//We want to avoid calling the handler twice.
		if player.ws == nil {
			return nil
		}

		player.State = Disconnected
		player.ws = nil
		isAnyPlayerConnected := false
		for _, otherPlayer := range lobby.Players {
			if otherPlayer.ws != nil && otherPlayer.State != Disconnected {
				isAnyPlayerConnected = true
				break
			}
		}

		if !isAnyPlayerConnected {
			RemoveLobby(lobbyID)
			log.Printf("There are currently %d open lobbies.\n", len(lobbies))
		} else {
			triggerPlayersUpdate(lobby)
		}

		return nil
	})

	if len(lobby.currentDrawing) > 0 {
		sendError := player.WriteAsJSON(&JSEvent{
			Type: "pixels",
			Data: lobby.currentDrawing,
		})
		if sendError != nil {
			log.Printf("Error sending drawing to player: %s", sendError)
		}
	}

	go wsListen(lobby, player, ws)
}

// PixelEvent is basically the same as JSEvent, but with a specific Data type.
// We use this for reparsing as soon as we know that the type is right. It's
// a bit unperformant, but will do for now.
type PixelEvent struct {
	Type string
	Data Pixel
}

func wsListen(lobby *Lobby, player *Player, socket *websocket.Conn) {
	for {
		messageType, data, err := socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) ||
				//This happens when the server closes the connection. It will cause 1000 retries followed by a panic.
				strings.Contains(err.Error(), "use of closed network connection") {
				//Make sure that the sockethandler is called
				socket.CloseHandler()
				log.Println(player.Name + " disconnected.")
				return
			} else {
				log.Printf("Error reading from socket: %s\n", err)
			}
		} else if messageType == websocket.TextMessage {
			received := &JSEvent{}
			err := json.Unmarshal(data, received)
			if err != nil {
				log.Printf("Error unmarshalling message: %s\n", err)
				sendError := player.WriteAsJSON(JSEvent{Type: "system-message", Data: fmt.Sprintf("An error occured trying to read your request, please report the error via GitHub: %s!", err)})
				if sendError != nil {
					log.Printf("Error sending errormessage: %s\n", sendError)
				}
				continue
			}

			if received.Type == "message" {
				dataAsString, isString := (received.Data).(string)
				if !isString {
					continue
				}

				if strings.HasPrefix(dataAsString, "!") {
					handleCommand(dataAsString[1:], player, lobby)
				} else {
					handleMessage(dataAsString, player, lobby)
				}
			} else if received.Type == "pixel" {
				if lobby.Drawer == player {
					var pixel PixelEvent
					pixelErr := json.Unmarshal(data, &pixel)
					if pixelErr != nil {
						log.Printf("Error unmarshalling pixel: %s", pixelErr)
					} else {
						pixel.Data.Type = "pixel"
						lobby.AppendPixel(&pixel.Data)
					}
					for _, otherPlayer := range lobby.Players {
						if otherPlayer != player && otherPlayer.State != Disconnected && otherPlayer.ws != nil {
							otherPlayer.WriteMessage(websocket.TextMessage, data)
						}
					}
				}
			} else if received.Type == "fill" {
				if lobby.Drawer == player {
					var pixel PixelEvent
					pixelErr := json.Unmarshal(data, &pixel)
					if pixelErr != nil {
						log.Printf("Error unmarshalling pixel: %s", pixelErr)
					} else {
						pixel.Data.Type = "fill"
						lobby.AppendPixel(&pixel.Data)
					}
					for _, otherPlayer := range lobby.Players {
						if otherPlayer != player && otherPlayer.State != Disconnected && otherPlayer.ws != nil {
							otherPlayer.WriteMessage(websocket.TextMessage, data)
						}
					}
				}

			} else if received.Type == "clear-drawing-board" {
				if lobby.Drawer == player {
					lobby.ClearDrawing()
					for _, otherPlayer := range lobby.Players {
						if otherPlayer.State != Disconnected && otherPlayer.ws != nil {
							otherPlayer.WriteMessage(websocket.TextMessage, data)
						}
					}
				}
			} else if received.Type == "choose-word" {
				chosenIndex, isInt := (received.Data).(int)
				if !isInt {
					asFloat, isFloat32 := (received.Data).(float64)
					if isFloat32 && asFloat < 4 {
						chosenIndex = int(asFloat)
					} else {
						fmt.Println("Invalid data")
						continue
					}
				}

				if player == lobby.Drawer && len(lobby.WordChoice) > 0 && chosenIndex >= 0 && chosenIndex <= 2 {
					lobby.CurrentWord = lobby.WordChoice[chosenIndex]
					lobby.WordChoice = nil
					lobby.WordHints = createWordHintFor(lobby.CurrentWord)
					lobby.WordHintsShown = showAllInWordHints(lobby.WordHints)
					triggerWordHintUpdate(lobby)
					if lobby.Drawer.State != Disconnected && lobby.Drawer.ws != nil {
						lobby.Drawer.WriteAsJSON(JSEvent{Type: "your-turn"})
					}
				}
			} else if received.Type == "kick-vote" {
				toKickID, isString := (received.Data).(string)
				if !isString {
					fmt.Println("Invalid data")
					continue
				}
				if !lobby.EnableVotekick {
					// Votekicking is disabled in the lobby
					// We tell the user and do not continue with the event
					player.WriteAsJSON(JSEvent{Type: "system-message", Data: "Votekick is disabled in this lobby!"})
				} else {
					handleKickEvent(lobby, player, toKickID)
				}
			}
		}
	}
}

func handleMessage(input string, sender *Player, lobby *Lobby) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return
	}

	if lobby.CurrentWord == "" {
		sendMessageToAll(trimmed, sender, lobby)
		return
	}

	if sender.State == Drawing || sender.State == Standby {
		sendMessageToAllNonGuessing(trimmed, sender, lobby)
	} else if sender.State == Guessing {
		lowerCasedInput := strings.ToLower(trimmed)
		lowerCasedSearched := strings.ToLower(lobby.CurrentWord)
		if lowerCasedSearched == lowerCasedInput {
			sender.LastScore = int(math.Ceil(math.Pow(math.Max(float64(lobby.TimeLeft), 1), 1.3) * 2))
			sender.Score += sender.LastScore
			lobby.scoreEarnedByGuessers += sender.LastScore
			sender.State = Standby
			sender.Icon = "✔️"
			if sender.State != Disconnected && sender.ws != nil {
				sender.WriteAsJSON(JSEvent{Type: "system-message", Data: "You have correctly guessed the word."})
			}

			if !lobby.isAnyoneStillGuessing() {
				endRound(lobby)
			} else {
				if sender.State != Disconnected && sender.ws != nil {
					sender.WriteAsJSON(JSEvent{Type: "update-wordhint"})
				}
				recalculateRanks(lobby)
				triggerCorrectGuessEvent(lobby)
			}

			return
		} else if levenshtein.ComputeDistance(lowerCasedInput, lowerCasedSearched) == 1 &&
			sender.State != Disconnected && sender.ws != nil {
			sender.WriteAsJSON(JSEvent{Type: "system-message", Data: fmt.Sprintf("'%s' is very close.", trimmed)})
		}

		sendMessageToAll(trimmed, sender, lobby)
	}
}

func (lobby *Lobby) isAnyoneStillGuessing() bool {
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.State == Guessing {
			return true
		}
	}

	return false
}

func sendMessageToAll(message string, sender *Player, lobby *Lobby) {
	escaped := html.EscapeString(discordemojimap.Replace(message))
	for _, target := range lobby.Players {
		if target.State != Disconnected && target.ws != nil {
			target.WriteAsJSON(JSEvent{Type: "message", Data: Message{
				Author:  html.EscapeString(sender.Name),
				Content: escaped,
			}})
		}
	}
}

func sendMessageToAllNonGuessing(message string, sender *Player, lobby *Lobby) {
	escaped := html.EscapeString(discordemojimap.Replace(message))
	for _, target := range lobby.Players {
		if target.State != Disconnected && target.State != Guessing && target.ws != nil {
			target.WriteAsJSON(JSEvent{Type: "non-guessing-player-message", Data: Message{
				Author:  html.EscapeString(sender.Name),
				Content: escaped,
			}})
		}
	}
}

func handleKickEvent(lobby *Lobby, player *Player, toKickID string) {
	//Kicking yourself isn't allowed
	if toKickID == player.ID {
		return
	}

	//A player can't vote twice to kick someone
	if player.votedForKick[toKickID] {
		return
	}

	toKick := -1
	for index, otherPlayer := range lobby.Players {
		if otherPlayer.ID == toKickID {
			toKick = index
			break
		}
	}

	//If we haven't found the player, we can't kick him/her.
	if toKick != -1 {
		player.votedForKick[toKickID] = true
		playerToKick := lobby.Players[toKick]

		var voteKickCount int
		for _, otherPlayer := range lobby.Players {
			if otherPlayer.votedForKick[toKickID] == true {
				voteKickCount++
			}
		}

		votesNeeded := 1
		if len(lobby.Players)%2 == 0 {
			votesNeeded = len(lobby.Players) / 2
		} else {
			votesNeeded = (len(lobby.Players) / 2) + 1
		}

		lobby.WriteGlobalSystemMessage(fmt.Sprintf("(%d/%d) players voted to kick %s", voteKickCount, votesNeeded, playerToKick.Name))

		if voteKickCount >= votesNeeded {
			//Since the player is already kicked, we first clean up the kicking information related to that player
			for _, otherPlayer := range lobby.Players {
				if otherPlayer.votedForKick[toKickID] == true {
					delete(player.votedForKick, toKickID)
					break
				}
			}

			lobby.WriteGlobalSystemMessage(fmt.Sprintf("%s has been kicked from the lobby", playerToKick.Name))

			if lobby.Drawer == playerToKick {
				lobby.WriteGlobalSystemMessage("Since the kicked player has been drawing, none of you will get any points this round.")
				//Since the drawing person has been kicked, that probably means that he/she was trolling, therefore
				//we redact everyones last earned score.
				for _, otherPlayer := range lobby.Players {
					otherPlayer.Score -= otherPlayer.LastScore
					otherPlayer.LastScore = 0
				}
				lobby.scoreEarnedByGuessers = 0
				//We must absolutely not set lobby.Drawer to nil, since this would cause the drawing order to be ruined.
			}

			if playerToKick.ws != nil {
				playerToKick.State = Disconnected
				playerToKick.ws.Close()
			}
			lobby.Players = append(lobby.Players[:toKick], lobby.Players[toKick+1:]...)

			recalculateRanks(lobby)

			//If the owner is kicked, we choose the next best person as the owner.
			if lobby.Owner == playerToKick {
				for _, otherPlayer := range lobby.Players {
					potentialOwner := otherPlayer
					if potentialOwner.State != Disconnected && potentialOwner.ws != nil {
						lobby.Owner = potentialOwner
						lobby.WriteGlobalSystemMessage(fmt.Sprintf("%s is the new lobby owner.", potentialOwner.Name))
						break
					}
				}
			}

			triggerPlayersUpdate(lobby)

			if lobby.Drawer == playerToKick || !lobby.isAnyoneStillGuessing() {
				endRound(lobby)
			}
		}
	}
}

func handleCommand(commandString string, caller *Player, lobby *Lobby) {
	command := commands.ParseCommand(commandString)
	if len(command) >= 1 {
		switch strings.ToLower(command[0]) {
		case "start":
			commandStart(caller, lobby)
		case "setmp":
			commandSetMP(caller, lobby, command)
		case "help":
			//TODO
		case "nick", "name", "username", "nickname", "playername", "alias":
			commandNick(caller, lobby, command)
		}
	}
}

func commandStart(caller *Player, lobby *Lobby) {
	if lobby.Round == 0 && caller == lobby.Owner {
		for _, otherPlayer := range lobby.Players {
			otherPlayer.Score = 0
		}

		advanceLobby(lobby)
	}
}

func commandNick(caller *Player, lobby *Lobby, args []string) {
	if len(args) == 1 {
		caller.Name = generatePlayerName()
		if caller.State != Disconnected && caller.ws != nil {
			caller.WriteAsJSON(JSEvent{Type: "reset-username"})
		}
		triggerPlayersUpdate(lobby)
	} else if len(args) == 2 {
		newName := strings.TrimSpace(args[1])
		if len(newName) == 0 {
			caller.Name = generatePlayerName()
			if caller.State != Disconnected && caller.ws != nil {
				caller.WriteAsJSON(JSEvent{Type: "reset-username"})
			}
			triggerPlayersUpdate(lobby)
		} else if len(newName) <= 30 {
			fmt.Printf("%s is now %s\n", caller.Name, newName)
			caller.Name = newName
			if caller.State != Disconnected && caller.ws != nil {
				caller.WriteAsJSON(JSEvent{Type: "persist-username", Data: caller.Name})
			}
			triggerPlayersUpdate(lobby)
		}
	}
	//TODO Else, show error
}

func commandSetMP(caller *Player, lobby *Lobby, args []string) {
	if caller == lobby.Owner {
		if len(args) < 2 {
			return
		}

		newMaxPlayersValue := strings.TrimSpace(args[1])
		newMaxPlayersValueInt, err := strconv.ParseInt(newMaxPlayersValue, 10, 64)
		if err == nil {
			if int(newMaxPlayersValueInt) >= len(lobby.Players) && newMaxPlayersValueInt <= lobbySettingBounds.MaxMaxPlayers && newMaxPlayersValueInt >= lobbySettingBounds.MinMaxPlayers {
				lobby.MaxPlayers = int(newMaxPlayersValueInt)

				lobby.WriteGlobalSystemMessage(fmt.Sprintf("MaxPlayers value has been changed to %d", lobby.MaxPlayers))
			} else {
				if len(lobby.Players) > int(lobbySettingBounds.MinMaxPlayers) {
					caller.WriteAsJSON(JSEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value should be between %d and %d.", len(lobby.Players), lobbySettingBounds.MaxMaxPlayers)})
				} else {
					caller.WriteAsJSON(JSEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value should be between %d and %d.", lobbySettingBounds.MinMaxPlayers, lobbySettingBounds.MaxMaxPlayers)})
				}
			}
		} else {
			caller.WriteAsJSON(JSEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value must be numeric.")})
		}
	} else {
		caller.WriteAsJSON(JSEvent{Type: "system-message", Data: fmt.Sprintf("Only the lobby owner can change MaxPlayers setting.")})
	}
}

func endRound(lobby *Lobby) {
	var roundOverMessage string
	if lobby.CurrentWord == "" {
		roundOverMessage = "Round over. No word was chosen."
	} else {
		roundOverMessage = fmt.Sprintf("Round over. The word was '%s'", lobby.CurrentWord)
	}

	//The drawer can potentially be null if he's kicked, in that case we proceed with the round if anyone has already
	drawer := lobby.Drawer
	if drawer != nil && lobby.scoreEarnedByGuessers > 0 {
		averageScore := float64(lobby.scoreEarnedByGuessers) / float64(len(lobby.Players)-1)
		if averageScore > 0 {
			drawer.LastScore = int(averageScore * 1.1)
			drawer.Score += drawer.LastScore
		}
	}

	lobby.scoreEarnedByGuessers = 0
	lobby.alreadyUsedWords = append(lobby.alreadyUsedWords, lobby.CurrentWord)
	lobby.CurrentWord = ""
	lobby.WordHints = nil

	//If the round ends and people still have guessing, that means the "Last" value
	////for the next turn has to be "no score earned".
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.State == Guessing {
			otherPlayer.LastScore = 0
		}
	}

	lobby.WriteGlobalSystemMessage(roundOverMessage)

	advanceLobby(lobby)
}

func advanceLobby(lobby *Lobby) {
	if lobby.timeLeftTicker != nil {
		lobby.timeLeftTicker.Stop()
		lobby.timeLeftTicker = nil
		lobby.timeLeftTickerReset <- struct{}{}
	}

	lobby.TimeLeft = lobby.DrawingTime

	for _, otherPlayer := range lobby.Players {
		otherPlayer.State = Guessing
		otherPlayer.Icon = ""
		otherPlayer.votedForKick = make(map[string]bool)
	}

	lobby.ClearDrawing()

	if lobby.Drawer == nil {
		lobby.Drawer = lobby.Players[0]
		lobby.Round++
	} else {
		//If everyone has drawn once (e.g. a round has passed)
		if lobby.Drawer == lobby.Players[len(lobby.Players)-1] {
			if lobby.Round == lobby.Rounds {
				lobby.Drawer = nil
				lobby.Round = 0

				recalculateRanks(lobby)
				triggerPlayersUpdate(lobby)

				lobby.WriteGlobalSystemMessage("Game over. Type !start again to start a new round.")

				return
			}

			lobby.Round++
			lobby.Drawer = lobby.Players[0]
		} else {
			selectNextDrawer(lobby)
		}
	}

	lobby.Drawer.State = Drawing
	lobby.Drawer.Icon = "✏️"
	lobby.WordChoice = GetRandomWords(lobby)
	if lobby.Drawer.State != Disconnected && lobby.Drawer.ws != nil {
		lobby.Drawer.WriteAsJSON(JSEvent{Type: "prompt-words", Data: lobby.WordChoice})
	}

	lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
	go func() {
		showNextHintInSeconds := lobby.DrawingTime / 3
		hintsLeft := 2

		for {
			select {
			case <-lobby.timeLeftTicker.C:
				lobby.TimeLeft--
				triggerTimeLeftUpdate(lobby)
				if hintsLeft > 0 {
					showNextHintInSeconds--
					if showNextHintInSeconds == 0 {
						showNextHintInSeconds = lobby.DrawingTime / 3
						hintsLeft--
						//FIXME If a word is chosen lates, less hints will come overall.
						if lobby.WordHints != nil {
							for {
								randomIndex := rand.Int() % len(lobby.WordHints)
								if !lobby.WordHints[randomIndex].Show {
									lobby.WordHints[randomIndex].Show = true
									triggerWordHintUpdate(lobby)
									break
								}
							}
						}
					}
				}
				if lobby.TimeLeft == 0 {
					go endRound(lobby)
				}
			case <-lobby.timeLeftTickerReset:
				return
			}
		}
	}()

	recalculateRanks(lobby)

	triggerNextTurn(lobby)
	triggerPlayersUpdate(lobby)
	triggerRoundsUpdate(lobby)
	triggerWordHintUpdate(lobby)
}

func recalculateRanks(lobby *Lobby) {
	for _, a := range lobby.Players {
		playersThatAreHigher := 0
		for _, b := range lobby.Players {
			if b.Score > a.Score {
				playersThatAreHigher++
			}
		}

		a.Rank = playersThatAreHigher + 1
	}
}

func selectNextDrawer(lobby *Lobby) {
	for playerIndex, otherPlayer := range lobby.Players {
		if otherPlayer == lobby.Drawer {
			lobby.Drawer = lobby.Players[playerIndex+1]
			return
		}
	}

	for _, otherPlayer := range lobby.Players {
		if otherPlayer == lobby.Drawer {
			return
		}

		if otherPlayer.State == Disconnected {
			continue
		}

		lobby.Drawer = otherPlayer
	}
}

func createWordHintFor(word string) []*WordHint {
	wordHints := make([]*WordHint, 0, len(word))
	for _, char := range word {
		irrelevantChar := char == ' ' || char == '_' || char == '-'
		wordHints = append(wordHints, &WordHint{
			Character: string(char),
			Show:      irrelevantChar,
			Underline: !irrelevantChar,
		})
	}

	return wordHints
}

func showAllInWordHints(hints []*WordHint) []*WordHint {
	newHints := make([]*WordHint, len(hints), len(hints))
	for index, hint := range hints {
		newHints[index] = &WordHint{
			Character: hint.Character,
			Show:      true,
			Underline: hint.Underline,
		}
	}

	return newHints
}

func triggerNextTurn(lobby *Lobby) {
	triggerSimpleUpdateEvent("next-turn", lobby)
}

func triggerPlayersUpdate(lobby *Lobby) {
	triggerSimpleUpdateEvent("update-players", lobby)
}

func triggerCorrectGuessEvent(lobby *Lobby) {
	triggerSimpleUpdateEvent("correct-guess", lobby)
}

func triggerWordHintUpdate(lobby *Lobby) {
	triggerSimpleUpdateEvent("update-wordhint", lobby)
}

func triggerRoundsUpdate(lobby *Lobby) {
	triggerSimpleUpdateEvent("update-rounds", lobby)
}

func triggerTimeLeftUpdate(lobby *Lobby) {
	event := &JSEvent{Type: "update-time", Data: lobby.TimeLeft}
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.State != Disconnected && otherPlayer.ws != nil {
			otherPlayer.WriteAsJSON(event)
		}
	}
}

func triggerSimpleUpdateEvent(eventType string, lobby *Lobby) {
	event := &JSEvent{Type: eventType}
	for _, otherPlayer := range lobby.Players {
		go func(player *Player) {
			player.WriteAsJSON(event)
		}(otherPlayer)
	}
}

// LobbyPageData is the data necessary for initially displaying all data of
// the lobbies webpage.
type LobbyPageData struct {
	Players        []*Player
	LobbyID        string
	WordHints      []*WordHint
	Round          int
	Rounds         int
	EnableVotekick bool
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
		templatingError := lobbyPage.ExecuteTemplate(w, "players", lobby)
		if templatingError != nil {
			errorPage.ExecuteTemplate(w, "error.html", templatingError.Error())
		}
	}
}

// GetWordHint returns the html structure and data for the current word hint.
func GetWordHint(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		sessionCookie, noCookieError := r.Cookie("usersession")
		if noCookieError != nil {
			errorPage.ExecuteTemplate(w, "error.html", "You aren't part of this lobby.")
			return
		}

		player := lobby.GetPlayer(sessionCookie.Value)
		if player == nil {
			errorPage.ExecuteTemplate(w, "error.html", "You aren't part of this lobby.")
			return
		}

		var wordHints []*WordHint
		if player.State == Drawing || player.State == Standby {
			wordHints = lobby.WordHintsShown
		} else {
			wordHints = lobby.WordHints
		}

		templatingError := lobbyPage.ExecuteTemplate(w, "word", wordHints)
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
	language, languageInvalid := parseLanguage(r.Form.Get("language"))
	drawingTime, drawingTimeInvalid := parseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := parseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := parseCustomWords(r.Form.Get("custom_words"))
	customWordChance, customWordChanceInvalid := parseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := parseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick := r.Form.Get("enable_votekick") == "true"

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := CreatePageData{
		SettingBounds:     lobbySettingBounds,
		Languages:         supportedLanguages,
		Password:          r.Form.Get("lobby_password"),
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
	if customWordChanceInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		pageData.Errors = append(pageData.Errors, clientsPerIPLimitInvalid.Error())
	}

	if len(pageData.Errors) != 0 {
		err := lobbyCreatePage.ExecuteTemplate(w, "lobby_create.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		lobby := createLobby(password, drawingTime, rounds, maxPlayers, customWords, customWordChance, clientsPerIPLimit, enableVotekick)
		var playerName string
		usernameCookie, noCookieError := r.Cookie("username")
		if noCookieError == nil {
			playerName = usernameCookie.Value
		} else {
			playerName = generatePlayerName()
		}

		player := createPlayer(playerName)

		//FIXME Make a dedicated method that uses a mutex?
		lobby.Players = append(lobby.Players, player)
		lobby.Owner = player
		// Read wordlist according to the chosen language
		lobby.Words = readWordList(language)

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

func generatePlayerName() string {
	adjective := strings.Title(petname.Adjective())
	adverb := strings.Title(petname.Adverb())
	name := strings.Title(petname.Name())
	return adverb + adjective + name
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

// ShowLobby opens a lobby, either opening it directly or asking for a lobby
// password.
func ShowLobby(w http.ResponseWriter, r *http.Request) {
	lobby := getLobby(w, r)
	if lobby != nil {
		// TODO Improve this. Return metadata or so instead.
		userAgent := strings.ToLower(r.UserAgent())
		if !(strings.Contains(userAgent, "gecko") || strings.Contains(userAgent, "chrom") || strings.Contains(userAgent, "opera") || strings.Contains(userAgent, "safari")) {
			errorPage.ExecuteTemplate(w, "error.html", "Sorry, no robots allowed.")
			return
		}

		//FIXME Temporary
		if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "android") {
			errorPage.ExecuteTemplate(w, "error.html", "Sorry, mobile is currently not supported.")
			return
		}

		sessionCookie, noCookieError := r.Cookie("usersession")
		var player *Player
		if noCookieError == nil {
			player = lobby.GetPlayer(sessionCookie.Value)
		}

		if player == nil {
			if len(lobby.Players) >= lobby.MaxPlayers {
				errorPage.ExecuteTemplate(w, "error.html", "Sorry, but the lobby is full.")
				return
			}

			matches := 0
			for _, otherPlayer := range lobby.Players {
				socket := otherPlayer.ws
				if socket != nil && remoteAddressToSimpleIP(socket.RemoteAddr().String()) == remoteAddressToSimpleIP(r.RemoteAddr) {
					matches++
				}
			}

			if matches >= lobby.clientsPerIPLimit {
				errorPage.ExecuteTemplate(w, "error.html", "Sorry, but you have exceeded the maximum number of clients per IP.")
				return
			}

			var playerName string
			usernameCookie, noCookieError := r.Cookie("username")
			if noCookieError == nil {
				playerName = usernameCookie.Value
			} else {
				playerName = generatePlayerName()
			}

			player := createPlayer(playerName)

			//FIXME Make a dedicated method that uses a mutex?
			lobby.Players = append(lobby.Players, player)

			recalculateRanks(lobby)

			pageData := &LobbyPageData{
				Players:        lobby.Players,
				LobbyID:        lobby.ID,
				Round:          lobby.Round,
				Rounds:         lobby.Rounds,
				EnableVotekick: lobby.EnableVotekick,
			}

			for _, player := range lobby.Players {
				if player.ws != nil {
					player.WriteAsJSON(JSEvent{Type: "update-players"})
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
				Players:        lobby.Players,
				LobbyID:        lobby.ID,
				Round:          lobby.Round,
				Rounds:         lobby.Rounds,
				EnableVotekick: lobby.EnableVotekick,
			}

			lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
		}
	}
}

func remoteAddressToSimpleIP(input string) string {
	address := input
	lastIndexOfDoubleColon := strings.LastIndex(address, ":")
	if lastIndexOfDoubleColon != -1 {
		address = address[:lastIndexOfDoubleColon]
	}

	return strings.TrimSuffix(strings.TrimPrefix(address, "["), "]")

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

func parseLanguage(value string) (string, error) {
	for _, supportedLanguage := range supportedLanguages {
		if value == supportedLanguage {
			return value, nil
		}
	}

	return "", errors.New("the given language doesn't match any supported langauge")
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

func parseClientsPerIPLimit(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the clients per IP limit must be numeric")
	}

	if result < lobbySettingBounds.MinClientsPerIPLimit {
		return 0, fmt.Errorf("the clients per IP limit must not be lower than %d", lobbySettingBounds.MinClientsPerIPLimit)
	}

	if result > lobbySettingBounds.MaxClientsPerIPLimit {
		return 0, fmt.Errorf("the clients per IP limit must not be higher than %d", lobbySettingBounds.MaxClientsPerIPLimit)
	}

	return int(result), nil
}

func parseCustomWordsChance(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the custom word chance must be numeric")
	}

	if result < 0 {
		return 0, errors.New("custom word chance must not be lower than 0")
	}

	if result > 100 {
		return 0, errors.New("custom word chance must not be higher than 100")
	}

	return int(result), nil
}
