package game

import (
	json "encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/lxzan/gws"
	"github.com/scribble-rs/scribble.rs/internal/sanitize"

	discordemojimap "github.com/Bios-Marcel/discordemojimap/v2"
	petname "github.com/Bios-Marcel/go-petname"
	"github.com/gofrs/uuid/v5"
)

var SupportedScoreCalculations = []string{
	"chill",
	"competitive",
}

var SupportedLanguages = map[string]string{
	"english_gb": "English (GB)",
	"english":    "English (US)",
	"italian":    "Italian",
	"german":     "German",
	"french":     "French",
	"dutch":      "Dutch",
	"ukrainian":  "Ukrainian",
	"russian":    "Russian",
	"polish":     "Polish",
}

const (
	DrawingBoardBaseWidth  = 1600
	DrawingBoardBaseHeight = 900
	MinBrushSize           = 8
	MaxBrushSize           = 32
	wordChoiceDurationMs   = 30000
)

// SettingBounds defines the lower and upper bounds for the user-specified
// lobby creation input.
type SettingBounds struct {
	MinDrawingTime        int `json:"minDrawingTime" env:"MIN_DRAWING_TIME"`
	MaxDrawingTime        int `json:"maxDrawingTime" env:"MAX_DRAWING_TIME"`
	MinRounds             int `json:"minRounds" env:"MIN_ROUNDS"`
	MaxRounds             int `json:"maxRounds" env:"MAX_ROUNDS"`
	MinMaxPlayers         int `json:"minMaxPlayers" env:"MIN_MAX_PLAYERS"`
	MaxMaxPlayers         int `json:"maxMaxPlayers" env:"MAX_MAX_PLAYERS"`
	MinClientsPerIPLimit  int `json:"minClientsPerIpLimit" env:"MIN_CLIENTS_PER_IP_LIMIT"`
	MaxClientsPerIPLimit  int `json:"maxClientsPerIpLimit" env:"MAX_CLIENTS_PER_IP_LIMIT"`
	MinCustomWordsPerTurn int `json:"minCustomWordsPerTurn" env:"MIN_CUSTOM_WORDS_PER_TURN"`
	MaxCustomWordsPerTurn int `json:"maxCustomWordsPerTurn" env:"MAX_CUSTOM_WORDS_PER_TURN"`
}

func (lobby *Lobby) HandleEvent(eventType string, payload []byte, player *Player) error {
	if eventType == EventTypeKeepAlive {
		// This is a known dummy event in order to avoid accidental websocket
		// connection closure. However, no action is required on the server.
		// Either way, we needn't needlessly lock the lobby.
		return nil
	}

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	// For all followup unmarshalling of the already unmarshalled Event, we
	// use mapstructure instead. It's cheaper in terms of CPU usage and
	// memory usage. There are benchmarks to prove this in json_test.go.

	if eventType == EventTypeToggleSpectate {
		player.SpectateToggleRequested = !player.SpectateToggleRequested
		if player.SpectateToggleRequested && lobby.State != Ongoing {
			if player.State == Spectating {
				player.State = Standby
			} else {
				player.State = Spectating
			}

			// Since we apply the state instantly, we reset it instantly as
			// well.
			player.SpectateToggleRequested = false
		}

		lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.Players})
	} else if eventType == EventTypeMessage {
		var message StringDataEvent
		if err := json.Unmarshal(payload, &message); err != nil {
			return fmt.Errorf("invalid data received: '%s'", string(payload))
		}

		handleMessage(message.Data, player, lobby)
	} else if eventType == EventTypeLine {
		if lobby.canDraw(player) {
			var line LineEvent
			if err := json.Unmarshal(payload, &line); err != nil {
				return fmt.Errorf("error decoding data: %w", err)
			}

			// In case the line is too big, we overwrite the data of the event.
			// This will prevent clients from lagging due to too thick lines.
			if line.Data.Width > MaxBrushSize {
				line.Data.Width = MaxBrushSize
			} else if line.Data.Width < MinBrushSize {
				line.Data.Width = MinBrushSize
			}

			now := time.Now()
			if now.Sub(lobby.lastDrawEvent) > 150*time.Millisecond || lobby.wasLastDrawEventFill() {
				lobby.ConnectedDrawEventsIndexStack = append(lobby.ConnectedDrawEventsIndexStack, len(lobby.CurrentDrawing))
			}
			lobby.lastDrawEvent = now

			lobby.AppendLine(&line)

			// We directly forward the event, as it seems to be valid.
			lobby.broadcastConditional(&line, ExcludePlayer(player))
		}
	} else if eventType == EventTypeFill {
		if lobby.canDraw(player) {
			var fill FillEvent
			if err := json.Unmarshal(payload, &fill); err != nil {
				return fmt.Errorf("error decoding data: %w", err)
			}

			lobby.ConnectedDrawEventsIndexStack = append(lobby.ConnectedDrawEventsIndexStack, len(lobby.CurrentDrawing))
			lobby.lastDrawEvent = time.Now()

			lobby.AppendFill(&fill)

			// We directly forward the event, as it seems to be valid.
			lobby.broadcastConditional(&fill, ExcludePlayer(player))
		}
	} else if eventType == EventTypeClearDrawingBoard {
		if lobby.canDraw(player) && len(lobby.CurrentDrawing) > 0 {
			lobby.ClearDrawing()
			lobby.broadcastConditional(
				EventTypeOnly{Type: EventTypeClearDrawingBoard},
				ExcludePlayer(player))
		}
	} else if eventType == EventTypeUndo {
		if lobby.canDraw(player) && len(lobby.CurrentDrawing) > 0 && len(lobby.ConnectedDrawEventsIndexStack) > 0 {
			undoFrom := lobby.ConnectedDrawEventsIndexStack[len(lobby.ConnectedDrawEventsIndexStack)-1]
			lobby.ConnectedDrawEventsIndexStack = lobby.ConnectedDrawEventsIndexStack[:len(lobby.ConnectedDrawEventsIndexStack)-1]
			if undoFrom < len(lobby.CurrentDrawing) {
				lobby.CurrentDrawing = lobby.CurrentDrawing[:undoFrom]
				lobby.Broadcast(&Event{Type: EventTypeDrawing, Data: lobby.CurrentDrawing})
			}
		}
	} else if eventType == EventTypeChooseWord {
		var wordChoice IntDataEvent
		if err := json.Unmarshal(payload, &wordChoice); err != nil {
			return fmt.Errorf("error decoding data: %w", err)
		}
		if player.State == Drawing {
			if err := lobby.selectWord(wordChoice.Data); err != nil {
				return err
			}
		}
	} else if eventType == EventTypeKickVote {
		var kickEvent StringDataEvent
		if err := json.Unmarshal(payload, &kickEvent); err != nil {
			return fmt.Errorf("invalid data received: '%s'", string(payload))
		}

		toKickID, err := uuid.FromString(kickEvent.Data)
		if err != nil {
			return fmt.Errorf("invalid data in kick-vote event: %v", payload)
		}

		handleKickVoteEvent(lobby, player, toKickID)
	} else if eventType == EventTypeToggleReadiness {
		lobby.handleToggleReadinessEvent(player)
	} else if eventType == EventTypeStart {
		if lobby.State != Ongoing && player.ID == lobby.OwnerID {
			lobby.startGame()
		}
	} else if eventType == EventTypeNameChange {
		var message StringDataEvent
		if err := json.Unmarshal(payload, &message); err != nil {
			return fmt.Errorf("invalid data received: '%s'", string(payload))
		}

		handleNameChangeEvent(player, lobby, message.Data)
	} else if eventType == EventTypeRequestDrawing {
		// Since the client shouldn't be blocking to wait for the drawing, it's
		// fine to emit the event if there's no drawing.
		if len(lobby.CurrentDrawing) != 0 {
			_ = lobby.WriteObject(player, Event{Type: EventTypeDrawing, Data: lobby.CurrentDrawing})
		}
	}

	return nil
}

