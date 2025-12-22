package game

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/lxzan/gws"
)

//
// This file contains all structs and constants that are shared with clients.
//

// Eevnts that are just incomming from the client.
const (
	EventTypeStart           = "start"
	EventTypeToggleReadiness = "toggle-readiness"
	EventTypeToggleSpectate  = "toggle-spectate"
	EventTypeRequestDrawing  = "request-drawing"
	EventTypeChooseWord      = "choose-word"
	EventTypeUndo            = "undo"
)

// Events that are outgoing only.
const (
	EventTypeUpdatePlayers            = "update-players"
	EventTypeUpdateWordHint           = "update-wordhint"
	EventTypeWordChosen               = "word-chosen"
	EventTypeCorrectGuess             = "correct-guess"
	EventTypeCloseGuess               = "close-guess"
	EventTypeSystemMessage            = "system-message"
	EventTypeNonGuessingPlayerMessage = "non-guessing-player-message"
	EventTypeReady                    = "ready"
	EventTypeGameOver                 = "game-over"
	EventTypeYourTurn                 = "your-turn"
	EventTypeNextTurn                 = "next-turn"
	EventTypeDrawing                  = "drawing"
	EventTypeDrawerKicked             = "drawer-kicked"
	EventTypeOwnerChange              = "owner-change"
	EventTypeLobbySettingsChanged     = "lobby-settings-changed"
	EventTypeShutdown                 = "shutdown"
	EventTypeKeepAlive                = "keep-alive"
)

// Events that are bidirectional.
var (
	EventTypeKickVote          = "kick-vote"
	EventTypeNameChange        = "name-change"
	EventTypeMessage           = "message"
	EventTypeLine              = "line"
	EventTypeFill              = "fill"
	EventTypeClearDrawingBoard = "clear-drawing-board"
)

type State string

const (
	// Unstarted means the lobby has been opened but never started.
	Unstarted State = "unstarted"
	// Ongoing means the lobby has already been started.
	Ongoing State = "ongoing"
	// GameOver means that the lobby had been start, but the max round limit
	// has already been reached.
	GameOver State = "gameOver"
)

// Event contains an eventtype and optionally any data.
type Event struct {
	Data any    `json:"data"`
	Type string `json:"type"`
}

type StringDataEvent struct {
	Data string `json:"data"`
}

type EventTypeOnly struct {
	Type string `json:"type"`
}

type IntDataEvent struct {
	Data int `json:"data"`
}

// WordHint describes a character of the word that is to be guessed, whether
// the character should be shown and whether it should be underlined on the
// UI.
type WordHint struct {
	Character rune `json:"character"`
	Underline bool `json:"underline"`
}

type LineEvent struct {
	Type string `json:"type"`
	// Data contains the coordinates, stroke width and color. The coors here
	// aren't uint16, as it allows us to easily allow implementing drawing on
	// the client, where the user drags the line over the canvas border.
	// If we were to not accept out of bounds values, the lines would be chopped
	// off before reaching the canvas border.
	Data struct {
		X  int16 `json:"x"`
		Y  int16 `json:"y"`
		X2 int16 `json:"x2"`
		Y2 int16 `json:"y2"`
		// Color is a color index. This was previously an rgb value, but since
		// the values are always the same, using an index saves bandwidth.
		Color uint8 `json:"color"`
		Width uint8 `json:"width"`
	} `json:"data"`
}

type FillEvent struct {
	Data *struct {
		X uint16 `json:"x"`
		Y uint16 `json:"y"`
		// Color is a color index. This was previously an rgb value, but since
		// the values are always the same, using an index saves bandwidth.
		Color uint8 `json:"color"`
	} `json:"data"`
	Type string `json:"type"`
}

// KickVote represents a players vote to kick another players. If the VoteCount
// is as great or greater than the RequiredVoteCount, the event indicates a
// successful kick vote. The voting is anonymous, meaning the voting player
// won't be exposed.
type KickVote struct {
	PlayerName        string    `json:"playerName"`
	PlayerID          uuid.UUID `json:"playerId"`
	VoteCount         int       `json:"voteCount"`
	RequiredVoteCount int       `json:"requiredVoteCount"`
}

type OwnerChangeEvent struct {
	PlayerName string    `json:"playerName"`
	PlayerID   uuid.UUID `json:"playerId"`
}

type NameChangeEvent struct {
	PlayerName string    `json:"playerName"`
	PlayerID   uuid.UUID `json:"playerId"`
}

// GameOverEvent is basically the ready event, but contains the last word.
// This is required in order to show the last player the word, in case they
// didn't manage to guess it in time. This is necessary since the last word
// is usually part of the "next-turn" event, which we don't send, since the
// game is over already.
type GameOverEvent struct {
	*ReadyEvent
	PreviousWord string `json:"previousWord"`
}

