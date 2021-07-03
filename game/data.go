package game

import (
	"strings"
	"sync"
	"time"

	discordemojimap "github.com/Bios-Marcel/discordemojimap/v2"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/text/cases"
)

const slotReservationTime = time.Minute * 5

// Lobby represents a game session.
// FIXME Field visibilities should be changed in case we ever serialize this.
type Lobby struct {
	// ID uniquely identified the Lobby.
	LobbyID string

	*EditableLobbySettings

	// DrawingTimeNew is the new value of the drawing time. If a round is
	// already ongoing, we can't simply change the drawing time, as it would
	// screw with the score calculation of the current turn.
	DrawingTimeNew int

	CustomWords []string
	words       []string

	// players references all participants of the Lobby.
	players []*Player

	// Whether the game has started, is ongoing or already over.
	State gameState
	// drawer references the Player that is currently drawing.
	drawer *Player
	// Owner references the Player that currently owns the lobby.
	// Meaning this player has rights to restart or change certain settings.
	Owner *Player
	// creator is the player that opened a lobby. Initially creator and owner
	// are set to the same player. While the owner can change throughout the
	// game, the creator can't.
	creator *Player
	// CurrentWord represents the word that was last selected. If no word has
	// been selected yet or the round is already over, this should be empty.
	CurrentWord string
	// wordHints for the current word.
	wordHints []*WordHint
	// wordHintsShown are the same as wordHints with characters visible.
	wordHintsShown []*WordHint
	// hintsLeft is the amount of hints still available for revelation.
	hintsLeft int
	// hintCount is the amount of hints that were initially available
	//for revelation.
	hintCount int
	// Round is the round that the Lobby is currently in. This is a number
	// between 0 and Rounds. 0 indicates that it hasn't started yet.
	Round int
	// wordChoice represents the current choice of words present to the drawer.
	wordChoice []string
	Wordpack   string
	// RoundEndTime represents the time at which the current round will end.
	// This is a UTC unix-timestamp in milliseconds.
	RoundEndTime int64

	timeLeftTicker        *time.Ticker
	scoreEarnedByGuessers int
	// currentDrawing represents the state of the current canvas. The elements
	// consist of LineEvent and FillEvent. Please do not modify the contents
	// of this array an only move AppendLine and AppendFill on the respective
	// lobby object.
	currentDrawing []interface{}

	// These variables are used to define the ranges of connected drawing events.
	// For example a line that has been drawn or a fill that has been executed.
	// Since we can't trust the client to tell us this, we use the time passed
	// between draw events as an indicator of which draw events make up one line.
	// An alternative approach could be using the coordinates and see if they are
	// connected, but that could technically undo a whole drawing.

	lastDrawEvent                 time.Time
	connectedDrawEventsIndexStack []int

	lowercaser cases.Caser

	//LastPlayerDisconnectTime is used to know since when a lobby is empty, in case
	//it is empty.
	LastPlayerDisconnectTime *time.Time

	mutex *sync.Mutex

	WriteJSON func(player *Player, object interface{}) error
}

// EditableLobbySettings represents all lobby settings that are editable by
// the lobby owner after the lobby has already been opened.
type EditableLobbySettings struct {
	// MaxPlayers defines the maximum amount of players in a single lobby.
	MaxPlayers int `json:"maxPlayers"`
	// CustomWords are additional words that will be used in addition to the
	// predefined words.
	// Public defines whether the lobby is being broadcast to clients asking
	// for available lobbies.
	Public bool `json:"public"`
	// EnableVotekick decides whether players are allowed to kick eachother
	// by casting majority votes.
	EnableVotekick bool `json:"enableVotekick"`
	// CustomWordsChance determines the chance of each word being a custom
	// word on the next word prompt. This needs to be an integer between
	// 0 and 100. The value represents a percentage.
	CustomWordsChance int `json:"customWordsChance"`
	// ClientsPerIPLimit helps preventing griefing by reducing each player
	// to one tab per IP address.
	ClientsPerIPLimit int `json:"clientsPerIpLimit"`
	// DrawingTime is the amount of seconds that each player has available to
	// finish their drawing.
	DrawingTime int `json:"drawingTime"`
	// Rounds defines how many iterations a lobby does before the game ends.
	// One iteration means every participant does one drawing.
	Rounds int `json:"rounds"`
}

type gameState string

const (
	// Unstarted means the lobby has been opened but never started.
	Unstarted gameState = "unstarted"
	// Ongoing means the lobby has already been started.
	Ongoing gameState = "ongoing"
	// GameOver means that the lobby had been start, but the max round limit
	// has already been reached.
	GameOver gameState = "gameOver"
)

// WordHint describes a character of the word that is to be guessed, whether
// the character should be shown and whether it should be underlined on the
// UI.
type WordHint struct {
	Character rune `json:"character"`
	Underline bool `json:"underline"`
}

// RGBColor represents a 24-bit color consisting of red, green and blue.
type RGBColor struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// Line is the struct that a client send when drawing
type Line struct {
	FromX     float32  `json:"fromX"`
	FromY     float32  `json:"fromY"`
	ToX       float32  `json:"toX"`
	ToY       float32  `json:"toY"`
	Color     RGBColor `json:"color"`
	LineWidth float32  `json:"lineWidth"`
}

// Fill represents the usage of the fill bucket.
type Fill struct {
	X     float32  `json:"x"`
	Y     float32  `json:"y"`
	Color RGBColor `json:"color"`
}