func (lobby *Lobby) handleToggleReadinessEvent(player *Player) {
	if lobby.State != Ongoing && player.State != Spectating {
		if player.State != Ready {
			player.State = Ready
		} else {
			player.State = Standby
		}

		if lobby.readyToStart() {
			lobby.startGame()
		} else {
			lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.Players})
		}
	}
}

func (lobby *Lobby) readyToStart() bool {
	// Otherwise the game will start and gameover instantly. This can happen
	// if a lobby is created and the owner refreshes.
	var hasConnectedPlayers bool

	for _, otherPlayer := range lobby.Players {
		if !otherPlayer.Connected {
			continue
		}

		if otherPlayer.State != Ready {
			return false
		}
		hasConnectedPlayers = true
	}

	return hasConnectedPlayers
}

func handleMessage(message string, sender *Player, lobby *Lobby) {
	// Very long message can cause lags and can therefore be easily abused.
	// While it is debatable whether a 10000 byte (not character) long
	// message makes sense, this is technically easy to manage and therefore
	// allowed for now.
	if len(message) > 10000 {
		return
	}

	trimmedMessage := strings.TrimSpace(message)
	// Empty message can neither be a correct guess nor are useful for
	// other players in the chat.
	if trimmedMessage == "" {
		return
	}

	// If no word is currently selected, all players can talk to each other
	// and we don't have to check for corrected guesses.
	if lobby.CurrentWord == "" {
		lobby.broadcastMessage(trimmedMessage, sender)
		return
	}

	if sender.State != Guessing {
		lobby.broadcastConditional(
			newMessageEvent(EventTypeNonGuessingPlayerMessage, trimmedMessage, sender),
			IsAllowedToSeeRevealedHints,
		)
		return
	}

	normInput := sanitize.CleanText(lobby.lowercaser.String(trimmedMessage))
	normSearched := sanitize.CleanText(lobby.CurrentWord)

	switch CheckGuess(normInput, normSearched) {
	case EqualGuess:
		{
			sender.LastScore = lobby.calculateGuesserScore()
			sender.Score += sender.LastScore

			sender.State = Standby

			lobby.Broadcast(&Event{Type: EventTypeCorrectGuess, Data: sender.ID})

			if !lobby.isAnyoneStillGuessing() {
				advanceLobby(lobby)
			} else {
				// Since the word has been guessed correctly, we reveal it.
				_ = lobby.WriteObject(sender, Event{Type: EventTypeUpdateWordHint, Data: lobby.WordHintsShown})
				recalculateRanks(lobby)
				lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.Players})
			}
		}
	case CloseGuess:
		{
			// In cases of a close guess, we still send the message to everyone.
			// This allows other players to guess the word by watching what the
			// other players are misstyping.
			lobby.broadcastMessage(trimmedMessage, sender)
			_ = lobby.WriteObject(sender, Event{Type: EventTypeCloseGuess, Data: trimmedMessage})
		}
	default:
		lobby.broadcastMessage(trimmedMessage, sender)
	}
}

