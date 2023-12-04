package game

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/lxzan/gws"
	"github.com/mailru/easyjson"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/sanitize"

	discordemojimap "github.com/Bios-Marcel/discordemojimap/v2"
	petname "github.com/Bios-Marcel/go-petname"
	"github.com/agnivade/levenshtein"
	"github.com/gofrs/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	LobbySettingBounds = SettingBounds{
		MinDrawingTime:        60,
		MaxDrawingTime:        300,
		MinRounds:             1,
		MaxRounds:             20,
		MinMaxPlayers:         2,
		MaxMaxPlayers:         24,
		MinClientsPerIPLimit:  1,
		MaxClientsPerIPLimit:  24,
		MinCustomWordsPerTurn: 1,
		MaxCustomWordsPerTurn: 3,
		MinWordSelectCount:    1,
		MaxWordSelectCount:    5,
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
	MinDrawingTime        int `json:"minDrawingTime"`
	MaxDrawingTime        int `json:"maxDrawingTime"`
	MinRounds             int `json:"minRounds"`
	MaxRounds             int `json:"maxRounds"`
	MinMaxPlayers         int `json:"minMaxPlayers"`
	MaxMaxPlayers         int `json:"maxMaxPlayers"`
	MinClientsPerIPLimit  int `json:"minClientsPerIpLimit"`
	MaxClientsPerIPLimit  int `json:"maxClientsPerIpLimit"`
	MinCustomWordsPerTurn int `json:"minCustomWordsPerTurn"`
	MaxCustomWordsPerTurn int `json:"maxCustomWordsPerTurn"`
	MinWordSelectCount    int `json:"minWordSelectCount"`
	MaxWordSelectCount    int `json:"maxWordSelectCount"`
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

	if eventType == EventTypeMessage {
		var message StringDataEvent
		if err := easyjson.Unmarshal(payload, &message); err != nil {
			return fmt.Errorf("invalid data received: '%s'", string(payload))
		}

		handleMessage(message.Data, player, lobby)
	} else if eventType == EventTypeLine {
		if lobby.canDraw(player) {
			var line LineEvent
			if err := easyjson.Unmarshal(payload, &line); err != nil {
				return fmt.Errorf("error decoding data: %w", err)
			}

			// In case the line is too big, we overwrite the data of the event.
			// This will prevent clients from lagging due to too thick lines.
			if line.Data.LineWidth > float32(MaxBrushSize) {
				line.Data.LineWidth = MaxBrushSize
			} else if line.Data.LineWidth < float32(MinBrushSize) {
				line.Data.LineWidth = MinBrushSize
			}

			now := time.Now()
			if now.Sub(lobby.lastDrawEvent) > 150*time.Millisecond || lobby.wasLastDrawEventFill() {
				lobby.connectedDrawEventsIndexStack = append(lobby.connectedDrawEventsIndexStack, len(lobby.currentDrawing))
			}
			lobby.lastDrawEvent = now

			lobby.AppendLine(&line)

			// We directly forward the event, as it seems to be valid.
			lobby.broadcastConditional(&line, ExcludePlayer(player))
		}
	} else if eventType == EventTypeFill {
		if lobby.canDraw(player) {
			var fill FillEvent
			if err := easyjson.Unmarshal(payload, &fill); err != nil {
				return fmt.Errorf("error decoding data: %w", err)
			}

			lobby.connectedDrawEventsIndexStack = append(lobby.connectedDrawEventsIndexStack, len(lobby.currentDrawing))
			lobby.lastDrawEvent = time.Now()

			lobby.AppendFill(&fill)

			// We directly forward the event, as it seems to be valid.
			lobby.broadcastConditional(&fill, ExcludePlayer(player))
		}
	} else if eventType == EventTypeClearDrawingBoard {
		if lobby.canDraw(player) && len(lobby.currentDrawing) > 0 {
			lobby.ClearDrawing()
			lobby.broadcastConditional(
				EventTypeOnly{Type: EventTypeClearDrawingBoard},
				ExcludePlayer(player))
		}
	} else if eventType == EventTypeUndo {
		if lobby.canDraw(player) && len(lobby.currentDrawing) > 0 && len(lobby.connectedDrawEventsIndexStack) > 0 {
			undoFrom := lobby.connectedDrawEventsIndexStack[len(lobby.connectedDrawEventsIndexStack)-1]
			lobby.connectedDrawEventsIndexStack = lobby.connectedDrawEventsIndexStack[:len(lobby.connectedDrawEventsIndexStack)-1]
			if undoFrom < len(lobby.currentDrawing) {
				lobby.currentDrawing = lobby.currentDrawing[:undoFrom]
				lobby.Broadcast(&Event{Type: EventTypeDrawing, Data: lobby.currentDrawing})
			}
		}
	} else if eventType == EventTypeChooseWord {
		var wordChoice IntDataEvent

        if lobby.TimerStart {
            startRoundTimer(lobby)

            lobby.WriteObject(player, Event{
                Type: EventTypeWordChosen,
                Data: &WordChosenEvent{
                    RoundEndTime: int(lobby.roundEndTime - getTimeAsMillis()),
                },
            })
        }

		if err := easyjson.Unmarshal(payload, &wordChoice); err != nil {
			return fmt.Errorf("error decoding data: %w", err)
		}
		chosenIndex := wordChoice.Data

		if len(lobby.wordChoice) == 0 {
			return errors.New("word was chosen, even though no choice was available")
		}

		if chosenIndex < 0 || chosenIndex >= len(lobby.wordChoice) {
			return fmt.Errorf("word choice was %d, but should've been >= 0 and < %d", chosenIndex, len(lobby.wordChoice))
		}

		if player.State == Drawing {
			lobby.selectWord(chosenIndex)

			wordHintData := &Event{
			    Type: EventTypeUpdateWordHint,
			    Data: &WordHintData{
			        WordHints: lobby.wordHints,
			        RoundEndTime: int(lobby.roundEndTime - getTimeAsMillis()),
			    },
			};
			lobby.broadcastConditional(wordHintData, IsGuessing)
			wordHintDataRevealed := &Event{
                Type: EventTypeUpdateWordHint,
                Data: &WordHintData{
                    WordHints: lobby.wordHintsShown,
                    RoundEndTime: int(lobby.roundEndTime - getTimeAsMillis()),
                },
            };
			lobby.broadcastConditional(wordHintDataRevealed, IsNotGuessing)
		}
	} else if eventType == EventTypeKickVote {
		var kickEvent StringDataEvent
		if err := easyjson.Unmarshal(payload, &kickEvent); err != nil {
			return fmt.Errorf("invalid data received: '%s'", string(payload))
		}

		toKickID, err := uuid.FromString(kickEvent.Data)
		if err != nil {
			return fmt.Errorf("invalid data in kick-vote event: %v", payload)
		}

		handleKickVoteEvent(lobby, player, toKickID)
	} else if eventType == EventTypeStart {
		if lobby.State != Ongoing && player == lobby.Owner {
			// We are reseting each players score, since players could
			// technically be player a second game after the last one
			// has already ended.
			for _, otherPlayer := range lobby.players {
				otherPlayer.Score = 0
				otherPlayer.LastScore = 0
				// Everyone has the same score and therefore the same rank.
				otherPlayer.Rank = 1
			}

			// Cause advanceLobby to start at round 1, starting the game anew.
			lobby.Round = 0

			advanceLobby(lobby)
		}
	} else if eventType == EventTypeNameChange {
		var message StringDataEvent
		if err := easyjson.Unmarshal(payload, &message); err != nil {
			return fmt.Errorf("invalid data received: '%s'", string(payload))
		}

		handleNameChangeEvent(player, lobby, message.Data)
	} else if eventType == EventTypeRequestDrawing {
		// Since the client shouldn't be blocking to wait for the drawing, it's
		// fine to emit the event if there's no drawing.
		if len(lobby.currentDrawing) != 0 {
			lobby.WriteObject(player, Event{Type: EventTypeDrawing, Data: lobby.currentDrawing})
		}
	}

	return nil
}

