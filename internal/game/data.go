package game

import (
	"strings"
	"sync"
	"time"

	discordemojimap "github.com/Bios-Marcel/discordemojimap/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/lxzan/gws"
	"golang.org/x/text/cases"
)

// slotReservationTime should give a player enough time to restart their browser
// without losing their slost.
const slotReservationTime = time.Minute * 1

// Lobby represents a game session. It must not be sent via the API, as it
// exposes gameplay relevant information.
type Lobby struct {
	// ID uniquely identified the Lobby.
	LobbyID string

	LobbySettings

	// DrawingTimeNew is the new value of the drawing time. If a round is
	// already ongoing, we can't simply change the drawing time, as it would
	// screw with the score calculation of the current turn.
	DrawingTimeNew int

	CustomWords []string
	Words       []string

	// Players references all participants of the Lobby. Indices are synced
	// with [Lobby.UserSessions].
	Players []*Player
	// Holds all UserSessions. Indices are synced with [Lobby.Players].
	UserSessions []uuid.UUID

	// Whether the game has started, is ongoing or already over.
	State State
	// OwnerID references the Player that currently owns the lobby.
	// Meaning this player has rights to restart or change certain settings.
	OwnerID uuid.UUID
	// ScoreCalculation decides how scores for both guessers and drawers are
	// determined.
	ScoreCalculation ScoreCalculation `json:"-"`
	// CurrentWord represents the word that was last selected. If no word has
	// been selected yet or the round is already over, this should be empty.
	CurrentWord string
	// WordHints for the current word.
	WordHints []*WordHint
	// WordHintsShown are the same as wordHints with characters visible.
	WordHintsShown []*WordHint
	// HintsLeft is the amount of hints still available for revelation.
	HintsLeft int
	// HintCount is the amount of hints that were initially available
	// for revelation.
	HintCount int
	// Round is the round that the Lobby is currently in. This is a number
	// between 0 and Rounds. 0 indicates that it hasn't started yet.
	Round             int
	wordChoiceEndTime int64
	preSelectedWord   int
	// wordChoice represents the current choice of words present to the drawer.
	wordChoice []string
	// RoundEndTime represents the time at which the current round will end.
	// This is a UTC unix-timestamp in milliseconds.
	RoundEndTime int64

	timeLeftTicker *time.Ticker
	// CurrentDrawing represents the state of the current canvas. The elements
	// consist of LineEvent and FillEvent. Please do not modify the contents
	// of this array an only move AppendLine and AppendFill on the respective
	// lobby object.
	CurrentDrawing []any

	// These variables are used to define the ranges of connected drawing events.
	// For example a line that has been drawn or a fill that has been executed.
	// Since we can't trust the client to tell us this, we use the time passed
	// between draw events as an indicator of which draw events make up one line.
	// An alternative approach could be using the coordinates and see if they are
	// connected, but that could technically undo a whole drawing.

	lastDrawEvent                 time.Time
	ConnectedDrawEventsIndexStack []int

	lowercaser cases.Caser

	// LastPlayerDisconnectTime is used to know since when a lobby is empty, in case
	// it is empty. If the time is nil, it's treated the same as when the
	// timelimit has been reached.
	LastPlayerDisconnectTime *time.Time

	mutex sync.Mutex

	WriteObject          func(*Player, any) error              `json:"-"`
	WritePreparedMessage func(*Player, *gws.Broadcaster) error `json:"-"`
}

type LobbyRestoreData struct {
	ShutdownTime time.Time
	Lobby        *Lobby
}

func (lobby *Lobby) ResurrectUnsynchronized(restoreData *LobbyRestoreData) {
	lobby.lowercaser = WordlistData[lobby.Wordpack].Lowercaser()

	// Since we don't know how long the restart took, we extend all timers.\
	// We add an additional second for good measure.
	now := time.Now()
	timeDiff := now.Sub(restoreData.ShutdownTime).Milliseconds() + 1000

	lobby.RoundEndTime = lobby.RoundEndTime + int64(timeDiff)

	if lobby.CurrentWord != "" {
		lobby.timeLeftTicker = time.NewTicker(1 * time.Second)
		go startTurnTimeTicker(lobby, lobby.timeLeftTicker)
	} else if len(lobby.wordChoice) > 0 {
		lobby.wordChoiceEndTime = lobby.wordChoiceEndTime + int64(timeDiff)
		go lobby.startWordChoiceTimer(lobby.wordChoiceEndTime - now.UTC().UnixMilli())
	}
}

// MaxPlayerNameLength defines how long a string can be at max when used
// as the playername.
const MaxPlayerNameLength int = 30

// GetLastKnownAddress returns the last known IP-Address used for an HTTP request.
func (player *Player) GetLastKnownAddress() string {
	return player.lastKnownAddress
}

// SetLastKnownAddress sets the last known IP-Address used for an HTTP request.
// Can be retrieved via GetLastKnownAddress().
func (player *Player) SetLastKnownAddress(address string) {
	player.lastKnownAddress = address
}

// GetWebsocket simply returns the players websocket connection. This method
// exists to encapsulate the websocket field and prevent accidental sending
// the websocket data via the network.
func (player *Player) GetWebsocket() *gws.Conn {
	return player.ws
}