func (lobby *Lobby) wasLastDrawEventFill() bool {
	if len(lobby.CurrentDrawing) == 0 {
		return false
	}
	_, isFillEvent := lobby.CurrentDrawing[len(lobby.CurrentDrawing)-1].(*FillEvent)
	return isFillEvent
}

func (lobby *Lobby) isAnyoneStillGuessing() bool {
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.State == Guessing && otherPlayer.Connected {
			return true
		}
	}

	return false
}

func ExcludePlayer(toExclude *Player) func(*Player) bool {
	return func(player *Player) bool {
		return player != toExclude
	}
}

func IsAllowedToSeeRevealedHints(player *Player) bool {
	return player.State == Standby || player.State == Drawing
}

func IsAllowedToSeeHints(player *Player) bool {
	return player.State == Guessing || player.State == Spectating
}

func newMessageEvent(messageType, message string, sender *Player) *Event {
	return &Event{Type: messageType, Data: OutgoingMessage{
		Author:   sender.Name,
		AuthorID: sender.ID,
		Content:  discordemojimap.Replace(message),
	}}
}

func (lobby *Lobby) broadcastMessage(message string, sender *Player) {
	lobby.Broadcast(newMessageEvent(EventTypeMessage, message, sender))
}

func (lobby *Lobby) Broadcast(data any) {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("error marshalling Broadcast message", err)
		return
	}

	message := gws.NewBroadcaster(gws.OpcodeText, bytes)
	for _, player := range lobby.GetPlayers() {
		lobby.WritePreparedMessage(player, message)
	}
}

func (lobby *Lobby) broadcastConditional(data any, condition func(*Player) bool) {
	var message *gws.Broadcaster
	for _, player := range lobby.Players {
		if condition(player) {
			if message == nil {
				bytes, err := json.Marshal(data)
				if err != nil {
					log.Println("error marshalling broadcastConditional message", err)
					return
				}

				// Message is created lazily, since the conditional events could
				// potentially not be sent at all. The cost of the nil-check is
				// much lower than the cost of creating the message.
				message = gws.NewBroadcaster(gws.OpcodeText, bytes)
			}
			lobby.WritePreparedMessage(player, message)
		}
	}
}

func (lobby *Lobby) startGame() {
	// We are reseting each players score, since players could
	// technically be player a second game after the last one
	// has already ended.
	for _, otherPlayer := range lobby.Players {
		otherPlayer.Score = 0
		otherPlayer.LastScore = 0
		// Everyone has the same score and therefore the same rank.
		otherPlayer.Rank = 1
	}

	// Cause advanceLobby to start at round 1, starting the game anew.
	lobby.Round = 0

	advanceLobby(lobby)
}

func handleKickVoteEvent(lobby *Lobby, player *Player, toKickID uuid.UUID) {
	// Kicking yourself isn't allowed
	if toKickID == player.ID {
		return
	}

	// A player can't vote twice to kick someone
	if player.votedForKick[toKickID] {
		return
	}

	playerToKickIndex := -1
	for index, otherPlayer := range lobby.Players {
		if otherPlayer.ID == toKickID {
			playerToKickIndex = index
			break
		}
	}

	// If we haven't found the player, we can't kick them.
	if playerToKickIndex == -1 {
		return
	}

	playerToKick := lobby.Players[playerToKickIndex]

	player.votedForKick[toKickID] = true
	var voteKickCount int
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.Connected && otherPlayer.votedForKick[toKickID] {
			voteKickCount++
		}
	}

	votesRequired := calculateVotesNeededToKick(lobby)

	// We send the kick event to all players, since it was a valid vote.
	lobby.Broadcast(&Event{
		Type: EventTypeKickVote,
		Data: &KickVote{
			PlayerID:          playerToKick.ID,
			PlayerName:        playerToKick.Name,
			VoteCount:         voteKickCount,
			RequiredVoteCount: votesRequired,
		},
	})

	// If the valid vote also happens to be the last vote needed, we kick the player.
	// Since we send the events to all players beforehand, the target player is automatically
	// being noteified of his own kick.
	if voteKickCount >= votesRequired {
		kickPlayer(lobby, playerToKick, playerToKickIndex)
	}
}

func (lobby *Lobby) removePlayerByIndex(index int) {
	lobby.UserSessions = append(lobby.UserSessions[:index], lobby.UserSessions[index+1:]...)
	lobby.Players = append(lobby.Players[:index], lobby.Players[index+1:]...)
}