// MaxPlayerNameLength defines how long a string can be at max when used
// as the playername.
const MaxPlayerNameLength int = 30

// Player represents a participant in a Lobby.
type Player struct {
	// userSession uniquely identifies the player.
	userSession      string
	ws               *websocket.Conn
	socketMutex      *sync.Mutex
	lastKnownAddress string
	// disconnectTime is used to kick a player in case the lobby doesn't have
	// space for new players. The player with the oldest disconnect.Time will
	// get kicked.
	disconnectTime *time.Time

	votedForKick map[string]bool

	// ID uniquely identified the Player.
	ID string `json:"id"`
	// Name is the players displayed name
	Name string `json:"name"`
	// Score is the points that the player got in the current Lobby.
	Score int `json:"score"`
	// Connected defines whether the players websocket connection is currently
	// established. This has previously been in state but has been moved out
	// in order to avoid losing the state on refreshing the page.
	// While checking the websocket against nil would be enough, we still need
	// this field for sending it via the APIs.
	Connected bool `json:"connected"`
	// Rank is the current ranking of the player in his Lobby
	LastScore int         `json:"lastScore"`
	Rank      int         `json:"rank"`
	State     PlayerState `json:"state"`
}

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
func (player *Player) GetWebsocket() *websocket.Conn {
	return player.ws
}

// SetWebsocket sets the given connection as the players websocket connection.
func (player *Player) SetWebsocket(socket *websocket.Conn) {
	player.ws = socket
}

// GetWebsocketMutex returns a mutex for locking the websocket connection.
// Since gorilla websockets shits it self when two calls happen at
// the same time, we need a mutex per player, since each player has their
// own socket. This getter extends to prevent accidentally sending the mutex
// via the network.
func (player *Player) GetWebsocketMutex() *sync.Mutex {
	return player.socketMutex
}

// GetUserSession returns the players current user session.
func (player *Player) GetUserSession() string {
	return player.userSession
}

type PlayerState string

const (
	Guessing PlayerState = "guessing"
	Drawing  PlayerState = "drawing"
	Standby  PlayerState = "standby"
)

// GetPlayer searches for a player, identifying them by usersession.
func (lobby *Lobby) GetPlayer(userSession string) *Player {
	for _, player := range lobby.players {
		if player.userSession == userSession {
			return player
		}
	}

	return nil
}

func (lobby *Lobby) ClearDrawing() {
	lobby.currentDrawing = make([]interface{}, 0)
}

// AppendLine adds a line direction to the current drawing. This exists in order
// to prevent adding arbitrary elements to the drawing, as the backing array is
// an empty interface type.
func (lobby *Lobby) AppendLine(line *LineEvent) {
	lobby.currentDrawing = append(lobby.currentDrawing, line)
}

// AppendFill adds a fill direction to the current drawing. This exists in order
// to prevent adding arbitrary elements to the drawing, as the backing array is
// an empty interface type.
func (lobby *Lobby) AppendFill(fill *FillEvent) {
	lobby.currentDrawing = append(lobby.currentDrawing, fill)
}

func createPlayer(name string) *Player {
	return &Player{
		Name:         SanitizeName(name),
		ID:           uuid.Must(uuid.NewV4()).String(),
		userSession:  uuid.Must(uuid.NewV4()).String(),
		votedForKick: make(map[string]bool),
		socketMutex:  &sync.Mutex{},
		State:        Guessing,
	}
}

//SanitizeName removes invalid characters from the players name, resolves
//emoji codes, limits the name length and generates a new name if necessary.
func SanitizeName(name string) string {
	//We trim and handle emojis beforehand to avoid taking this into account
	//when checking the name length, so we don't cut off too much of the name.
	newName := discordemojimap.Replace(strings.TrimSpace(name))

	//We don't want super-long names
	if len(newName) > MaxPlayerNameLength {
		return newName[:MaxPlayerNameLength+1]
	}

	if newName != "" {
		return newName
	}

	return generatePlayerName()
}

// GameEvent contains an eventtype and optionally any data.
type GameEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// GetConnectedPlayerCount returns the amount of player that have currently
// established a socket connection.
func (lobby *Lobby) GetConnectedPlayerCount() int {
	var count int
	for _, player := range lobby.players {
		if player.Connected {
			count++
		}
	}

	return count
}

func (lobby *Lobby) HasConnectedPlayers() bool {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	return lobby.hasConnectedPlayersInternal()
}

func (lobby *Lobby) hasConnectedPlayersInternal() bool {
	for _, otherPlayer := range lobby.players {
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
	return lobby.players
}

// GetOccupiedPlayerSlots counts the available slots which can be taken by new
// players. Whether a slot is available is determined by the player count and
// whether a player is disconnect or furthermore how long they have been
// disconnected for. Therefore the result of this function will differ from
// Lobby.GetConnectedPlayerCount.
func (lobby *Lobby) GetOccupiedPlayerSlots() int {
	var occupiedPlayerSlots int
	now := time.Now()
	for _, player := range lobby.players {
		if player.Connected {
			occupiedPlayerSlots++
		} else {
			disconnectTime := player.disconnectTime

			//If a player hasn't been disconnected for a certain
			//timeframe, we will reserve the slot. This avoids frustration
			//in situations where a player has to restart their PC or so.
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
	if len(lobby.players) < lobby.MaxPlayers {
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