// SetWebsocket sets the given connection as the players websocket connection.
func (player *Player) SetWebsocket(socket *gws.Conn) {
	player.ws = socket
}

// GetUserSession returns the players current user session.
func (player *Player) GetUserSession() uuid.UUID {
	return player.userSession
}

type PlayerState string

const (
	Guessing   PlayerState = "guessing"
	Drawing    PlayerState = "drawing"
	Standby    PlayerState = "standby"
	Ready      PlayerState = "ready"
	Spectating PlayerState = "spectating"
)

func (lobby *Lobby) GetPlayerByID(id uuid.UUID) *Player {
	for _, player := range lobby.Players {
		if player.ID == player.ID {
			return player
		}
	}

	return nil
}

func (lobby *Lobby) GetPlayerBySession(userSession uuid.UUID) *Player {
	for index, uuid := range lobby.UserSessions {
		if uuid == userSession {
			return lobby.Players[index]
		}
	}
	return nil
}

func (lobby *Lobby) GetOwner() *Player {
	return lobby.GetPlayerByID(lobby.OwnerID)
}

func (lobby *Lobby) ClearDrawing() {
	lobby.CurrentDrawing = make([]any, 0)
	lobby.ConnectedDrawEventsIndexStack = nil
}

// AppendLine adds a line direction to the current drawing. This exists in order
// to prevent adding arbitrary elements to the drawing, as the backing array is
// an empty interface type.
func (lobby *Lobby) AppendLine(line *LineEvent) {
	lobby.CurrentDrawing = append(lobby.CurrentDrawing, line)
}

// AppendFill adds a fill direction to the current drawing. This exists in order
// to prevent adding arbitrary elements to the drawing, as the backing array is
// an empty interface type.
func (lobby *Lobby) AppendFill(fill *FillEvent) {
	lobby.CurrentDrawing = append(lobby.CurrentDrawing, fill)
}

// SanitizeName removes invalid characters from the players name, resolves
// emoji codes, limits the name length and generates a new name if necessary.
func SanitizeName(name string) string {
	// We trim and handle emojis beforehand to avoid taking this into account
	// when checking the name length, so we don't cut off too much of the name.
	newName := discordemojimap.Replace(strings.TrimSpace(name))

	// We don't want super-long names
	if len(newName) > MaxPlayerNameLength {
		return newName[:MaxPlayerNameLength+1]
	}

	if newName != "" {
		return newName
	}

	return generatePlayerName()
}

// GetConnectedPlayerCount returns the amount of player that have currently
// established a socket connection.
func (lobby *Lobby) GetConnectedPlayerCount() int {
	var count int
	for _, player := range lobby.Players {
		if player.Connected {
			count++
		}
	}

	return count
}

func (lobby *Lobby) HasConnectedPlayers() bool {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	for _, otherPlayer := range lobby.Players {
		if otherPlayer.Connected {
			return true
		}
	}

	return false
}

// CanIPConnect checks whether the IP is still allowed regarding the lobbies
// clients per IP address limit. This function should only be called for
// players that aren't already in the lobby.
func (lobby *Lobby) CanIPConnect(address string) bool {
	var clientsWithSameIP int
	for _, player := range lobby.GetPlayers() {
		if player.GetLastKnownAddress() == address {
			clientsWithSameIP++
			if clientsWithSameIP >= lobby.ClientsPerIPLimit {
				return false
			}
		}
	}

	return true
}

func (lobby *Lobby) IsPublic() bool {
	return lobby.Public
}

func (lobby *Lobby) GetPlayers() []*Player {
	return lobby.Players
}

// GetOccupiedPlayerSlots counts the available slots which can be taken by new
// players. Whether a slot is available is determined by the player count and
// whether a player is disconnect or furthermore how long they have been
// disconnected for. Therefore the result of this function will differ from
// Lobby.GetConnectedPlayerCount.
func (lobby *Lobby) GetOccupiedPlayerSlots() int {
	var occupiedPlayerSlots int
	now := time.Now()
	for _, player := range lobby.Players {
		if player.Connected {
			occupiedPlayerSlots++
		} else {
			disconnectTime := player.disconnectTime

			// If a player hasn't been disconnected for a certain
			// timeframe, we will reserve the slot. This avoids frustration
			// in situations where a player has to restart their PC or so.
			if disconnectTime == nil || now.Sub(*disconnectTime) < slotReservationTime {
				occupiedPlayerSlots++
			}
		}
	}

	return occupiedPlayerSlots
}

// HasFreePlayerSlot determines whether the lobby still has a slot for at
// least one more player. If a player has disconnected recently, the slot
// will be preserved for 5 minutes. This function should be used over
// Lobby.GetOccupiedPlayerSlots, as it is potentially faster.
func (lobby *Lobby) HasFreePlayerSlot() bool {
	if len(lobby.Players) < lobby.MaxPlayers {
		return true
	}

	return lobby.GetOccupiedPlayerSlots() < lobby.MaxPlayers
}

// Synchronized allows running a function while keeping the lobby locked via
// it's own mutex. This is useful in order to avoid having to relock a lobby
// multiple times, which might cause unexpected inconsistencies.
func (lobby *Lobby) Synchronized(logic func()) {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	logic()
}