// kickPlayer kicks the given player from the lobby, updating the lobby
// state and sending all necessary events.
func kickPlayer(lobby *Lobby, playerToKick *Player, playerToKickIndex int) {
	// Avoiding nilpointer in case playerToKick disconnects during this event unluckily.
	if playerToKickSocket := playerToKick.ws; playerToKickSocket != nil {
		playerToKickSocket.WriteClose(1000, nil)
	}

	// Since the player is already kicked, we first clean up the kicking information related to that player
	for _, otherPlayer := range lobby.Players {
		delete(otherPlayer.votedForKick, playerToKick.ID)
	}

	// If the owner is kicked, we choose the next best person as the owner.
	if lobby.OwnerID == playerToKick.ID {
		for _, otherPlayer := range lobby.Players {
			potentialOwner := otherPlayer
			if potentialOwner.Connected {
				lobby.OwnerID = potentialOwner.ID
				lobby.Broadcast(&Event{
					Type: EventTypeOwnerChange,
					Data: &OwnerChangeEvent{
						PlayerID:   potentialOwner.ID,
						PlayerName: potentialOwner.Name,
					},
				})
				break
			}
		}
	}

	if playerToKick.State == Drawing {
		newDrawer, roundOver := determineNextDrawer(lobby)
		lobby.removePlayerByIndex(playerToKickIndex)
		lobby.Broadcast(&EventTypeOnly{Type: EventTypeDrawerKicked})

		// Since the drawer has been kicked, that probably means that they were
		// probably trolling, therefore we redact everyones last earned score.
		for _, otherPlayer := range lobby.Players {
			otherPlayer.Score -= otherPlayer.LastScore
			otherPlayer.LastScore = 0
		}

		advanceLobbyPredefineDrawer(lobby, roundOver, newDrawer)
	} else {
		lobby.removePlayerByIndex(playerToKickIndex)

		if lobby.isAnyoneStillGuessing() {
			// This isn't necessary in case we need to advanced the lobby, as it has
			// to happen anyways and sending events twice would be wasteful.
			recalculateRanks(lobby)
			lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.Players})
		} else {
			advanceLobby(lobby)
		}
	}
}

func (lobby *Lobby) Drawer() *Player {
	for _, player := range lobby.Players {
		if player.State == Drawing {
			return player
		}
	}
	return nil
}

func calculateVotesNeededToKick(lobby *Lobby) int {
	connectedPlayerCount := lobby.GetConnectedPlayerCount()

	// If there are only two players, e.g. none of them should be able to
	// kick the other.
	if connectedPlayerCount <= 2 {
		return 2
	}

	// If the amount of players equals an even number, such as 6, we will always
	// need half of that. If the amount is uneven, we'll get a floored result.
	// therefore we always add one to the amount.
	// examples:
	//    (6+1)/2 = 3
	//    (5+1)/2 = 3
	// Therefore it'll never be possible for a minority to kick a player.
	return (connectedPlayerCount + 1) / 2
}

func handleNameChangeEvent(caller *Player, lobby *Lobby, name string) {
	oldName := caller.Name
	newName := SanitizeName(name)

	// We'll avoid sending the event in this case, as it's useless, but still log
	// the event, as it might be useful to know that this happened.
	if oldName != newName {
		caller.Name = newName
		lobby.Broadcast(&Event{
			Type: EventTypeNameChange,
			Data: &NameChangeEvent{
				PlayerID:   caller.ID,
				PlayerName: newName,
			},
		})
	}
}

func (lobby *Lobby) calculateGuesserScore() int {
	return lobby.ScoreCalculation.CalculateGuesserScore(lobby)
}

func (lobby *Lobby) calculateDrawerScore() int {
	return lobby.ScoreCalculation.CalculateDrawerScore(lobby)
}

// advanceLobbyPredefineDrawer is required in cases where the drawer is removed
// from the game.
func advanceLobbyPredefineDrawer(lobby *Lobby, roundOver bool, newDrawer *Player) {
	if lobby.timeLeftTicker != nil {
		// We want to create a new ticker later on. By setting the current
		// ticker to nil, we'll cause the ticker routine to stop the ticker
		// and then stop itself. Later on we create a new routine.
		// This way we won't have race conditions or wrongly executed logic.
		lobby.timeLeftTicker = nil
	}

	// The drawer can potentially be null if kicked or the game just started.
	if drawer := lobby.Drawer(); drawer != nil {
		newDrawerScore := lobby.calculateDrawerScore()
		drawer.LastScore = newDrawerScore
		drawer.Score += newDrawerScore
	}

	// We need this for the next-turn / game-over event, in order to allow the
	// client to know which word was previously supposed to be guessed.
	previousWord := lobby.CurrentWord
	lobby.CurrentWord = ""
	lobby.WordHints = nil

	if lobby.DrawingTimeNew != 0 {
		lobby.DrawingTime = lobby.DrawingTimeNew
	}

	for _, otherPlayer := range lobby.Players {
		// If the round ends and people are still guessing, that means the
		// "LastScore" value for the next turn has to be "no score earned".
		// We also reset spectating players, to prevent any score fuckups.
		if otherPlayer.State == Guessing || otherPlayer.State == Spectating {
			otherPlayer.LastScore = 0
		}

		if otherPlayer.SpectateToggleRequested {
			otherPlayer.SpectateToggleRequested = false
			if otherPlayer.State != Spectating {
				otherPlayer.State = Spectating
			} else {
				otherPlayer.State = Guessing
			}
			continue
		}

		if otherPlayer.State == Spectating {
			continue
		}

		// Initially all players are in guessing state, as the drawer gets
		// defined further at the bottom.
		otherPlayer.State = Guessing
	}

	recalculateRanks(lobby)

	if roundOver {
		// Game over, meaning all rounds have been played out. Alternatively
		// We can reach this state if all players are spectating and or are not
		// connected anymore.
		if lobby.Round == lobby.Rounds || newDrawer == nil {
			lobby.State = GameOver

			for _, player := range lobby.Players {
				readyData := generateReadyData(lobby, player)
				// The drawing is always available on the client, as the
				// game-over event is only sent to already connected players.
				readyData.CurrentDrawing = nil

				lobby.WriteObject(player, Event{
					Type: EventTypeGameOver,
					Data: &GameOverEvent{
						PreviousWord: previousWord,
						ReadyEvent:   readyData,
					},
				})
			}

			// Omit rest of events, since we don't need to advance.
			return
		}

		lobby.Round++
	}

	lobby.ClearDrawing()
	newDrawer.State = Drawing
	lobby.State = Ongoing
	lobby.wordChoice = GetRandomWords(3, lobby)
	lobby.preSelectedWord = rand.IntN(len(lobby.wordChoice))

	lobby.Broadcast(&Event{
		Type: EventTypeNextTurn,
		Data: &NextTurn{
			Round:          lobby.Round,
			Players:        lobby.Players,
			ChoiceTimeLeft: wordChoiceDurationMs,
			PreviousWord:   previousWord,
		},
	})

	lobby.wordChoiceEndTime = getTimeAsMillis() + wordChoiceDurationMs
	go lobby.startWordChoiceTimer(wordChoiceDurationMs)

	lobby.SendYourTurnEvent(newDrawer)
}

