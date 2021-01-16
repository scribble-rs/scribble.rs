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
	"time"
	"unicode/utf8"

	commands "github.com/Bios-Marcel/cmdp"
	"github.com/Bios-Marcel/discordemojimap"
	"github.com/agnivade/levenshtein"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/kennygrant/sanitize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// LineEvent is basically the same as GameEvent, but with a specific Data type.
// We use this for reparsing as soon as we know that the type is right. It's
// a bit unperformant, but will do for now.
type LineEvent struct {
	Type string `json:"type"`
	Data *Line  `json:"data"`
}

// FillEvent is basically the same as GameEvent, but with a specific Data type.
// We use this for reparsing as soon as we know that the type is right. It's
// a bit unperformant, but will do for now.
type FillEvent struct {
	Type string `json:"type"`
	Data *Fill  `json:"data"`
}

type KickVote struct {
	PlayerID          string `json:"playerId"`
	PlayerName        string `json:"playerName"`
	VoteCount         int    `json:"voteCount"`
	RequiredVoteCount int    `json:"requiredVoteCount"`
}

func HandleEvent(raw []byte, received *GameEvent, lobby *Lobby, player *Player) error {
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

			//In case the line is too big, we overwrite the data of the event.
			//This will prevent clients from lagging due to too thick lines.
			if line.Data.LineWidth > 40.0 {
				line.Data.LineWidth = 40.0
				received.Data = line
			}

			lobby.AppendLine(line)

			//We directly forward the event, as it seems to be valid.
			SendDataToEveryoneExceptSender(player, lobby, received)
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
			SendDataToEveryoneExceptSender(player, lobby, received)
		}
	} else if received.Type == "clear-drawing-board" {
		if lobby.canDraw(player) && len(lobby.currentDrawing) > 0 {
			lobby.ClearDrawing()
			SendDataToEveryoneExceptSender(player, lobby, received)
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

		drawer := lobby.drawer
		if player == drawer && len(lobby.wordChoice) > 0 && chosenIndex >= 0 && chosenIndex <= 2 {
			lobby.CurrentWord = lobby.wordChoice[chosenIndex]

			//Depending on how long the word is, a fixed amount of hints
			//would be too easy or too hard.
			runeCount := utf8.RuneCountInString(lobby.CurrentWord)
			if runeCount <= 2 {
				lobby.hintCount = 0
			} else if runeCount <= 4 {
				lobby.hintCount = 1
			} else if runeCount <= 9 {
				lobby.hintCount = 2
			} else {
				lobby.hintCount = 3
			}
			lobby.hintsLeft = lobby.hintCount

			lobby.wordChoice = nil
			lobby.wordHints = createWordHintFor(lobby.CurrentWord, false)
			lobby.wordHintsShown = createWordHintFor(lobby.CurrentWord, true)
			triggerWordHintUpdate(lobby)
		}
	} else if received.Type == "kick-vote" {
		if lobby.EnableVotekick {
			toKickID, isString := (received.Data).(string)
			if !isString {
				return fmt.Errorf("invalid data in kick-vote event: %v", received.Data)
			}

			handleKickEvent(lobby, player, toKickID)
		}
	} else if received.Type == "start" {
		if lobby.Round == 0 && player == lobby.owner {
			//We are reseting each players score, since players could
			//technically be player a second game after the last one
			//has already ended.
			for _, otherPlayer := range lobby.players {
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
	} else if received.Type == "request-drawing" {
		WriteAsJSON(player, GameEvent{Type: "drawing", Data: lobby.currentDrawing})
	} else if received.Type == "keep-alive" {
		//This is a known dummy event in order to avoid accidental websocket
		//connection closure. However, no action is required on the server.
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
		lowerCasedInput := lobby.lowercaser.String(trimmed)
		currentWord := lobby.CurrentWord

		normInput := removeAccents(lowerCasedInput)
		normSearched := removeAccents(currentWord)

		if normSearched == normInput {
			secondsLeft := lobby.RoundEndTime/1000 - time.Now().UTC().UnixNano()/1000000000

			sender.LastScore = int(math.Ceil(math.Pow(math.Max(float64(secondsLeft), 1), 1.3) * 2))
			sender.Score += sender.LastScore
			lobby.scoreEarnedByGuessers += sender.LastScore
			sender.State = Standby

			TriggerUpdateEvent("correct-guess", sender.ID, lobby)

			if !lobby.isAnyoneStillGuessing() {
				advanceLobby(lobby)
			} else {
				//Since the word has been guessed correctly, we reveal it.
				WriteAsJSON(sender, GameEvent{Type: "update-wordhint", Data: lobby.wordHintsShown})
				recalculateRanks(lobby)
				triggerPlayersUpdate(lobby)
			}
		} else if levenshtein.ComputeDistance(normInput, normSearched) == 1 {
			WriteAsJSON(sender, GameEvent{Type: "close-guess", Data: trimmed})
			//In cases of a close guess, we still send the message to everyone.
			//This allows other players to guess the word by watching what the
			//other players are misstyping.
			sendMessageToAll(trimmed, sender, lobby)
		} else {
			sendMessageToAll(trimmed, sender, lobby)
		}
	}
}

func (lobby *Lobby) isAnyoneStillGuessing() bool {
	for _, otherPlayer := range lobby.players {
		if otherPlayer.State == Guessing && otherPlayer.Connected {
			return true
		}
	}

	return false
}

func sendMessageToAll(message string, sender *Player, lobby *Lobby) {
	escaped := html.EscapeString(discordemojimap.Replace(message))
	for _, target := range lobby.players {
		WriteAsJSON(target, GameEvent{Type: "message", Data: Message{
			Author:   html.EscapeString(sender.Name),
			AuthorID: sender.ID,
			Content:  escaped,
		}})
	}
}

func sendMessageToAllNonGuessing(message string, sender *Player, lobby *Lobby) {
	escaped := html.EscapeString(discordemojimap.Replace(message))
	for _, target := range lobby.players {
		if target.State != Guessing {
			WriteAsJSON(target, GameEvent{Type: "non-guessing-player-message", Data: Message{
				Author:   html.EscapeString(sender.Name),
				AuthorID: sender.ID,
				Content:  escaped,
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
	for index, otherPlayer := range lobby.players {
		if otherPlayer.ID == toKickID {
			toKick = index
			break
		}
	}

	//If we haven't found the player, we can't kick him/her.
	if toKick != -1 {
		player.votedForKick[toKickID] = true
		playerToKick := lobby.players[toKick]

		var voteKickCount int
		for _, otherPlayer := range lobby.players {
			if otherPlayer.votedForKick[toKickID] == true {
				voteKickCount++
			}
		}

		votesNeeded := calculateVotesNeededToKick(len(lobby.players))

		kickEvent := &GameEvent{
			Type: "kick-vote",
			Data: &KickVote{
				PlayerID:          playerToKick.ID,
				PlayerName:        playerToKick.Name,
				VoteCount:         voteKickCount,
				RequiredVoteCount: votesNeeded,
			},
		}

		for _, otherPlayer := range lobby.players {
			WriteAsJSON(otherPlayer, kickEvent)
		}

		if voteKickCount >= votesNeeded {
			//Since the player is already kicked, we first clean up the kicking information related to that player
			for _, otherPlayer := range lobby.players {
				if otherPlayer.votedForKick[toKickID] == true {
					delete(player.votedForKick, toKickID)
					break
				}
			}

			if lobby.drawer == playerToKick {
				WritePublicSystemMessage(lobby, "Since the kicked player has been drawing, none of you will get any points this round.")
				//Since the drawing person has been kicked, that probably means that he/she was trolling, therefore
				//we redact everyones last earned score.
				for _, otherPlayer := range lobby.players {
					otherPlayer.Score -= otherPlayer.LastScore
					otherPlayer.LastScore = 0
				}
				lobby.scoreEarnedByGuessers = 0
				//We must absolutely not set lobby.drawer to nil, since this would cause the drawing order to be ruined.
			}

			if playerToKick.ws != nil {
				playerToKick.ws.Close()
			}
			lobby.players = append(lobby.players[:toKick], lobby.players[toKick+1:]...)

			recalculateRanks(lobby)

			//If the owner is kicked, we choose the next best person as the owner.
			if lobby.owner == playerToKick {
				for _, otherPlayer := range lobby.players {
					potentialOwner := otherPlayer
					if potentialOwner.Connected {
						lobby.owner = potentialOwner
						TriggerUpdateEvent("owner-change", &OwnerChangeEvent{
							PlayerID:   potentialOwner.ID,
							PlayerName: potentialOwner.Name,
						}, lobby)
						break
					}
				}
			}

			triggerPlayersUpdate(lobby)

			if lobby.drawer == playerToKick || !lobby.isAnyoneStillGuessing() {
				advanceLobby(lobby)
			}
		}
	}
}

type OwnerChangeEvent struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
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
	if caller == lobby.owner {
		if len(args) < 2 {
			return
		}

		newMaxPlayersValue := strings.TrimSpace(args[1])
		newMaxPlayersValueInt, err := strconv.ParseInt(newMaxPlayersValue, 10, 64)
		if err == nil {
			if int(newMaxPlayersValueInt) >= len(lobby.players) && newMaxPlayersValueInt <= LobbySettingBounds.MaxMaxPlayers && newMaxPlayersValueInt >= LobbySettingBounds.MinMaxPlayers {
				lobby.MaxPlayers = int(newMaxPlayersValueInt)

				WritePublicSystemMessage(lobby, fmt.Sprintf("MaxPlayers value has been changed to %d", lobby.MaxPlayers))
			} else {
				if len(lobby.players) > int(LobbySettingBounds.MinMaxPlayers) {
					WriteAsJSON(caller, GameEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value should be between %d and %d.", len(lobby.players), LobbySettingBounds.MaxMaxPlayers)})
				} else {
					WriteAsJSON(caller, GameEvent{Type: "system-message", Data: fmt.Sprintf("MaxPlayers value should be between %d and %d.", LobbySettingBounds.MinMaxPlayers, LobbySettingBounds.MaxMaxPlayers)})
				}
			}
		} else {
			WriteAsJSON(caller, GameEvent{Type: "system-message", Data: "MaxPlayers value must be numeric."})
		}
	} else {
		WriteAsJSON(caller, GameEvent{Type: "system-message", Data: "Only the lobby owner can change MaxPlayers setting."})
	}
}

// advanceLobby will either start the game or jump over to the next turn.
func advanceLobby(lobby *Lobby) {
	if lobby.timeLeftTicker != nil {
		lobby.timeLeftTicker.Stop()
		lobby.timeLeftTicker = nil
	}

	//The drawer can potentially be null if he's kicked, in that case we proceed with the round if anyone has already
	drawer := lobby.drawer
	if drawer != nil && lobby.scoreEarnedByGuessers > 0 {
		averageScore := float64(lobby.scoreEarnedByGuessers) / float64(len(lobby.players)-1)
		if averageScore > 0 {
			drawer.LastScore = int(averageScore * 1.1)
			drawer.Score += drawer.LastScore
		}
	}

	//We need this for the next-turn event, in order to allow the client
	//to know which word was previously supposed to be guessed.
	previousWord := lobby.CurrentWord

	lobby.scoreEarnedByGuessers = 0
	lobby.CurrentWord = ""
	lobby.wordHints = nil

	//If the round ends and people still have guessing, that means the "Last" value
	//for the next turn has to be "no score earned".
	for _, otherPlayer := range lobby.players {
		if otherPlayer.State == Guessing {
			otherPlayer.LastScore = 0
		}
	}

	for _, otherPlayer := range lobby.players {
		otherPlayer.State = Guessing
		otherPlayer.votedForKick = make(map[string]bool)
	}

	newDrawer, roundOver := selectNextDrawer(lobby)
	if roundOver {
		if lobby.Round == lobby.MaxRounds {
			endGame(lobby)
			return
		}

		lobby.Round++
	}

	firstTurn := lobby.state != ongoing

	lobby.ClearDrawing()
	lobby.drawer = newDrawer
	lobby.drawer.State = Drawing
	lobby.state = ongoing
	lobby.wordChoice = GetRandomWords(3, lobby)

	recalculateRanks(lobby)

	//We use milliseconds for higher accuracy
	lobby.RoundEndTime = time.Now().UTC().UnixNano()/1000000 + int64(lobby.DrawingTime)*1000
	lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
	go roundTimerTicker(lobby)

	nextTurnEvent := &NextTurn{
		Round:        lobby.Round,
		Players:      lobby.players,
		RoundEndTime: int(lobby.RoundEndTime - getTimeAsMillis()),
	}

	//In the first turn, we set this field to null to signal that
	//there hasn't been no choice, but no turn before this round.
	//Meaning that for the API user empty string and nil are to be
	//treated with a different meaning.
	if !firstTurn {
		nextTurnEvent.PreviousWord = &previousWord
	}
	TriggerUpdateEvent("next-turn", nextTurnEvent, lobby)

	WriteAsJSON(lobby.drawer, &GameEvent{Type: "your-turn", Data: lobby.wordChoice})
}

func endGame(lobby *Lobby) {
	lobby.drawer = nil
	lobby.Round = 0
	lobby.state = gameOver

	recalculateRanks(lobby)

	for _, player := range lobby.players {
		WriteAsJSON(player, GameEvent{
			Type: "ready",
			Data: generateReadyData(lobby, player),
		})
	}
}

// selectNextDrawer returns the next person that's supposed to be drawing, but
// doesn't tell the lobby yet. The boolean signals whether the current round is
// over.
func selectNextDrawer(lobby *Lobby) (*Player, bool) {
	for index, otherPlayer := range lobby.players {
		if otherPlayer == lobby.drawer {
			//If we have someone that's drawing, take the next one
			for i := index + 1; i < len(lobby.players); i++ {
				player := lobby.players[i]
				if player.Connected {
					return player, false
				}
			}
		}
	}

	return lobby.players[0], true
}

func roundTimerTicker(lobby *Lobby) {
	for {
		ticker := lobby.timeLeftTicker
		if ticker == nil {
			return
		}

		select {
		case <-ticker.C:
			currentTime := getTimeAsMillis()
			if currentTime >= lobby.RoundEndTime {
				go advanceLobby(lobby)
			}

			if lobby.hintsLeft > 0 && lobby.wordHints != nil {
				revealHintEveryXMilliseconds := int64(lobby.DrawingTime * 1000 / (lobby.hintCount + 1))
				//If you have a drawingtime of 120 seconds and three hints, you
				//want to reveal a hint every 40 seconds, so that the two hints
				//are visible for at least a third of the time. //If the word
				//was chosen at 60 seconds, we'll still reveal one hint
				//instantly, as the time is already lower than 80.
				revealHintAtXOrLower := revealHintEveryXMilliseconds * int64(lobby.hintsLeft)
				timeLeft := lobby.RoundEndTime - currentTime
				if timeLeft <= revealHintAtXOrLower {
					lobby.hintsLeft--

					for {
						randomIndex := rand.Int() % len(lobby.wordHints)
						if lobby.wordHints[randomIndex].Character == 0 {
							lobby.wordHints[randomIndex].Character = []rune(lobby.CurrentWord)[randomIndex]
							triggerWordHintUpdate(lobby)
							break
						}
					}
				}
			}
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
	PreviousWord *string   `json:"previousWord"`
}

// recalculateRanks will assign each player his respective rank in the lobby
// according to everyones current score. This will not trigger any events.
func recalculateRanks(lobby *Lobby) {
	for _, a := range lobby.players {
		if !a.Connected {
			continue
		}
		playersThatAreHigher := 0
		for _, b := range lobby.players {
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

var TriggerUpdatePerPlayerEvent func(eventType string, data func(*Player) interface{}, lobby *Lobby)
var TriggerUpdateEvent func(eventType string, data interface{}, lobby *Lobby)
var SendDataToEveryoneExceptSender func(sender *Player, lobby *Lobby, data interface{})
var WriteAsJSON func(player *Player, object interface{}) error
var WritePublicSystemMessage func(lobby *Lobby, text string)

func triggerPlayersUpdate(lobby *Lobby) {
	TriggerUpdateEvent("update-players", lobby.players, lobby)
}

func triggerWordHintUpdate(lobby *Lobby) {
	if lobby.CurrentWord == "" {
		return
	}

	TriggerUpdatePerPlayerEvent("update-wordhint", func(player *Player) interface{} {
		return lobby.GetAvailableWordHints(player)
	}, lobby)
}

type Rounds struct {
	Round     int `json:"round"`
	MaxRounds int `json:"maxRounds"`
}

// CreateLobby allows creating a lobby, optionally returning errors that
// occurred during creation.
func CreateLobby(playerName, chosenLanguage string, publicLobby bool, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit int, customWords []string, enableVotekick bool) (*Player, *Lobby, error) {
	lobby := createLobby(drawingTime, rounds, maxPlayers, customWords, customWordChance, clientsPerIPLimit, enableVotekick)
	lobby.Wordpack = chosenLanguage
	lobby.public = publicLobby

	//Neccessary to correctly treat words from player, however, custom words might be treated incorrectly.
	lobby.lowercaser = cases.Lower(language.Make(getLanguageIdentifier(chosenLanguage)))

	//customWords are lowercased afterwards, as they are direct user input.
	if len(customWords) > 0 {
		for customWordIndex, customWord := range customWords {
			customWords[customWordIndex] = lobby.lowercaser.String(customWord)
		}
	}

	player := createPlayer(playerName)

	lobby.players = append(lobby.players, player)
	lobby.owner = player

	words, err := readWordList(lobby.lowercaser, chosenLanguage)
	if err != nil {
		return nil, nil, err
	}

	lobby.words = words

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
	// AuthorID is the unique identifier of the authors player object.
	AuthorID string `json:"authorId"`
	// Content is the actual message text.
	Content string `json:"content"`
}

// Ready represents the initial state that a user needs upon connection.
// This includes all the necessary things for properly running a client
// without receiving any more data.
type Ready struct {
	PlayerID     string `json:"playerId"`
	PlayerName   string `json:"playerName"`
	AllowDrawing bool   `json:"allowDrawing"`

	VotekickEnabled bool          `json:"votekickEnabled"`
	GameState       gameState     `json:"gameState"`
	OwnerID         string        `json:"ownerId"`
	Round           int           `json:"round"`
	MaxRound        int           `json:"maxRounds"`
	RoundEndTime    int           `json:"roundEndTime"`
	WordHints       []*WordHint   `json:"wordHints"`
	Players         []*Player     `json:"players"`
	CurrentDrawing  []interface{} `json:"currentDrawing"`
}

func generateReadyData(lobby *Lobby, player *Player) *Ready {
	ready := &Ready{
		PlayerID:     player.ID,
		AllowDrawing: player.State == Drawing,
		PlayerName:   player.Name,

		VotekickEnabled: lobby.EnableVotekick,
		GameState:       lobby.state,
		OwnerID:         lobby.owner.ID,
		Round:           lobby.Round,
		MaxRound:        lobby.MaxRounds,
		WordHints:       lobby.GetAvailableWordHints(player),
		Players:         lobby.players,
		CurrentDrawing:  lobby.currentDrawing,
	}

	//Game over already
	if lobby.state != ongoing {
		//0 is interpreted as "no time left".
		ready.RoundEndTime = 0
	} else {
		ready.RoundEndTime = int(lobby.RoundEndTime - getTimeAsMillis())
	}

	return ready
}

func OnConnected(lobby *Lobby, player *Player) {
	player.Connected = true
	WriteAsJSON(player, GameEvent{Type: "ready", Data: generateReadyData(lobby, player)})

	//This state is reached when the player refreshes before having chosen a word.
	if lobby.drawer == player && lobby.CurrentWord == "" {
		WriteAsJSON(lobby.drawer, &GameEvent{Type: "your-turn", Data: lobby.wordChoice})
	}

	updateRocketChat(lobby, player)

	//TODO Only send to everyone except for the new player, since it's part of the ready event.
	triggerPlayersUpdate(lobby)
}

func OnDisconnected(lobby *Lobby, player *Player) {
	//We want to avoid calling the handler twice.
	if player.ws == nil {
		return
	}

	log.Printf("Player %s(%s) disconnected.\n", player.Name, player.ID)
	player.Connected = false
	player.ws = nil

	disconnectTime := time.Now()
	lobby.LastPlayerDisconnectTime = &disconnectTime

	updateRocketChat(lobby, player)

	if lobby.HasConnectedPlayers() {
		triggerPlayersUpdate(lobby)
	}
}

func (lobby *Lobby) GetAvailableWordHints(player *Player) []*WordHint {
	//The draw simple gets every character as a word-hint. We basically abuse
	//the hints for displaying the word, instead of having yet another GUI
	//element that wastes space.
	if player.State == Drawing || player.State == Standby {
		return lobby.wordHintsShown
	} else {
		return lobby.wordHints
	}
}

// JoinPlayer creates a new player object using the given name and adds it
// to the lobbies playerlist. The new players is returned.
func (lobby *Lobby) JoinPlayer(playerName string) *Player {
	player := createPlayer(playerName)

	//FIXME Make a dedicated method that uses a mutex?
	lobby.players = append(lobby.players, player)

	return player
}

func (lobby *Lobby) canDraw(player *Player) bool {
	return lobby.drawer == player && lobby.CurrentWord != ""
}

func removeAccents(s string) string {
	return strings.
		NewReplacer(" ", "", "-", "", "_", "").
		Replace(sanitize.Accents(s))
}
