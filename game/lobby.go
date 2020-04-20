package game

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kennygrant/sanitize"
	commands "github.com/Bios-Marcel/cmdp"
	"github.com/Bios-Marcel/discordemojimap"
	"github.com/agnivade/levenshtein"
	petname "github.com/dustinkirkland/golang-petname"
)

var (
	createDeleteMutex          = &sync.Mutex{}
	lobbies           []*Lobby = nil
)

var (
	LobbySettingBounds = &SettingBounds{
		MinDrawingTime:       60,
		MaxDrawingTime:       300,
		MinRounds:            1,
		MaxRounds:            20,
		MinMaxPlayers:        2,
		MaxMaxPlayers:        24,
		MinClientsPerIPLimit: 1,
		MaxClientsPerIPLimit: 24,
	}
	SupportedLanguages = map[string]string{
		"english": "English",
		"italian": "Italian",
		"german":  "German",
		"french":  "French",
		"dutch":   "Dutch",
	}
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

// LineEvent is basically the same as JSEvent, but with a specific Data type.
// We use this for reparsing as soon as we know that the type is right. It's
// a bit unperformant, but will do for now.
type LineEvent struct {
	Type string `json:"type"`
	Data *Line  `json:"data"`
}

// LineEvent is basically the same as JSEvent, but with a specific Data type.
// We use this for reparsing as soon as we know that the type is right. It's
// a bit unperformant, but will do for now.
type FillEvent struct {
	Type string `json:"type"`
	Data *Fill  `json:"data"`
}

func HandleEvent(raw []byte, received *JSEvent, lobby *Lobby, player *Player) error {
	if received.Type == "message" {
		dataAsString, isString := (received.Data).(string)
		if !isString {
			return fmt.Errorf("invalid data received: '%s'", received.Data)
		}

		if strings.HasPrefix(dataAsString, "!") {
			handleCommand(dataAsString[1:], player, lobby)
		} else {
			handleMessage(dataAsString, player, lobby)
		}
	} else if received.Type == "line" {
		if lobby.canDraw(player) {
			line := &LineEvent{}
			jsonError := json.Unmarshal(raw, line)
			if jsonError != nil {
				return fmt.Errorf("error decoding data: %s", jsonError)
			}
			lobby.AppendLine(line)

			//We directly forward the event, as it seems to be valid.
			SendDataToConnectedPlayers(player, lobby, received)
		}
	} else if received.Type == "fill" {
		if lobby.canDraw(player) {
			fill := &FillEvent{}
			jsonError := json.Unmarshal(raw, fill)
			if jsonError != nil {
				return fmt.Errorf("error decoding data: %s", jsonError)
			}
			lobby.AppendFill(fill)

			//We directly forward the event, as it seems to be valid.
			SendDataToConnectedPlayers(player, lobby, received)
		}
	} else if received.Type == "clear-drawing-board" {
		if lobby.canDraw(player) && len(lobby.CurrentDrawing) > 0 {
			lobby.ClearDrawing()
			SendDataToConnectedPlayers(player, lobby, received)
		}
	} else if received.Type == "choose-word" {
		chosenIndex, isInt := (received.Data).(int)
		if !isInt {
			asFloat, isFloat32 := (received.Data).(float64)
			if isFloat32 && asFloat < 4 {
				chosenIndex = int(asFloat)
			} else {
				return fmt.Errorf("invalid data in choose-word event: %v", received.Data)
			}
		}

		drawer := lobby.Drawer
		if player == drawer && len(lobby.WordChoice) > 0 && chosenIndex >= 0 && chosenIndex <= 2 {
			lobby.CurrentWord = lobby.WordChoice[chosenIndex]
			lobby.WordChoice = nil
			lobby.WordHints = createWordHintFor(lobby.CurrentWord, false)
			lobby.WordHintsShown = createWordHintFor(lobby.CurrentWord, true)
			triggerWordHintUpdate(lobby)
		}
	} else if received.Type == "kick-vote" {
		if !lobby.EnableVotekick {
			// Votekicking is disabled in the lobby
			// We tell the user and do not continue with the event
			WriteAsJSON(player, JSEvent{Type: "system-message", Data: "Votekick is disabled in this lobby!"})
		} else {
			toKickID, isString := (received.Data).(string)
			if !isString {
				return fmt.Errorf("invalid data in kick-vote event: %v", received.Data)
			}

			handleKickEvent(lobby, player, toKickID)
		}
	} else if received.Type == "start" {
		if lobby.Round == 0 && player == lobby.Owner {
			//We are reseting each players score, since players could
			//technically be player a second game after the last one
			//has already ended.
			for _, otherPlayer := range lobby.Players {
				otherPlayer.Score = 0
				otherPlayer.LastScore = 0
				//Since nobody has any points in the beginning, everyone has practically
				//the same rank, therefore y'll winners for now.
				otherPlayer.Rank = 1
			}

			advanceLobby(lobby)
		}
	} else if received.Type == "name-change" {
		newName, isString := (received.Data).(string)
		if !isString {
			return fmt.Errorf("invalid data in name-change event: %v", received.Data)
		}
		commandNick(player, lobby, newName)
	}

	return nil
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

		normInput := removeAccents(lowerCasedInput)
		normSearched := removeAccents(lowerCasedSearched)

		if normSearched == normInput {
			secondsLeft := lobby.RoundEndTime/1000 - time.Now().UTC().UnixNano()/1000000000
			sender.LastScore = int(math.Ceil(math.Pow(math.Max(float64(secondsLeft), 1), 1.3) * 2))
			sender.Score += sender.LastScore
			lobby.scoreEarnedByGuessers += sender.LastScore
			sender.State = Standby
			WriteAsJSON(sender, JSEvent{Type: "system-message", Data: "You have correctly guessed the word."})

			if !lobby.isAnyoneStillGuessing() {
				endTurn(lobby)
			} else {
				//Since the word has been guessed correctly, we reveal it.
				WriteAsJSON(sender, JSEvent{Type: "update-wordhint", Data: lobby.WordHintsShown})
				recalculateRanks(lobby)
				triggerCorrectGuessEvent(lobby)
				triggerPlayersUpdate(lobby)
			}

			return
		} else if levenshtein.ComputeDistance(normInput, normSearched) == 1 {
			WriteAsJSON(sender, JSEvent{Type: "system-message", Data: fmt.Sprintf("'%s' is very close.", trimmed)})
		}

		sendMessageToAll(trimmed, sender, lobby)
	}
}