func (lobby *Lobby) startWordChoiceTimer(durationMs int64) {
	timer := time.NewTimer(time.Duration(durationMs) * time.Millisecond)
	<-timer.C

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	// We let the timer run out as long as it doesn't seem to cause any
	// issues and make sure it doesn't fire when it would break stuff.
	if err := lobby.selectWord(int(lobby.preSelectedWord)); err != nil {
		log.Println("Error automatically selecting word:", err)
	}
}

// advanceLobby will either start the game or jump over to the next turn.
func advanceLobby(lobby *Lobby) {
	newDrawer, roundOver := determineNextDrawer(lobby)
	advanceLobbyPredefineDrawer(lobby, roundOver, newDrawer)
}

func (player *Player) desiresToDraw() bool {
	// If a player is in standby, it would break the gameloop. However, if the
	// player desired to change anyway, then it is fine.
	if player.State == Spectating {
		return player.SpectateToggleRequested
	}
	return !player.SpectateToggleRequested
}

// determineNextDrawer returns the next person that's supposed to be drawing, but
// doesn't tell the lobby yet. The boolean signals whether the current round
// is over.
func determineNextDrawer(lobby *Lobby) (*Player, bool) {
	for index, player := range lobby.Players {
		if player.State == Drawing {
			// If we have someone that's drawing, take the next one
			for i := index + 1; i < len(lobby.Players); i++ {
				nextPlayer := lobby.Players[i]
				if !nextPlayer.desiresToDraw() || !nextPlayer.Connected {
					continue
				}

				return nextPlayer, false
			}

			// No player below the current drawer has been found, therefore we
			// fallback to our default logic at the bottom.
			break
		}
	}

	// We prefer the first connected player and non-spectating.
	for _, player := range lobby.Players {
		if !player.desiresToDraw() || !player.Connected {
			continue
		}
		return player, true
	}

	// If no player is available, we will simply end the game.
	return nil, true
}

// startTurnTimeTicker executes a loop that listens to the lobbies
// timeLeftTicker and executes a tickLogic on each tick. This method
// blocks until the turn ends.
func startTurnTimeTicker(lobby *Lobby, ticker *time.Ticker) {
	for {
		<-ticker.C
		if !lobby.tickLogic(ticker) {
			break
		}
	}
}

// tickLogic checks whether the lobby needs to proceed to the next round and
// updates the available word hints if required. The return value indicates
// whether additional ticks are necessary or not. The ticker is automatically
// stopped if no additional ticks are required.
func (lobby *Lobby) tickLogic(expectedTicker *time.Ticker) bool {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	// Since we have a lock on the lobby, we can find out if the ticker we are
	// listening to is still valid. If not, we want to kill the outer routine.
	if lobby.timeLeftTicker != expectedTicker {
		expectedTicker.Stop()
		return false
	}

	currentTime := getTimeAsMillis()
	if currentTime >= lobby.RoundEndTime {
		expectedTicker.Stop()
		advanceLobby(lobby)
		// Kill outer goroutine and therefore avoid executing hint logic.
		return false
	}

	if lobby.HintsLeft > 0 && lobby.WordHints != nil {
		revealHintEveryXMilliseconds := int64(lobby.DrawingTime * 1000 / (lobby.HintCount + 1))
		// If you have a drawingtime of 120 seconds and three hints, you
		// want to reveal a hint every 40 seconds, so that the two hints
		// are visible for at least a third of the time. //If the word
		// was chosen at 60 seconds, we'll still reveal one hint
		// instantly, as the time is already lower than 80.
		revealHintAtXOrLower := revealHintEveryXMilliseconds * int64(lobby.HintsLeft)
		timeLeft := lobby.RoundEndTime - currentTime
		if timeLeft <= revealHintAtXOrLower {
			lobby.HintsLeft--

			// We are trying til we find a yet unshown wordhint. Since we have
			// thread safety and have already checked that there's a hint
			// left, this loop can never spin forever.
			for {
				randomIndex := rand.Int() % len(lobby.WordHints)
				if lobby.WordHints[randomIndex].Character == 0 {
					lobby.WordHints[randomIndex].Character = []rune(lobby.CurrentWord)[randomIndex]
					wordHintData := &Event{
						Type: EventTypeUpdateWordHint,
						Data: lobby.WordHints,
					}
					lobby.broadcastConditional(wordHintData, IsAllowedToSeeHints)
					break
				}
			}
		}
	}

	return true
}