func startRoundTimer(lobby *Lobby) {
    // We use milliseconds for higher accuracy
    lobby.roundEndTime = getTimeAsMillis() + int64(lobby.DrawingTime)*1000
    lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
    go startTurnTimeTicker(lobby, lobby.timeLeftTicker)
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
			IsNotGuessing,
		)
		return
	}

	normInput := sanitize.CleanText(lobby.lowercaser.String(trimmedMessage))
	normSearched := sanitize.CleanText(lobby.CurrentWord)

	// Since correct guess are probably the least common case, we'll always
	// calculate the distance, as usually have to do it anyway.
	switch levenshtein.ComputeDistance(normInput, normSearched) {
	case 0:
		{
			secondsLeft := int(lobby.roundEndTime/1000 - time.Now().UTC().Unix())

			sender.LastScore = calculateGuesserScore(lobby.hintCount, lobby.hintsLeft, secondsLeft, lobby.DrawingTime)
			sender.Score += sender.LastScore

			lobby.scoreEarnedByGuessers += sender.LastScore
			sender.State = Standby

			lobby.Broadcast(&Event{Type: EventTypeCorrectGuess, Data: sender.ID})

			if !lobby.isAnyoneStillGuessing() {
				advanceLobby(lobby)
			} else {
				// Since the word has been guessed correctly, we reveal it.
				lobby.WriteObject(sender, Event{
                    Type: EventTypeUpdateWordHint,
                    Data: &WordHintData{
                        WordHints: lobby.wordHints,
                        RoundEndTime: int(lobby.roundEndTime - getTimeAsMillis()),
                    },
                });

				recalculateRanks(lobby)
				lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.players})
			}
		}
	case 1:
		{
			// In cases of a close guess, we still send the message to everyone.
			// This allows other players to guess the word by watching what the
			// other players are misstyping.
			lobby.broadcastMessage(trimmedMessage, sender)
			lobby.WriteObject(sender, Event{Type: EventTypeCloseGuess, Data: trimmedMessage})
		}
	default:
		lobby.broadcastMessage(trimmedMessage, sender)
	}
}