type WordChosen struct {
	TimeLeft int         `json:"timeLeft"`
	Hints    []*WordHint `json:"hints"`
}

type YourTurn struct {
	TimeLeft        int      `json:"timeLeft"`
	PreSelectedWord int      `json:"preSelectedWord"`
	Words           []string `json:"words"`
}

// NextTurn represents the data necessary for displaying the lobby state right
// after a new turn started. Meaning that no word has been chosen yet and
// therefore there are no wordhints and no current drawing instructions.
type NextTurn struct {
	// PreviousWord signals the last chosen word. If empty, no word has been
	// chosen. The client can now themselves whether there has been a previous
	// turn, by looking at the current gamestate.
	PreviousWord   string    `json:"previousWord"`
	Players        []*Player `json:"players"`
	ChoiceTimeLeft int       `json:"choiceTimeLeft"`
	Round          int       `json:"round"`
}

// OutgoingMessage represents a message in the chatroom.
type OutgoingMessage struct {
	// Content is the actual message text.
	Content string `json:"content"`
	// Author is the player / thing that wrote the message
	Author string `json:"author"`
	// AuthorID is the unique identifier of the authors player object.
	AuthorID uuid.UUID `json:"authorId"`
}

// ReadyEvent represents the initial state that a user needs upon connection.
// This includes all the necessary things for properly running a client
// without receiving any more data.
type ReadyEvent struct {
	WordHints          []*WordHint `json:"wordHints"`
	PlayerName         string      `json:"playerName"`
	Players            []*Player   `json:"players"`
	GameState          State       `json:"gameState"`
	CurrentDrawing     []any       `json:"currentDrawing"`
	PlayerID           uuid.UUID   `json:"playerId"`
	OwnerID            uuid.UUID   `json:"ownerId"`
	Round              int         `json:"round"`
	Rounds             int         `json:"rounds"`
	TimeLeft           int         `json:"timeLeft"`
	DrawingTimeSetting int         `json:"drawingTimeSetting"`
	AllowDrawing       bool        `json:"allowDrawing"`
}

// Player represents a participant in a Lobby.
type Player struct {
	// userSession uniquely identifies the player.
	userSession uuid.UUID
	ws          *gws.Conn
	// disconnectTime is used to kick a player in case the lobby doesn't have
	// space for new players. The player with the oldest disconnect.Time will
	// get kicked.
	disconnectTime   *time.Time
	votedForKick     map[uuid.UUID]bool
	lastKnownAddress string
	// messageTimestamps tracks the timestamps of recent messages for rate limiting.
	// Stores up to 30 timestamps (max messages in 20 seconds).
	messageTimestamps []time.Time

	// Name is the players displayed name
	Name  string      `json:"name"`
	State PlayerState `json:"state"`
	// SpectateToggleRequested is used for state changes between spectator and
	// player. We want to prevent people from switching in and out of the Player
	// state. While this will allow people to skip being the drawer, it will
	// also cause them to lose points for that round.
	SpectateToggleRequested bool `json:"spectateToggleRequested"`
	// Rank is the current ranking of the player in his Lobby
	// Score is the points that the player got in the current Lobby.
	Score     int `json:"score"`
	LastScore int `json:"lastScore"`
	Rank      int `json:"rank"`
	// Connected defines whether the players websocket connection is currently
	// established. This has previously been in state but has been moved out
	// in order to avoid losing the state on refreshing the page.
	// While checking the websocket against nil would be enough, we still need
	// this field for sending it via the APIs.
	Connected bool `json:"connected"`
	// ID uniquely identified the Player.
	ID uuid.UUID `json:"id"`
}

// EditableLobbySettings represents all lobby settings that are editable by
// the lobby owner after the lobby has already been opened.
type EditableLobbySettings struct {
	// CustomWords are additional words that will be used in addition to the
	// predefined words.
	// Public defines whether the lobby is being broadcast to clients asking
	// for available lobbies.
	Public bool `json:"public"`
	// MaxPlayers defines the maximum amount of players in a single lobby.
	MaxPlayers         int `json:"maxPlayers"`
	CustomWordsPerTurn int `json:"customWordsPerTurn"`
	// ClientsPerIPLimit helps preventing griefing by reducing each player
	// to one tab per IP address.
	ClientsPerIPLimit int `json:"clientsPerIpLimit"`
	// Rounds defines how many iterations a lobby does before the game ends.
	// One iteration means every participant does one drawing.
	Rounds int `json:"rounds"`
	// DrawingTime is the amount of seconds that each player has available to
	// finish their drawing.
	DrawingTime int `json:"drawingTime"`
}