func getTimeAsMillis() int64 {
	return time.Now().UTC().UnixMilli()
}

// recalculateRanks will assign each player his respective rank in the lobby
// according to everyones current score. This will not trigger any events.
func recalculateRanks(lobby *Lobby) {
	// We don't directly sort the players, since the order determines in which
	// order the players will have to draw.
	sortedPlayers := make([]*Player, len(lobby.Players))
	copy(sortedPlayers, lobby.Players)
	sort.Slice(sortedPlayers, func(a, b int) bool {
		return sortedPlayers[a].Score > sortedPlayers[b].Score
	})

	// We start at maxint32, since we want the first player to cause an
	// increment of the score, which will always happen this way, as
	// no player can have a score this high.
	lastScore := math.MaxInt32
	var lastRank int
	for _, player := range sortedPlayers {
		if !player.Connected && player.State != Spectating {
			continue
		}

		if player.Score < lastScore {
			lastRank++
			lastScore = player.Score
		}

		player.Rank = lastRank
	}
}

func (lobby *Lobby) selectWord(index int) error {
	if lobby.State != Ongoing {
		return errors.New("word was chosen, even though the game wasn't ongoing")
	}

	if len(lobby.wordChoice) == 0 {
		return errors.New("word was chosen, even though no choice was available")
	}

	if index < 0 || index >= len(lobby.wordChoice) {
		return fmt.Errorf("word choice was %d, but should've been >= 0 and < %d",
			index, len(lobby.wordChoice))
	}

	lobby.CurrentWord = lobby.wordChoice[index]
	lobby.wordChoice = nil

	// Depending on how long the word is, a fixed amount of hints
	// would be too easy or too hard.
	runeCount := utf8.RuneCountInString(lobby.CurrentWord)
	if runeCount <= 2 {
		lobby.HintCount = 0
	} else if runeCount <= 4 {
		lobby.HintCount = 1
	} else if runeCount <= 9 {
		lobby.HintCount = 2
	} else {
		lobby.HintCount = 3
	}
	lobby.HintsLeft = lobby.HintCount

	// We generate both the "empty" word hints and the hints for the
	// drawer. Since the length is the same, we do it in one run.
	lobby.WordHints = make([]*WordHint, 0, runeCount)
	lobby.WordHintsShown = make([]*WordHint, 0, runeCount)

	for _, char := range lobby.CurrentWord {
		// These characters are part of the word, but aren't relevant for the
		// guess. In order to make the word hints more useful to the
		// guesser, those are always shown. An example would be "Pac-Man".
		// Because these characters aren't relevant for the guess, they
		// aren't being underlined.
		isAlwaysVisibleCharacter := char == ' ' || char == '_' || char == '-'

		// The hints for the drawer are always visible, therefore they
		// don't require any handling of different cases.
		lobby.WordHintsShown = append(lobby.WordHintsShown, &WordHint{
			Character: char,
			Underline: !isAlwaysVisibleCharacter,
		})

		if isAlwaysVisibleCharacter {
			lobby.WordHints = append(lobby.WordHints, &WordHint{
				Character: char,
				Underline: false,
			})
		} else {
			lobby.WordHints = append(lobby.WordHints, &WordHint{
				Underline: true,
			})
		}
	}
	// We use milliseconds for higher accuracy
	lobby.RoundEndTime = getTimeAsMillis() + int64(lobby.DrawingTime)*1000
	lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
	go startTurnTimeTicker(lobby, lobby.timeLeftTicker)

	wordHintData := &Event{
		Type: EventTypeWordChosen,
		Data: &WordChosen{
			Hints:    lobby.WordHints,
			TimeLeft: int(lobby.RoundEndTime - getTimeAsMillis()),
		},
	}
	lobby.broadcastConditional(wordHintData, IsAllowedToSeeHints)
	wordHintDataRevealed := &Event{
		Type: EventTypeWordChosen,
		Data: &WordChosen{
			Hints:    lobby.WordHintsShown,
			TimeLeft: int(lobby.RoundEndTime - getTimeAsMillis()),
		},
	}
	lobby.broadcastConditional(wordHintDataRevealed, IsAllowedToSeeRevealedHints)

	return nil
}