func (lobby *Lobby) isAnyoneStillGuessing() bool {
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.State == Guessing && otherPlayer.Connected {
			return true
		}
	}

	return false
}

func sendMessageToAll(message string, sender *Player, lobby *Lobby) {
	escaped := html.EscapeString(discordemojimap.Replace(message))
	for _, target := range lobby.Players {
		WriteAsJSON(target, JSEvent{Type: "message", Data: Message{
			Author:  html.EscapeString(sender.Name),
			Content: escaped,
		}})
	}
}

func sendMessageToAllNonGuessing(message string, sender *Player, lobby *Lobby) {
	escaped := html.EscapeString(discordemojimap.Replace(message))
	for _, target := range lobby.Players {
		if target.State != Guessing {
			WriteAsJSON(target, JSEvent{Type: "non-guessing-player-message", Data: Message{
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

		votesNeeded := calculateVotesNeededToKick(len(lobby.Players))

		WritePublicSystemMessage(lobby, fmt.Sprintf("(%d/%d) players voted to kick %s", voteKickCount, votesNeeded, playerToKick.Name))

		if voteKickCount >= votesNeeded {
			//Since the player is already kicked, we first clean up the kicking information related to that player
			for _, otherPlayer := range lobby.Players {
				if otherPlayer.votedForKick[toKickID] == true {
					delete(player.votedForKick, toKickID)
					break
				}
			}

			WritePublicSystemMessage(lobby, fmt.Sprintf("%s has been kicked from the lobby", playerToKick.Name))

			if lobby.Drawer == playerToKick {
				WritePublicSystemMessage(lobby, "Since the kicked player has been drawing, none of you will get any points this round.")
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
				playerToKick.ws.Close()
			}
			lobby.Players = append(lobby.Players[:toKick], lobby.Players[toKick+1:]...)

			recalculateRanks(lobby)

			//If the owner is kicked, we choose the next best person as the owner.
			if lobby.Owner == playerToKick {
				for _, otherPlayer := range lobby.Players {
					potentialOwner := otherPlayer
					if potentialOwner.Connected {
						lobby.Owner = potentialOwner
						WritePublicSystemMessage(lobby, fmt.Sprintf("%s is the new lobby owner.", potentialOwner.Name))
						break
					}
				}
			}

			triggerPlayersUpdate(lobby)

			if lobby.Drawer == playerToKick || !lobby.isAnyoneStillGuessing() {
				endTurn(lobby)
			}
		}
	}
}

func calculateVotesNeededToKick(amountOfPlayers int) int {
	//If the amount of players equals an even number, such as 6, we will always
	//need half of that. If the amount is uneven, we'll get a floored result.
	//therefore we always add one to the amount.
	//examples:
	//    (6+1)/2 = 3
	//    (5+1)/2 = 3
	//Therefore it'll never be possible for a minority to kick a player.
	return (amountOfPlayers + 1) / 2
}

func handleCommand(commandString string, caller *Player, lobby *Lobby) {
	command := commands.ParseCommand(commandString)
	if len(command) >= 1 {
		switch strings.ToLower(command[0]) {
		case "setmp":
			commandSetMP(caller, lobby, command)
		case "help":
			//TODO
		}
	}
}

func commandNick(caller *Player, lobby *Lobby, name string) {
	newName := html.EscapeString(strings.TrimSpace(name))

	//We don't want super-long names
	if len(newName) > 30 {
		newName = newName[:31]
	}

	if newName == "" {
		caller.Name = GeneratePlayerName()
	} else {
		caller.Name = newName
	}

	fmt.Printf("%s is now %s\n", caller.Name, newName)

	triggerPlayersUpdate(lobby)
}

func commandSetMP(caller *Player, lobby *Lobby, args []string) {
	if caller == lobby.Owner {
		if len(args) < 2 {
			return
		}

		newMaxPlayersValue := strings.TrimSpace(args[1])
		newMaxPlayersValueInt, err := strconv.ParseInt(newMaxPlayersValue, 10, 64)
		if err == nil {
			if int(newMaxPlayersValueInt) >= len(lobby.Players) && newMaxPlayersValueInt <= LobbySettingBounds.MaxMaxPlayers && newMaxPlayersValueInt >= LobbySettingBounds.MinMaxPlayers {
				lobby.MaxPlayers = int(newMaxPlayersValueInt)

				WritePublicSystemMessage(lobby, fmt.Sprintf("MaxPlayers value has been changed to %d", lobby.MaxPlayers))
			} else {
				if len(lobby.Players) > int(LobbySettingBounds.MinMaxPlayers) {
					WriteAsJSON(caller, JSEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value should be between %d and %d.", len(lobby.Players), LobbySettingBounds.MaxMaxPlayers)})
				} else {
					WriteAsJSON(caller, JSEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value should be between %d and %d.", LobbySettingBounds.MinMaxPlayers, LobbySettingBounds.MaxMaxPlayers)})
				}
			}
		} else {
			WriteAsJSON(caller, JSEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value must be numeric.")})
		}
	} else {
		WriteAsJSON(caller, JSEvent{Type: "system-message", Data: fmt.Sprintf("Only the lobby owner can change MaxPlayers setting.")})
	}
}

func endTurn(lobby *Lobby) {
	if lobby.timeLeftTicker != nil {
		lobby.timeLeftTicker.Stop()
		lobby.timeLeftTicker = nil
		lobby.timeLeftTickerReset <- struct{}{}
	}

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

	WritePublicSystemMessage(lobby, roundOverMessage)

	advanceLobby(lobby)
}

// advanceLobby will either start the game or jump over to the next turn.
func advanceLobby(lobby *Lobby) {
	for _, otherPlayer := range lobby.Players {
		otherPlayer.State = Guessing
		otherPlayer.votedForKick = make(map[string]bool)
	}

	lobby.ClearDrawing()

	newDrawer, roundOver := selectNextDrawer(lobby)
	if roundOver {
		if lobby.Round == lobby.MaxRounds {
			endGame(lobby)
			return
		}

		lobby.Round++
	}

	lobby.Drawer = newDrawer
	lobby.Drawer.State = Drawing
	lobby.WordChoice = GetRandomWords(lobby)

	recalculateRanks(lobby)

	//We use milliseconds for higher accuracy
	lobby.RoundEndTime = time.Now().UTC().UnixNano()/1000000 + int64(lobby.DrawingTime)*1000
	lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
	go roundTimerTicker(lobby)

	TriggerComplexUpdateEvent("next-turn", &NextTurn{
		Round:        lobby.Round,
		Players:      lobby.Players,
		RoundEndTime: int(lobby.RoundEndTime - getTimeAsMillis()),
	}, lobby)

	WriteAsJSON(lobby.Drawer, &JSEvent{Type: "your-turn", Data: lobby.WordChoice})
}

func endGame(lobby *Lobby) {
	lobby.Drawer = nil
	lobby.Round = 0

	recalculateRanks(lobby)
	triggerPlayersUpdate(lobby)

	WritePublicSystemMessage(lobby, "Game over. Type !start again to start a new round.")
}

// selectNextDrawer returns the next person that's supposed to be drawing, but
// doesn't tell the lobby yet. The boolean signals whether the current round is
// over.
func selectNextDrawer(lobby *Lobby) (*Player, bool) {
	for index, otherPlayer := range lobby.Players {
		if otherPlayer == lobby.Drawer {
			//If we have someone that's drawing, take the next one
			for i := index + 1; i < len(lobby.Players); i++ {
				player := lobby.Players[i]
				if player.Connected {
					return player, false
				}
			}
		}
	}

	return lobby.Players[0], true
}

func roundTimerTicker(lobby *Lobby) {
	hintsLeft := 2
	revealHintAtMillisecondsLeft := lobby.DrawingTime * 1000 / 3

	for {
		select {
		case <-lobby.timeLeftTicker.C:
			currentTime := getTimeAsMillis()
			if currentTime >= lobby.RoundEndTime {
				go endTurn(lobby)
			}

			if hintsLeft > 0 && lobby.WordHints != nil {
				timeLeft := lobby.RoundEndTime - currentTime
				if timeLeft <= int64(revealHintAtMillisecondsLeft*hintsLeft) {
					hintsLeft--

					for {
						randomIndex := rand.Int() % len(lobby.WordHints)
						if lobby.WordHints[randomIndex].Character == 0 {
							lobby.WordHints[randomIndex].Character = []rune(lobby.CurrentWord)[randomIndex]
							triggerWordHintUpdate(lobby)
							break
						}
					}
				}
			}
		case <-lobby.timeLeftTickerReset:
			return
		}
	}
}

func getTimeAsMillis() int64 {
	return time.Now().UTC().UnixNano() / 1000000
}

// NextTurn represents the data necessary for displaying the lobby state right
// after a new turn started. Meaning that no word has been chosen yet and
// therefore there are no wordhints and no current drawing instructions.
type NextTurn struct {
	Round        int       `json:"round"`
	Players      []*Player `json:"players"`
	RoundEndTime int       `json:"roundEndTime"`
}

// recalculateRanks will assign each player his respective rank in the lobby
// according to everyones current score. This will not trigger any events.
func recalculateRanks(lobby *Lobby) {
	for _, a := range lobby.Players {
		if !a.Connected {
			continue
		}
		playersThatAreHigher := 0
		for _, b := range lobby.Players {
			if !b.Connected {
				continue
			}
			if b.Score > a.Score {
				playersThatAreHigher++
			}
		}

		a.Rank = playersThatAreHigher + 1
	}
}

func createWordHintFor(word string, showAll bool) []*WordHint {
	wordHints := make([]*WordHint, 0, len(word))
	for _, char := range word {
		irrelevantChar := char == ' ' || char == '_' || char == '-'
		if showAll {
			wordHints = append(wordHints, &WordHint{
				Character: char,
				Underline: !irrelevantChar,
			})
		} else {
			if irrelevantChar {
				wordHints = append(wordHints, &WordHint{
					Character: char,
					Underline: !irrelevantChar,
				})
			} else {
				wordHints = append(wordHints, &WordHint{
					Underline: !irrelevantChar,
				})
			}
		}
	}

	return wordHints
}

var TriggerSimpleUpdateEvent func(eventType string, lobby *Lobby)
var TriggerComplexUpdatePerPlayerEvent func(eventType string, data func(*Player) interface{}, lobby *Lobby)
var TriggerComplexUpdateEvent func(eventType string, data interface{}, lobby *Lobby)
var SendDataToConnectedPlayers func(sender *Player, lobby *Lobby, data interface{})
var WriteAsJSON func(player *Player, object interface{}) error
var WritePublicSystemMessage func(lobby *Lobby, text string)

func triggerPlayersUpdate(lobby *Lobby) {
	TriggerComplexUpdateEvent("update-players", lobby.Players, lobby)
}

func triggerCorrectGuessEvent(lobby *Lobby) {
	TriggerSimpleUpdateEvent("correct-guess", lobby)
}

func triggerWordHintUpdate(lobby *Lobby) {
	if lobby.CurrentWord == "" {
		return
	}

	TriggerComplexUpdatePerPlayerEvent("update-wordhint", func(player *Player) interface{} {
		return lobby.GetAvailableWordHints(player)
	}, lobby)
}

type Rounds struct {
	Round     int `json:"round"`
	MaxRounds int `json:"maxRounds"`
}

// CreateLobby allows creating a lobby, optionally returning errors that
// occurred during creation.
func CreateLobby(playerName, language string, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit int, customWords []string, enableVotekick bool) (*Player, *Lobby, error) {
	lobby := createLobby(drawingTime, rounds, maxPlayers, customWords, customWordChance, clientsPerIPLimit, enableVotekick)
	player := createPlayer(playerName)

	lobby.Players = append(lobby.Players, player)
	lobby.Owner = player

	// Read wordlist according to the chosen language
	words, err := readWordList(language)
	if err != nil {
		RemoveLobby(lobby.ID)
		return nil, nil, err
	}

	lobby.Words = words

	return player, lobby, nil
}

// GeneratePlayerName creates a new playername. A so called petname. It consists
// of an adverb, an adjective and a animal name. The result can generally be
// trusted to be sane.
func GeneratePlayerName() string {
	adjective := strings.Title(petname.Adjective())
	adverb := strings.Title(petname.Adverb())
	name := strings.Title(petname.Name())
	return adverb + adjective + name
}

// Message represents a message in the chatroom.
type Message struct {
	// Author is the player / thing that wrote the message
	Author string `json:"author"`
	// Content is the actual message text.
	Content string `json:"content"`
}

// Ready represents the initial state that a user needs upon connection.
// This includes all the necessary things for properly running a client
// without receiving any more data.
type Ready struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Drawing    bool   `json:"drawing"`

	OwnerID        string        `json:"ownerId"`
	Round          int           `json:"round"`
	MaxRound       int           `json:"maxRounds"`
	RoundEndTime   int           `json:"roundEndTime"`
	WordHints      []*WordHint   `json:"wordHints"`
	Players        []*Player     `json:"players"`
	CurrentDrawing []interface{} `json:"currentDrawing"`
}

func OnConnected(lobby *Lobby, player *Player) {
	player.Connected = true
	WriteAsJSON(player, JSEvent{Type: "ready", Data: &Ready{
		PlayerID:   player.ID,
		Drawing:    player.State == Drawing,
		PlayerName: player.Name,

		OwnerID:        lobby.Owner.ID,
		Round:          lobby.Round,
		MaxRound:       lobby.MaxRounds,
		RoundEndTime:   int(lobby.RoundEndTime - getTimeAsMillis()),
		WordHints:      lobby.GetAvailableWordHints(player),
		Players:        lobby.Players,
		CurrentDrawing: lobby.CurrentDrawing,
	}})

	//This state is reached when the player refreshes before having chosen a word.
	if lobby.Drawer == player && lobby.CurrentWord == "" {
		WriteAsJSON(lobby.Drawer, &JSEvent{Type: "your-turn", Data: lobby.WordChoice})
	}

	//TODO Only send to everyone except for the new player, since it's part of the ready event.
	triggerPlayersUpdate(lobby)
}

func OnDisconnected(lobby *Lobby, player *Player) {
	//We want to avoid calling the handler twice.
	if player.ws == nil {
		return
	}

	player.Connected = false
	player.ws = nil

	if !lobby.HasConnectedPlayers() {
		RemoveLobby(lobby.ID)
		log.Printf("Closing lobby %s. There are currently %d open lobbies left.\n", lobby.ID, len(lobbies))
	} else {
		triggerPlayersUpdate(lobby)
	}
}

func (lobby *Lobby) GetAvailableWordHints(player *Player) []*WordHint {
	//The draw simple gets every character as a word-hint. We basically abuse
	//the hints for displaying the word, instead of having yet another GUI
	//element that wastes space.
	if player.State == Drawing || player.State == Standby {
		return lobby.WordHintsShown
	} else {
		return lobby.WordHints
	}
}

func (lobby *Lobby) JoinPlayer(playerName string) *Player {
	player := createPlayer(playerName)

	//FIXME Make a dedicated method that uses a mutex?
	lobby.Players = append(lobby.Players, player)
	recalculateRanks(lobby)
	triggerPlayersUpdate(lobby)

	return player
}

func (lobby *Lobby) canDraw(player *Player) bool {
	return lobby.Drawer == player && lobby.CurrentWord != ""
}

func removeAccents(s string) string {
	return sanitize.Accents(s)
}
