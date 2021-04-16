package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	discordemojimap "github.com/Bios-Marcel/discordemojimap/v2"
	"github.com/agnivade/levenshtein"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/gofrs/uuid"
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
		"english_gb": "English (GB)",
		"english":    "English (US)",
		"italian":    "Italian",
		"german":     "German",
		"french":     "French",
		"dutch":      "Dutch",
	}
)

const (
	DrawingBoardBaseWidth  = 1600
	DrawingBoardBaseHeight = 900
	MinBrushSize           = 8
	MaxBrushSize           = 32

	maxBaseScore      = 200
	maxHintBonusScore = 60
)

// SettingBounds defines the lower and upper bounds for the user-specified
// lobby creation input.
type SettingBounds struct {
	MinDrawingTime       int64 `json:"minDrawingTime"`
	MaxDrawingTime       int64 `json:"maxDrawingTime"`
	MinRounds            int64 `json:"minRounds"`
	MaxRounds            int64 `json:"maxRounds"`
	MinMaxPlayers        int64 `json:"minMaxPlayers"`
	MaxMaxPlayers        int64 `json:"maxMaxPlayers"`
	MinClientsPerIPLimit int64 `json:"minClientsPerIpLimit"`
	MaxClientsPerIPLimit int64 `json:"maxClientsPerIpLimit"`
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

// KickVote represents a players vote to kick another players. If the VoteCount
// is as great or greater than the RequiredVoteCount, the event indicates a
// successful kick vote. The voting is anonymous, meaning the voting player
// won't be exposed.
type KickVote struct {
	PlayerID          string `json:"playerId"`
	PlayerName        string `json:"playerName"`
	VoteCount         int    `json:"voteCount"`
	RequiredVoteCount int    `json:"requiredVoteCount"`
}

func (lobby *Lobby) HandleEvent(raw []byte, received *GameEvent, player *Player) error {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	if received.Type == "message" {
		dataAsString, isString := (received.Data).(string)
		if !isString {
			return fmt.Errorf("invalid data received: '%s'", received.Data)
		}

		handleMessage(dataAsString, player, lobby)
	} else if received.Type == "line" {
		if lobby.canDraw(player) {
			line := &LineEvent{}
			jsonError := json.Unmarshal(raw, line)
			if jsonError != nil {
				return fmt.Errorf("error decoding data: %s", jsonError)
			}

			//In case the line is too big, we overwrite the data of the event.
			//This will prevent clients from lagging due to too thick lines.
			if line.Data.LineWidth > float32(MaxBrushSize) {
				line.Data.LineWidth = MaxBrushSize
				received.Data = line.Data
			} else if line.Data.LineWidth < float32(MinBrushSize) {
				line.Data.LineWidth = MinBrushSize
				received.Data = line.Data
			}

			lobby.AppendLine(line)

			//We directly forward the event, as it seems to be valid.
			lobby.sendDataToEveryoneExceptSender(player, received)
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
			lobby.sendDataToEveryoneExceptSender(player, received)
		}
	} else if received.Type == "clear-drawing-board" {
		if lobby.canDraw(player) && len(lobby.currentDrawing) > 0 {
			lobby.ClearDrawing()
			lobby.sendDataToEveryoneExceptSender(player, received)
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
			lobby.triggerWordHintUpdate()
		}
	} else if received.Type == "kick-vote" {
		if lobby.EnableVotekick {
			toKickID, isString := (received.Data).(string)
			if !isString {
				return fmt.Errorf("invalid data in kick-vote event: %v", received.Data)
			}

			handleKickVoteEvent(lobby, player, toKickID)
		}
	} else if received.Type == "start" {
		if lobby.Round == 0 && player == lobby.Owner {
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
		handleNameChangeEvent(player, lobby, newName)
	} else if received.Type == "request-drawing" {
		lobby.WriteJSON(player, GameEvent{Type: "drawing", Data: lobby.currentDrawing})
	}
	/* else if received.Type == "keep-alive" {
		This is a known dummy event in order to avoid accidental websocket
		connection closure. However, no action is required on the server.
	}*/

	return nil
}

func handleMessage(message string, sender *Player, lobby *Lobby) {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		return
	}

	if lobby.CurrentWord == "" {
		sendMessageToAll(trimmedMessage, sender, lobby)
		return
	}

	if sender.State == Drawing || sender.State == Standby {
		lobby.sendMessageToAllNonGuessing(trimmedMessage, sender)
	} else if sender.State == Guessing {
		lowerCasedInput := lobby.lowercaser.String(trimmedMessage)
		currentWord := lobby.CurrentWord

		normInput := simplifyText(lowerCasedInput)
		normSearched := simplifyText(currentWord)

		if normSearched == normInput {
			secondsLeft := int(lobby.RoundEndTime/1000 - time.Now().UTC().UnixNano()/1000000000)

			sender.LastScore = calculateGuesserScore(lobby.hintCount, lobby.hintsLeft, secondsLeft, lobby.DrawingTime)
			sender.Score += sender.LastScore

			lobby.scoreEarnedByGuessers += sender.LastScore
			sender.State = Standby

			lobby.TriggerUpdateEvent("correct-guess", sender.ID)

			if !lobby.isAnyoneStillGuessing() {
				advanceLobby(lobby)
			} else {
				//Since the word has been guessed correctly, we reveal it.
				lobby.WriteJSON(sender, GameEvent{Type: "update-wordhint", Data: lobby.wordHintsShown})
				recalculateRanks(lobby)
				lobby.triggerPlayersUpdate()
			}
		} else if levenshtein.ComputeDistance(normInput, normSearched) == 1 {
			//In cases of a close guess, we still send the message to everyone.
			//This allows other players to guess the word by watching what the
			//other players are misstyping.
			sendMessageToAll(trimmedMessage, sender, lobby)
			lobby.WriteJSON(sender, GameEvent{Type: "close-guess", Data: trimmedMessage})
		} else {
			sendMessageToAll(trimmedMessage, sender, lobby)
		}
	}
}

func calculateGuesserScore(hintCount, hintsLeft, secondsLeft, drawingTime int) int {
	//The base score is based on the general time taken.
	//The formula here represents an exponential decline based on the time taken.
	//This way fast players get more points, however not a lot more.
	//The bonus gained by guessing before hints are shown is therefore still somewhat relevant.
	declineFactor := 1.0 / float64(drawingTime)
	baseScore := int(maxBaseScore * math.Pow(1.0-declineFactor, float64(drawingTime-secondsLeft)))

	//Every hint not shown, e.g. not needed, will give the player bonus points.
	if hintCount < 1 {
		return baseScore
	}

	return baseScore + hintsLeft*(maxHintBonusScore/hintCount)
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
	messageEvent := GameEvent{Type: "message", Data: Message{
		Author:   sender.Name,
		AuthorID: sender.ID,
		Content:  discordemojimap.Replace(message),
	}}
	for _, target := range lobby.players {
		lobby.WriteJSON(target, messageEvent)
	}
}

func (lobby *Lobby) sendMessageToAllNonGuessing(message string, sender *Player) {
	messageEvent := GameEvent{Type: "non-guessing-player-message", Data: Message{
		Author:   sender.Name,
		AuthorID: sender.ID,
		Content:  discordemojimap.Replace(message),
	}}
	for _, target := range lobby.players {
		if target.State != Guessing {
			lobby.WriteJSON(target, messageEvent)
		}
	}
}

func handleKickVoteEvent(lobby *Lobby, player *Player, toKickID string) {
	//Kicking yourself isn't allowed
	if toKickID == player.ID {
		return
	}

	//A player can't vote twice to kick someone
	if player.votedForKick[toKickID] {
		return
	}

	playerToKickIndex := -1
	for index, otherPlayer := range lobby.players {
		if otherPlayer.ID == toKickID {
			playerToKickIndex = index
			break
		}
	}

	//If we haven't found the player, we can't kick them.
	if playerToKickIndex == -1 {
		return
	}

	playerToKick := lobby.players[playerToKickIndex]

	player.votedForKick[toKickID] = true
	var voteKickCount int
	for _, otherPlayer := range lobby.players {
		if otherPlayer.Connected && otherPlayer.votedForKick[toKickID] {
			voteKickCount++
		}
	}

	votesRequired := calculateVotesNeededToKick(playerToKick, lobby)

	kickEvent := &GameEvent{
		Type: "kick-vote",
		Data: &KickVote{
			PlayerID:          playerToKick.ID,
			PlayerName:        playerToKick.Name,
			VoteCount:         voteKickCount,
			RequiredVoteCount: votesRequired,
		},
	}

	//We send the kick event to all players, since it was a valid vote.
	for _, otherPlayer := range lobby.players {
		lobby.WriteJSON(otherPlayer, kickEvent)
	}

	//If the valid vote also happens to be the last vote needed, we kick the player.
	//Since we send the events to all players beforehand, the target player is automatically
	//being noteified of his own kick.
	if voteKickCount >= votesRequired {
		kickPlayer(lobby, playerToKick, playerToKickIndex)
	}
}

// kickPlayer kicks the given player from the lobby, updating the lobby
// state and sending all necessary events.
func kickPlayer(lobby *Lobby, playerToKick *Player, playerToKickIndex int) {
	//Avoiding nilpointer in case playerToKick disconnects during this event unluckily.
	playerToKickSocket := playerToKick.ws
	if playerToKickSocket != nil {
		disconnectError := playerToKickSocket.Close()
		if disconnectError != nil {
			log.Printf("Error disconnecting kicked player:\n\t%s\n", disconnectError)
		}
	}

	//Since the player is already kicked, we first clean up the kicking information related to that player
	for _, otherPlayer := range lobby.players {
		delete(otherPlayer.votedForKick, playerToKick.ID)
	}

	lobby.players = append(lobby.players[:playerToKickIndex], lobby.players[playerToKickIndex+1:]...)

	if lobby.drawer == playerToKick {
		lobby.TriggerUpdateEvent("drawer-kicked", nil)
		//Since the drawing person has been kicked, that probably means that he/she was trolling, therefore
		//we redact everyones last earned score.
		for _, otherPlayer := range lobby.players {
			otherPlayer.Score -= otherPlayer.LastScore
			otherPlayer.LastScore = 0
		}
		lobby.scoreEarnedByGuessers = 0
		//We must absolutely not set lobby.drawer to nil, since this would cause the drawing order to be ruined.
	}

	//If the owner is kicked, we choose the next best person as the owner.
	if lobby.Owner == playerToKick {
		for _, otherPlayer := range lobby.players {
			potentialOwner := otherPlayer
			if potentialOwner.Connected {
				lobby.Owner = potentialOwner
				lobby.TriggerUpdateEvent("owner-change", &OwnerChangeEvent{
					PlayerID:   potentialOwner.ID,
					PlayerName: potentialOwner.Name,
				})
				break
			}
		}
	}

	if lobby.drawer == playerToKick || !lobby.isAnyoneStillGuessing() {
		advanceLobby(lobby)
	} else {
		//This isn't necessary in case we need to advanced the lobby, as it has
		//to happen anyways and sending events twice would be wasteful.
		recalculateRanks(lobby)
		lobby.triggerPlayersUpdate()
	}
}

type OwnerChangeEvent struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

type NameChangeEvent struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

func calculateVotesNeededToKick(playerToKick *Player, lobby *Lobby) int {
	connectedPlayerCount := lobby.GetConnectedPlayerCount()

	//If there are only two players, e.g. none of them should be able to
	//kick the other.
	if connectedPlayerCount <= 2 {
		return 2
	}

	if playerToKick == lobby.creator {
		//We don't want to allow people to kick the creator, as this could
		//potentially annoy certain creators. For example a streamer playing
		//a game with viewers could get trolled this way. Just one
		//hypothetical scenario, I am sure there are more ;)

		//All players excluding the owner themselves.
		return connectedPlayerCount - 1
	}

	//If the amount of players equals an even number, such as 6, we will always
	//need half of that. If the amount is uneven, we'll get a floored result.
	//therefore we always add one to the amount.
	//examples:
	//    (6+1)/2 = 3
	//    (5+1)/2 = 3
	//Therefore it'll never be possible for a minority to kick a player.
	return (connectedPlayerCount + 1) / 2
}

func handleNameChangeEvent(caller *Player, lobby *Lobby, name string) {
	oldName := caller.Name
	newName := SanitizeName(name)

	log.Printf("%s is now %s\n", oldName, newName)

	//We'll avoid sending the event in this case, as it's useless, but still log
	//the event, as it might be useful to know that this happened.
	if oldName != newName {
		caller.Name = newName
		lobby.TriggerUpdateEvent("name-change", &NameChangeEvent{
			PlayerID:   caller.ID,
			PlayerName: newName,
		})
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

		//Average score, but minus one player, since the own score is 0 and doesn't count.
		playerCount := lobby.GetConnectedPlayerCount()
		//If the drawer isn't connected though, we mustn't subtract from the count.
		if drawer.Connected {
			playerCount--
		}

		var averageScore int
		if playerCount > 0 {
			averageScore = lobby.scoreEarnedByGuessers / playerCount
		}

		drawer.LastScore = averageScore
		drawer.Score += drawer.LastScore
	}

	//We need this for the next-turn event, in order to allow the client
	//to know which word was previously supposed to be guessed.
	previousWord := lobby.CurrentWord

	if lobby.DrawingTimeNew != 0 {
		lobby.DrawingTime = lobby.DrawingTimeNew
	}
	lobby.scoreEarnedByGuessers = 0
	lobby.CurrentWord = ""
	lobby.wordHints = nil

	//If the round ends and people still have guessing, that means the "Last" value
	//for the next turn has to be "no score earned".
	for _, otherPlayer := range lobby.players {
		if otherPlayer.State == Guessing {
			otherPlayer.LastScore = 0
		}
		otherPlayer.State = Guessing
		otherPlayer.votedForKick = make(map[string]bool)
	}

	newDrawer, roundOver := selectNextDrawer(lobby)
	if roundOver {
		if lobby.Round == lobby.Rounds {
			endGame(lobby)
			return
		}

		lobby.Round++
	}

	firstTurn := lobby.State != Ongoing

	lobby.ClearDrawing()
	lobby.drawer = newDrawer
	lobby.drawer.State = Drawing
	lobby.State = Ongoing
	lobby.wordChoice = GetRandomWords(3, lobby)

	recalculateRanks(lobby)

	//We use milliseconds for higher accuracy
	lobby.RoundEndTime = time.Now().UTC().UnixNano()/1000000 + int64(lobby.DrawingTime)*1000
	lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
	go startTurnTimeTicker(lobby)

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
	lobby.TriggerUpdateEvent("next-turn", nextTurnEvent)

	lobby.WriteJSON(lobby.drawer, &GameEvent{Type: "your-turn", Data: lobby.wordChoice})
}

func endGame(lobby *Lobby) {
	lobby.drawer = nil
	lobby.Round = 0
	lobby.State = GameOver

	recalculateRanks(lobby)

	for _, player := range lobby.players {
		lobby.WriteJSON(player, GameEvent{
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

// startTurnTimeTicker executes a loop that listens to the lobbies
// timeLeftTicker and executes a tickLogic on each tick. This method
// blocks until the turn ends.
func startTurnTimeTicker(lobby *Lobby) {
	for {
		ticker := lobby.timeLeftTicker
		if ticker == nil {
			return
		}
		<-ticker.C

		if !lobby.tickLogic() {
			break
		}
	}
}

// tickLogic checks whether the lobby needs to proceed to the next round and
// updates the available word hints if required. The return value indicates
// whether additional ticks are necessary or not.
func (lobby *Lobby) tickLogic() bool {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	currentTime := getTimeAsMillis()
	if currentTime >= lobby.RoundEndTime {
		advanceLobby(lobby)
		//Kill outer goroutine and therefore avoid executing hint logic.
		return false
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
					lobby.triggerWordHintUpdate()
					break
				}
			}
		}
	}

	return true
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
	sortedPlayers := make([]*Player, len(lobby.players))
	copy(sortedPlayers, lobby.players)
	sort.Slice(sortedPlayers, func(a, b int) bool {
		return sortedPlayers[a].Score > sortedPlayers[b].Score
	})

	//We start at maxint32, since we want the first player to cause an
	//increment of the the score, which will always happen this way, as
	//no player can have a score this high.
	lastScore := math.MaxInt32
	var lastRank int
	for _, player := range sortedPlayers {
		if !player.Connected {
			continue
		}

		if player.Score < lastScore {
			lastRank++
			player.Rank = lastRank
		} else {
			//Since the players are already sorted from high to low, we only
			//have the cases equal or higher.
			player.Rank = lastRank
		}

		lastScore = player.Score
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

func (lobby *Lobby) sendDataToEveryoneExceptSender(sender *Player, data interface{}) {
	for _, otherPlayer := range lobby.GetPlayers() {
		if otherPlayer != sender {
			lobby.WriteJSON(otherPlayer, data)
		}
	}
}

func (lobby *Lobby) TriggerUpdateEvent(eventType string, data interface{}) {
	event := &GameEvent{Type: eventType, Data: data}
	for _, otherPlayer := range lobby.GetPlayers() {
		lobby.WriteJSON(otherPlayer, event)
	}
}

func (lobby *Lobby) triggerUpdatePerPlayerEvent(eventType string, data func(*Player) interface{}) {
	for _, otherPlayer := range lobby.GetPlayers() {
		lobby.WriteJSON(otherPlayer, &GameEvent{Type: eventType, Data: data(otherPlayer)})
	}
}

func (lobby *Lobby) triggerPlayersUpdate() {
	lobby.TriggerUpdateEvent("update-players", lobby.players)
}

func (lobby *Lobby) triggerWordHintUpdate() {
	if lobby.CurrentWord == "" {
		return
	}

	lobby.triggerUpdatePerPlayerEvent("update-wordhint", func(player *Player) interface{} {
		return lobby.GetAvailableWordHints(player)
	})
}

// CreateLobby creates a new lobby including the initial player (owner) and
// optionally returns an error, if any occurred during creation.
func CreateLobby(playerName, chosenLanguage string, publicLobby bool, drawingTime, rounds, maxPlayers, customWordsChance, clientsPerIPLimit int, customWords []string, enableVotekick bool) (*Player, *Lobby, error) {
	lobby := &Lobby{
		LobbyID: uuid.Must(uuid.NewV4()).String(),
		EditableLobbySettings: &EditableLobbySettings{
			Rounds:            rounds,
			DrawingTime:       drawingTime,
			MaxPlayers:        maxPlayers,
			CustomWordsChance: customWordsChance,
			ClientsPerIPLimit: clientsPerIPLimit,
			EnableVotekick:    enableVotekick,
			Public:            publicLobby,
		},
		CustomWords:    customWords,
		currentDrawing: make([]interface{}, 0),
		State:          Unstarted,
		mutex:          &sync.Mutex{},
	}

	if len(customWords) > 1 {
		rand.Shuffle(len(lobby.CustomWords), func(i, j int) {
			lobby.CustomWords[i], lobby.CustomWords[j] = lobby.CustomWords[j], lobby.CustomWords[i]
		})
	}

	lobby.Wordpack = chosenLanguage

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
	lobby.Owner = player
	lobby.creator = player

	words, err := readWordList(lobby.lowercaser, chosenLanguage)
	if err != nil {
		return nil, nil, err
	}

	lobby.words = words

	return player, lobby, nil
}

// generatePlayerName creates a new playername. A so called petname. It consists
// of an adverb, an adjective and a animal name. The result can generally be
// trusted to be sane.
func generatePlayerName() string {
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
	Rounds          int           `json:"rounds"`
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
		GameState:       lobby.State,
		OwnerID:         lobby.Owner.ID,
		Round:           lobby.Round,
		Rounds:          lobby.Rounds,
		WordHints:       lobby.GetAvailableWordHints(player),
		Players:         lobby.players,
		CurrentDrawing:  lobby.currentDrawing,
	}

	if lobby.State != Ongoing {
		//Clients should interpret 0 as "time over", unless the gamestate isn't "ongoing"
		ready.RoundEndTime = 0
	} else {
		ready.RoundEndTime = int(lobby.RoundEndTime - getTimeAsMillis())
	}

	return ready
}

func (lobby *Lobby) OnPlayerConnectUnsynchronized(player *Player) {
	player.Connected = true
	recalculateRanks(lobby)
	lobby.WriteJSON(player, GameEvent{Type: "ready", Data: generateReadyData(lobby, player)})

	//This state is reached if the player reconnects before having chosen a word.
	//This can happen if the player refreshes his browser page or the socket
	//loses connection and reconnects quickly.
	if lobby.drawer == player && lobby.CurrentWord == "" {
		lobby.WriteJSON(lobby.drawer, &GameEvent{Type: "your-turn", Data: lobby.wordChoice})
	}

	//TODO Only send to everyone except for the new player, since it's part of the ready event.
	lobby.triggerPlayersUpdate()
}

func (lobby *Lobby) OnPlayerDisconnect(player *Player) {
	//We want to avoid calling the handler twice.
	if player.ws == nil {
		return
	}

	disconnectTime := time.Now()

	//It is important to properly disconnect the player before aqcuiring the mutex
	//in order to avoid false assumptions about the players connection state
	//and avoid attempting to send events.
	log.Printf("Player %s(%s) disconnected.\n", player.Name, player.ID)
	player.Connected = false
	player.ws = nil

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	player.disconnectTime = &disconnectTime
	lobby.LastPlayerDisconnectTime = &disconnectTime

	recalculateRanks(lobby)
	if lobby.hasConnectedPlayersInternal() {
		lobby.triggerPlayersUpdate()
	}
}

// GetAvailableWordHints returns a WordHint array depending on the players
// game state, since people that are drawing or have already guessed correctly
// can see all hints.
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

	lobby.players = append(lobby.players, player)

	return player
}

func (lobby *Lobby) canDraw(player *Player) bool {
	return lobby.drawer == player && lobby.CurrentWord != ""
}

var connectionCharacterReplacer = strings.NewReplacer(" ", "", "-", "", "_", "")

// simplifyText prepares the string for a more lax comparison of two words.
// Spaces, dashes, underscores and accented characters are removed or replaced.
func simplifyText(s string) string {
	return connectionCharacterReplacer.
		Replace(sanitize.Accents(s))
}