// CreateLobby creates a new lobby including the initial player (owner) and
// optionally returns an error, if any occurred during creation.
func CreateLobby(
	desiredLobbyId string,
	playerName, chosenLanguage string,
	publicLobby bool,
	drawingTime, rounds, maxPlayers, customWordsPerTurn, clientsPerIPLimit int,
	customWords []string,
	scoringCalculation ScoreCalculation,
) (*Player, *Lobby, error) {
	if desiredLobbyId == "" {
		desiredLobbyId = uuid.Must(uuid.NewV4()).String()
	}
	lobby := &Lobby{
		LobbyID: desiredLobbyId,
		LobbySettings: LobbySettings{
			ScoreCalculationIdentifier: scoringCalculation.Identifier(),
			EditableLobbySettings: EditableLobbySettings{
				Rounds:             rounds,
				DrawingTime:        drawingTime,
				MaxPlayers:         maxPlayers,
				CustomWordsPerTurn: customWordsPerTurn,
				ClientsPerIPLimit:  clientsPerIPLimit,
				Public:             publicLobby,
			},
		},
		CustomWords:      customWords,
		CurrentDrawing:   make([]any, 0),
		State:            Unstarted,
		ScoreCalculation: scoringCalculation,
	}

	if len(customWords) > 1 {
		rand.Shuffle(len(lobby.CustomWords), func(i, j int) {
			lobby.CustomWords[i], lobby.CustomWords[j] = lobby.CustomWords[j], lobby.CustomWords[i]
		})
	}

	lobby.Wordpack = chosenLanguage

	// Necessary to correctly treat words from player, however, custom words
	// might be treated incorrectly, as they might not be the same language as
	// the one specified for the lobby. If for example you chose 100 french
	// custom words, but keep english_us as the lobby language, the casing rules
	// will most likely be faulty.
	lobby.lowercaser = WordlistData[chosenLanguage].Lowercaser()

	// customWords are lowercased afterwards, as they are direct user input.
	if len(customWords) > 0 {
		for customWordIndex, customWord := range customWords {
			customWords[customWordIndex] = lobby.lowercaser.String(customWord)
		}
	}

	player := lobby.JoinPlayer(playerName)
	lobby.OwnerID = player.ID

	return player, lobby, nil
}

// generatePlayerName creates a new playername. A so called petname. It consists
// of an adverb, an adjective and a animal name. The result can generally be
// trusted to be sane.
func generatePlayerName() string {
	return petname.Generate(3, petname.Title, petname.None)
}

func generateReadyData(lobby *Lobby, player *Player) *ReadyEvent {
	ready := &ReadyEvent{
		PlayerID:     player.ID,
		AllowDrawing: player.State == Drawing,
		PlayerName:   player.Name,

		GameState:          lobby.State,
		OwnerID:            lobby.OwnerID,
		Round:              lobby.Round,
		Rounds:             lobby.Rounds,
		DrawingTimeSetting: lobby.DrawingTime,
		WordHints:          lobby.GetAvailableWordHints(player),
		Players:            lobby.Players,
		CurrentDrawing:     lobby.CurrentDrawing,
	}

	if lobby.State != Ongoing {
		// Clients should interpret 0 as "time over", unless the gamestate isn't "ongoing"
		ready.TimeLeft = 0
	} else {
		ready.TimeLeft = int(lobby.RoundEndTime - getTimeAsMillis())
	}

	return ready
}

func (lobby *Lobby) SendYourTurnEvent(player *Player) {
	lobby.WriteObject(player, &Event{
		Type: EventTypeYourTurn,
		Data: &YourTurn{
			TimeLeft:        int(lobby.wordChoiceEndTime - getTimeAsMillis()),
			PreSelectedWord: lobby.preSelectedWord,
			Words:           lobby.wordChoice,
		},
	})
}

func (lobby *Lobby) OnPlayerConnectUnsynchronized(player *Player) {
	player.Connected = true
	recalculateRanks(lobby)
	lobby.WriteObject(player, Event{Type: EventTypeReady, Data: generateReadyData(lobby, player)})

	// This state is reached if the player reconnects before having chosen a word.
	// This can happen if the player refreshes his browser page or the socket
	// loses connection and reconnects quickly.
	if player.State == Drawing && lobby.CurrentWord == "" {
		lobby.SendYourTurnEvent(player)
	}

	// The player that just joined already has the most up-to-date data due
	// to the ready event being sent. Therefeore it'd be wasteful to send
	// that player and update event for players.
	lobby.broadcastConditional(&Event{
		Type: EventTypeUpdatePlayers,
		Data: lobby.Players,
	}, ExcludePlayer(player))
}

func (lobby *Lobby) OnPlayerDisconnect(player *Player) {
	// We want to avoid calling the handler twice.
	if player.ws == nil {
		return
	}

	disconnectTime := time.Now()

	// It is important to properly disconnect the player before aqcuiring the mutex
	// in order to avoid false assumptions about the players connection state
	// and avoid attempting to send events.
	player.Connected = false
	player.ws = nil

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	player.disconnectTime = &disconnectTime
	lobby.LastPlayerDisconnectTime = &disconnectTime

	// Reset from potentially ready to standby
	if lobby.State != Ongoing {
		// FIXME Should we not set spectators to standby? Currently there's no
		// indication you are spectating right now.
		player.State = Standby
		if lobby.readyToStart() {
			lobby.startGame()
			// Rank Calculation and sending out player updates happened anyway,
			// so there's no need to keep going.
			return
		}
	}

	// Necessary to prevent gaps in the ranking. While players preserve their
	// points when disconnecting, they shouldn't preserve their ranking. Upon
	// reconnecting, the ranking will be recalculated though.
	recalculateRanks(lobby)
	lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.Players})
}