func (lobby *Lobby) wasLastDrawEventFill() bool {
	if len(lobby.currentDrawing) == 0 {
		return false
	}
	_, isFillEvent := lobby.currentDrawing[len(lobby.currentDrawing)-1].(*FillEvent)
	return isFillEvent
}

func calculateGuesserScore(hintCount, hintsLeft, secondsLeft, drawingTime int) int {
	// The base score is based on the general time taken.
	// The formula here represents an exponential decline based on the time taken.
	// This way fast players get more points, however not a lot more.
	// The bonus gained by guessing before hints are shown is therefore still somewhat relevant.
	declineFactor := 1.0 / float64(drawingTime)
	baseScore := int(maxBaseScore * math.Pow(1.0-declineFactor, float64(drawingTime-secondsLeft)))

	// Prevent zero division panic. This could happen with two letter words.
	if hintCount <= 0 {
		return baseScore
	}

	// If all hints are shown, or the word is too short to show hints, the
	// calculation will basically always be baseScore + 0.
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

func ExcludePlayer(toExclude *Player) func(*Player) bool {
	return func(player *Player) bool {
		return player != toExclude
	}
}

func IsNotGuessing(player *Player) bool {
	return player.State != Guessing
}

func IsGuessing(player *Player) bool {
	return player.State == Guessing
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

func (lobby *Lobby) Broadcast(data easyjson.Marshaler) {
	bytes, err := easyjson.Marshal(data)
	if err != nil {
		log.Println("error marshalling Broadcast message", err)
		return
	}

	message := gws.NewBroadcaster(gws.OpcodeText, bytes)
	for _, player := range lobby.GetPlayers() {
		lobby.WritePreparedMessage(player, message)
	}
}

func (lobby *Lobby) broadcastConditional(data easyjson.Marshaler, condition func(*Player) bool) {
	var message *gws.Broadcaster
	for _, player := range lobby.players {
		if condition(player) {
			if message == nil {
				bytes, err := easyjson.Marshal(data)
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
	for index, otherPlayer := range lobby.players {
		if otherPlayer.ID == toKickID {
			playerToKickIndex = index
			break
		}
	}

	// If we haven't found the player, we can't kick them.
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

// kickPlayer kicks the given player from the lobby, updating the lobby
// state and sending all necessary events.
func kickPlayer(lobby *Lobby, playerToKick *Player, playerToKickIndex int) {
	// Avoiding nilpointer in case playerToKick disconnects during this event unluckily.
	if playerToKickSocket := playerToKick.ws; playerToKickSocket != nil {
		playerToKickSocket.WriteClose(1000, nil)
	}

	// Since the player is already kicked, we first clean up the kicking information related to that player
	for _, otherPlayer := range lobby.players {
		delete(otherPlayer.votedForKick, playerToKick.ID)
	}

	// If the owner is kicked, we choose the next best person as the owner.
	if lobby.Owner == playerToKick {
		for _, otherPlayer := range lobby.players {
			potentialOwner := otherPlayer
			if potentialOwner.Connected {
				lobby.Owner = potentialOwner
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
		lobby.players = append(lobby.players[:playerToKickIndex], lobby.players[playerToKickIndex+1:]...)
		lobby.Broadcast(&EventTypeOnly{Type: EventTypeDrawerKicked})

		// Since the drawer has been kicked, that probably means that they were
		// probably trolling, therefore we redact everyones last earned score.
		for _, otherPlayer := range lobby.players {
			otherPlayer.Score -= otherPlayer.LastScore
			otherPlayer.LastScore = 0
		}
		lobby.scoreEarnedByGuessers = 0

		advanceLobbyPredefineDrawer(lobby, roundOver, newDrawer)
	} else {
		lobby.players = append(lobby.players[:playerToKickIndex], lobby.players[playerToKickIndex+1:]...)

		if lobby.isAnyoneStillGuessing() {
			// This isn't necessary in case we need to advanced the lobby, as it has
			// to happen anyways and sending events twice would be wasteful.
			recalculateRanks(lobby)
			lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.players})
		} else {
			advanceLobby(lobby)
		}
	}
}

func (lobby *Lobby) Drawer() *Player {
	for _, player := range lobby.players {
		if player.State == Drawing {
			return player
		}
	}
	return nil
}

func calculateVotesNeededToKick(playerToKick *Player, lobby *Lobby) int {
	connectedPlayerCount := lobby.GetConnectedPlayerCount()

	// If there are only two players, e.g. none of them should be able to
	// kick the other.
	if connectedPlayerCount <= 2 {
		return 2
	}

	if playerToKick == lobby.creator {
		// We don't want to allow people to kick the creator, as this could
		// potentially annoy certain creators. For example a streamer playing
		// a game with viewers could get trolled this way. Just one
		// hypothetical scenario, I am sure there are more ;)

		// All players excluding the owner themselves.
		return connectedPlayerCount - 1
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

	log.Printf("%s is now %s\n", oldName, newName)

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
		if lobby.scoreEarnedByGuessers <= 0 {
			drawer.LastScore = 0
		} else {
			// Average score, but minus one player, since the own score is 0 and doesn't count.
			playerCount := lobby.GetConnectedPlayerCount()
			// If the drawer isn't connected though, we mustn't subtract from the count.
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
	}

	// We need this for the next-turn / game-over event, in order to allow the
	// client to know which word was previously supposed to be guessed.
	previousWord := lobby.CurrentWord
	lobby.CurrentWord = ""
	lobby.wordHints = nil

	if lobby.DrawingTimeNew != 0 {
		lobby.DrawingTime = lobby.DrawingTimeNew
	}
	lobby.scoreEarnedByGuessers = 0

	for _, otherPlayer := range lobby.players {
		// If the round ends and people still have guessing, that means the
		// "LastScore" value for the next turn has to be "no score earned".
		if otherPlayer.State == Guessing {
			otherPlayer.LastScore = 0
		}
		// Initially all players are in guessing state, as the drawer gets
		// defined further at the bottom.
		otherPlayer.State = Guessing
	}

	recalculateRanks(lobby)

	if roundOver {
		// Game over
		if lobby.Round == lobby.Rounds {
			lobby.State = GameOver

			for _, player := range lobby.players {
				readyData := generateReadyData(lobby, player)
				// The drawing is always available on the client, as the
				// game-over event is only sent to already connected players.
				readyData.CurrentDrawing = nil

				lobby.WriteObject(player, Event{
					Type: EventTypeGameOver,
					Data: &GameOverEvent{
						PreviousWord: previousWord,
						Ready:        readyData,
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
	lobby.wordChoice = GetRandomWords(lobby.WordSelectCount, lobby)


    if !lobby.TimerStart {
        startRoundTimer(lobby)
    }

    lobby.Broadcast(&Event{
        Type: EventTypeNextTurn,
        Data: &NextTurn{
            Round:        lobby.Round,
            Players:      lobby.players,
            RoundEndTime: int(lobby.roundEndTime - getTimeAsMillis()),
            PreviousWord: previousWord,
        },
    })

	lobby.WriteObject(newDrawer, &Event{Type: EventTypeYourTurn, Data: lobby.wordChoice})
}

// advanceLobby will either start the game or jump over to the next turn.
func advanceLobby(lobby *Lobby) {
	newDrawer, roundOver := determineNextDrawer(lobby)
	advanceLobbyPredefineDrawer(lobby, roundOver, newDrawer)
}

// determineNextDrawer returns the next person that's supposed to be drawing, but
// doesn't tell the lobby yet. The boolean signals whether the current round
// is over.
func determineNextDrawer(lobby *Lobby) (*Player, bool) {
	for index, player := range lobby.players {
		if player.State == Drawing {
			// If we have someone that's drawing, take the next one
			for i := index + 1; i < len(lobby.players); i++ {
				player := lobby.players[i]
				if player.Connected {
					return player, false
				}
			}

			// No player below the current drawer has been found, therefore we
			// fallback to our default logic at the bottom.
			break
		}
	}

	// We prefer the first connected player.
	for _, player := range lobby.players {
		if player.Connected {
			return player, true
		}
	}

	// If no player is connected, we simply chose the first player.
	// Safe, since the lobby can't be empty, as leaving doesn't remove players
	// from the array, but only sets them to a disconnected state.
	return lobby.players[0], true
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
	if currentTime >= lobby.roundEndTime {
		expectedTicker.Stop()
		advanceLobby(lobby)
		// Kill outer goroutine and therefore avoid executing hint logic.
		return false
	}

	if lobby.hintsLeft > 0 && lobby.wordHints != nil {
		revealHintEveryXMilliseconds := int64(lobby.DrawingTime * 1000 / (lobby.hintCount + 1))
		// If you have a drawingtime of 120 seconds and three hints, you
		// want to reveal a hint every 40 seconds, so that the two hints
		// are visible for at least a third of the time. //If the word
		// was chosen at 60 seconds, we'll still reveal one hint
		// instantly, as the time is already lower than 80.
		revealHintAtXOrLower := revealHintEveryXMilliseconds * int64(lobby.hintsLeft)
		timeLeft := lobby.roundEndTime - currentTime
		if timeLeft <= revealHintAtXOrLower {
			lobby.hintsLeft--

			// We are trying til we find a yet unshown wordhint. Since we have
			// thread safety and have already checked that there's a hint
			// left, this loop can never spin forever.
			for {
				randomIndex := rand.Int() % len(lobby.wordHints)
				if lobby.wordHints[randomIndex].Character == 0 {
					lobby.wordHints[randomIndex].Character = []rune(lobby.CurrentWord)[randomIndex]
					wordHintData := &Event{
                        Type: EventTypeUpdateWordHint,
                        Data: &WordHintData{
                            WordHints: lobby.wordHints,
                            RoundEndTime: int(lobby.roundEndTime - getTimeAsMillis()),
                        },
                    };
					lobby.broadcastConditional(wordHintData, IsGuessing)
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
	sortedPlayers := make([]*Player, len(lobby.players))
	copy(sortedPlayers, lobby.players)
	sort.Slice(sortedPlayers, func(a, b int) bool {
		return sortedPlayers[a].Score > sortedPlayers[b].Score
	})

	// We start at maxint32, since we want the first player to cause an
	// increment of the score, which will always happen this way, as
	// no player can have a score this high.
	lastScore := math.MaxInt32
	var lastRank int
	for _, player := range sortedPlayers {
		if !player.Connected {
			continue
		}

		if player.Score < lastScore {
			lastRank++
			player.Rank = lastRank
			lastScore = player.Score
		} else {
			// Since the players are already sorted from high to low, we only
			// have the cases higher or equal.
			player.Rank = lastRank
		}
	}
}

func (lobby *Lobby) selectWord(wordChoiceIndex int) {
	lobby.CurrentWord = lobby.wordChoice[wordChoiceIndex]
	lobby.wordChoice = nil

	// Depending on how long the word is, a fixed amount of hints
	// would be too easy or too hard.
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

	// We generate both the "empty" word hints and the hints for the
	// drawer. Since the length is the same, we do it in one run.
	lobby.wordHints = make([]*WordHint, 0, runeCount)
	lobby.wordHintsShown = make([]*WordHint, 0, runeCount)

	for _, char := range lobby.CurrentWord {
		// These characters are part of the word, but aren't relevant for the
		// guess. In order to make the word hints more useful to the
		// guesser, those are always shown. An example would be "Pac-Man".
		// Because these characters aren't relevant for the guess, they
		// aren't being underlined.
		isAlwaysVisibleCharacter := char == ' ' || char == '_' || char == '-'

		// The hints for the drawer are always visible, therefore they
		// don't require any handling of different cases.
		lobby.wordHintsShown = append(lobby.wordHintsShown, &WordHint{
			Character: char,
			Underline: !isAlwaysVisibleCharacter,
		})

		if isAlwaysVisibleCharacter {
			lobby.wordHints = append(lobby.wordHints, &WordHint{
				Character: char,
				Underline: false,
			})
		} else {
			lobby.wordHints = append(lobby.wordHints, &WordHint{
				Underline: true,
			})
		}
	}
}

// CreateLobby creates a new lobby including the initial player (owner) and
// optionally returns an error, if any occurred during creation.
func CreateLobby(
	cfg *config.Config,
	playerName, chosenLanguage string,
	publicLobby, timerStart bool,
	drawingTime, wordSelectCount, rounds, maxPlayers, customWordsPerTurn, clientsPerIPLimit int,
	customWords []string,
) (*Player, *Lobby, error) {
	lobby := &Lobby{
		LobbyID: uuid.Must(uuid.NewV4()).String(),
		EditableLobbySettings: EditableLobbySettings{
			Rounds:             rounds,
			DrawingTime:        drawingTime,
			WordSelectCount:    wordSelectCount,
			MaxPlayers:         maxPlayers,
			CustomWordsPerTurn: customWordsPerTurn,
			ClientsPerIPLimit:  clientsPerIPLimit,
			Public:             publicLobby,
			TimerStart:         timerStart,
		},
		CustomWords:    customWords,
		currentDrawing: make([]any, 0),
		State:          Unstarted,
		mutex:          &sync.Mutex{},
	}

	if len(customWords) > 1 {
		rand.Shuffle(len(lobby.CustomWords), func(i, j int) {
			lobby.CustomWords[i], lobby.CustomWords[j] = lobby.CustomWords[j], lobby.CustomWords[i]
		})
	}

	lobby.Wordpack = chosenLanguage

	// Necessary to correctly treat words from player, however, custom words might be treated incorrectly.
	lobby.lowercaser = cases.Lower(language.Make(getLanguageIdentifier(chosenLanguage)))

	// customWords are lowercased afterwards, as they are direct user input.
	if len(customWords) > 0 {
		for customWordIndex, customWord := range customWords {
			customWords[customWordIndex] = lobby.lowercaser.String(customWord)
		}
	}

	player := createPlayer(playerName)

	lobby.players = append(lobby.players, player)
	lobby.Owner = player
	lobby.creator = player

	return player, lobby, nil
}

// generatePlayerName creates a new playername. A so called petname. It consists
// of an adverb, an adjective and a animal name. The result can generally be
// trusted to be sane.
func generatePlayerName() string {
	return petname.Generate(3, petname.Title, petname.None)
}

func generateReadyData(lobby *Lobby, player *Player) *Ready {
	ready := &Ready{
		PlayerID:     player.ID,
		AllowDrawing: player.State == Drawing,
		PlayerName:   player.Name,

		GameState:              lobby.State,
		OwnerID:                lobby.Owner.ID,
		Round:                  lobby.Round,
		Rounds:                 lobby.Rounds,
		DrawingTimeSetting:     lobby.DrawingTime,
		WordSelectCountSetting: lobby.WordSelectCount,
		WordHints:              lobby.GetAvailableWordHints(player),
		Players:                lobby.players,
		CurrentDrawing:         lobby.currentDrawing,
	}

	if lobby.State != Ongoing {
// 		Clients should interpret 0 as "time over", unless the gamestate isn't "ongoing"
		ready.RoundEndTime = 0
	} else {
		ready.RoundEndTime = int(lobby.roundEndTime - getTimeAsMillis())
	}

	return ready
}

func (lobby *Lobby) OnPlayerConnectUnsynchronized(player *Player) {
	player.Connected = true
	recalculateRanks(lobby)
	lobby.WriteObject(player, Event{Type: EventTypeReady, Data: generateReadyData(lobby, player)})

	// This state is reached if the player reconnects before having chosen a word.
	// This can happen if the player refreshes his browser page or the socket
	// loses connection and reconnects quickly.
	if player.State == Drawing && lobby.CurrentWord == "" {
		lobby.WriteObject(player, &Event{Type: EventTypeYourTurn, Data: lobby.wordChoice})
	}

	// The player that just joined already has the most up-to-date data due
	// to the ready event being sent. Therefeore it'd be wasteful to send
	// that player and update event for players.
	lobby.broadcastConditional(&Event{
		Type: EventTypeUpdatePlayers,
		Data: lobby.players,
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
	log.Printf("Player %s(%s) disconnected.\n", player.Name, player.ID)
	player.Connected = false
	player.ws = nil

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	player.disconnectTime = &disconnectTime
	lobby.LastPlayerDisconnectTime = &disconnectTime

	recalculateRanks(lobby)
	lobby.Broadcast(&Event{Type: EventTypeUpdatePlayers, Data: lobby.players})
}

// GetAvailableWordHints returns a WordHint array depending on the players
// game state, since people that are drawing or have already guessed correctly
// can see all hints.
func (lobby *Lobby) GetAvailableWordHints(player *Player) []*WordHint {
	// The draw simple gets every character as a word-hint. We basically abuse
	// the hints for displaying the word, instead of having yet another GUI
	// element that wastes space.
	if player.State == Drawing || player.State == Standby {
		return lobby.wordHintsShown
	}

	return lobby.wordHints
}

// JoinPlayer creates a new player object using the given name and adds it
// to the lobbies playerlist. The new players is returned.
func (lobby *Lobby) JoinPlayer(playerName string) *Player {
	player := createPlayer(playerName)

	lobby.players = append(lobby.players, player)

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

	lobby.Broadcast(&EventTypeOnly{Type: EventTypeShutdown})
}