// GetAvailableWordHints returns a WordHint array depending on the players
// game state, since people that are drawing or have already guessed correctly
// can see all hints.
func (lobby *Lobby) GetAvailableWordHints(player *Player) []*WordHint {
	// The draw simple gets every character as a word-hint. We basically abuse
	// the hints for displaying the word, instead of having yet another GUI
	// element that wastes space.
	if player.State != Guessing {
		return lobby.WordHintsShown
	}

	return lobby.WordHints
}

// JoinPlayer creates a new player object using the given name and adds it
// to the lobbies playerlist. The new players is returned.
func (lobby *Lobby) JoinPlayer(name string) *Player {
	player := &Player{
		Name:         SanitizeName(name),
		ID:           uuid.Must(uuid.NewV4()),
		userSession:  uuid.Must(uuid.NewV4()),
		votedForKick: make(map[uuid.UUID]bool),
	}

	if lobby.State == Ongoing {
		// Joining an existing game will mark you as a guesser, as someone is
		// always drawing, given there is no pause-state.
		player.State = Guessing
	} else {
		player.State = Standby
	}
	lobby.Players = append(lobby.Players, player)
	lobby.UserSessions = append(lobby.UserSessions, player.userSession)

	return player
}

func (lobby *Lobby) canDraw(player *Player) bool {
	return player.State == Drawing && lobby.CurrentWord != "" && lobby.State == Ongoing
}

// Shutdown sends all players an event, indicating that the lobby
// will be shut down. The caller of this function should take care of not
// allowing new connections. Clients should gracefully disconnect.
func (lobby *Lobby) Shutdown() {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	log.Println("Lobby Shutdown: Mutex acquired")

	state, err := json.Marshal(LobbyRestoreData{
		ShutdownTime: time.Now(),
		Lobby:        lobby,
	})
	if err != nil {
		log.Println("Error marshalling lobby:", err)
		lobby.Broadcast(&EventTypeOnly{Type: EventTypeShutdown})
	} else {
		lobby.Broadcast(&Event{Type: EventTypeShutdown, Data: state})
	}

	// Since broadcast is synchronous, we gotta use the asynchronous queue, to
	// make sure the message is received before closing.
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(lobby.Players))
	for _, player := range lobby.Players {
		player.ws.Async(func() {
			defer waitGroup.Done()
			player.ws.WriteClose(1012, []byte("server_restart"))
		})
	}
	waitGroup.Wait()
}

// ScoreCalculation allows having different scoring systems for a lobby.
type ScoreCalculation interface {
	Identifier() string
	CalculateGuesserScore(lobby *Lobby) int
	CalculateDrawerScore(lobby *Lobby) int
}

var ChillScoring = &adjustableScoringAlgorithm{
	identifier:                  "chill",
	baseScore:                   100.0,
	maxBonusBaseScore:           100.0,
	bonusBaseScoreDeclineFactor: 2.0,
	maxHintBonusScore:           60.0,
}

var CompetitiveScoring = &adjustableScoringAlgorithm{
	identifier:                  "competitive",
	baseScore:                   10.0,
	maxBonusBaseScore:           290.0,
	bonusBaseScoreDeclineFactor: 3.0,
	maxHintBonusScore:           120.0,
}

type adjustableScoringAlgorithm struct {
	identifier                  string
	baseScore                   float64
	maxBonusBaseScore           float64
	bonusBaseScoreDeclineFactor float64
	maxHintBonusScore           float64
}

func (s *adjustableScoringAlgorithm) Identifier() string {
	return s.identifier
}

func (s *adjustableScoringAlgorithm) CalculateGuesserScore(lobby *Lobby) int {
	return s.CalculateGuesserScoreInternal(lobby.HintCount, lobby.HintsLeft, lobby.DrawingTime, lobby.RoundEndTime)
}

func (s *adjustableScoringAlgorithm) MaxScore() int {
	return int(s.baseScore + s.maxBonusBaseScore + s.maxHintBonusScore)
}

func (s *adjustableScoringAlgorithm) CalculateGuesserScoreInternal(
	hintCount, hintsLeft, drawingTime int,
	roundEndTimeMillis int64,
) int {
	secondsLeft := int(roundEndTimeMillis/1000 - time.Now().UTC().Unix())

	declineFactor := s.bonusBaseScoreDeclineFactor / float64(drawingTime)
	score := int(
		s.baseScore + s.maxBonusBaseScore*math.Pow(1.0-declineFactor, float64(drawingTime-secondsLeft)))

	// Prevent zero division panic. This could happen with two letter words.
	if hintCount > 0 {
		score += hintsLeft * (int(s.maxHintBonusScore) / hintCount)
	}

	return score
}

func (s *adjustableScoringAlgorithm) CalculateDrawerScore(lobby *Lobby) int {
	// The drawer can get points even if disconnected. But if they are
	// connected, we need to ignore them when calculating their score.
	var (
		playerCount int
		scoreSum    int
	)
	for _, player := range lobby.GetPlayers() {
		if player.State != Drawing &&
			// Switch to spectating is only possible after score calculation, so
			// this can't be used to manipulate score.
			player.State != Spectating &&
			// If the player has guessed, we want to take them into account,
			// even if they aren't connected anymore. If the player is
			// connected, but hasn't guessed, it is still as well, as the
			// drawing must've not been good enough to be guessable.
			(player.Connected || player.LastScore > 0) {
			scoreSum += player.LastScore
			playerCount++
		}
	}

	if playerCount > 0 {
		return scoreSum / playerCount
	}

	return 0
}
