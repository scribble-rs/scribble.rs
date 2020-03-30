package game

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

// Lobby represents a game session.
type Lobby struct {
	// ID uniquely identified the Lobby.
	ID string

	// DrawingTime is the amount of seconds that each player has available to
	// finish their drawing.
	DrawingTime int
	// Rounds defines how many iterations a lobby does before the game ends.
	// One iteration means every participant does one drawing.
	Rounds int
	// MaxPlayers defines the maximum amount of players in a single lobby.
	MaxPlayers int
	// CustomWords are additional words that will be used in addition to the
	// predefined words.
	CustomWords []string
	Words       []string

	// Players references all participants of the Lobby.
	Players []*Player

	// Drawer references the Player that is currently drawing.
	Drawer *Player
	// Owner references the Player that created the lobby.
	Owner *Player
	// CurrentWord represents the word that was last selected. If no word has
	// been selected yet or the round is already over, this should be empty.
	CurrentWord string
	// WordHints for the current word.
	WordHints []*WordHint
	// WordHintsShown are the same as WordHints with characters visible.
	WordHintsShown []*WordHint
	// Round is the round that the Lobby is currently in. This is a number
	// between 0 and Rounds. 0 indicates that it hasn't started yet.
	Round int
	// WordChoice represents the current choice of words.
	WordChoice []string
	// TimeLeft is the current TimeLeft during a turn.
	TimeLeft int

	timeLeftTicker        *time.Ticker
	timeLeftTickerReset   chan struct{}
	scoreEarnedByGuessers int
	alreadyUsedWords      []string
	CustomWordsChance     int
	ClientsPerIPLimit     int
	CurrentDrawing        []*Pixel
	EnableVotekick        bool
}

// WordHint describes a character of the word that is to be guessed, whether
// the character should be shown and whether it should be underlined on the
// UI.
type WordHint struct {
	Character string
	Show      bool
	Underline bool
}

// Pixel is the struct that a client send when drawing
type Pixel struct {
	FromX     float32
	FromY     float32
	ToX       float32
	ToY       float32
	Color     string
	LineWidth float32
	Type      string // either "pixel" or "fill"
}

// Player represents a participant in a Lobby.
type Player struct {
	// UserSession uniquely identifies the player.
	UserSession string
	//Ws is a reference to the players websocket connection.
	Ws *websocket.Conn
	// Since gorilla websockets shits it self when two calls happen at
	// the same time, we need a mutex per player, since each player has their
	// own socket.
	SocketMutex *sync.Mutex

	// ID uniquely identified the Player.
	ID string
	// Name is the players displayed name
	Name string
	// Score is the points that the player got in the current Lobby.
	Score int
	// Rank is the current ranking of the player in his Lobby
	LastScore    int
	Rank         int
	State        PlayerState
	Icon         string
	votedForKick map[string]bool
}

type PlayerState int

const (
	Guessing     PlayerState = 0
	Drawing      PlayerState = 1
	Standby      PlayerState = 2
	Disconnected PlayerState = 3
)

// GetPlayer searches for a player, identifying them by usersession.
func (lobby *Lobby) GetPlayer(userSession string) *Player {
	for _, player := range lobby.Players {
		if player.UserSession == userSession {
			return player
		}
	}

	return nil
}

func (lobby *Lobby) ClearDrawing() {
	lobby.CurrentDrawing = []*Pixel{}
}

func (lobby *Lobby) AppendPixel(pixel *Pixel) {
	lobby.CurrentDrawing = append(lobby.CurrentDrawing, pixel)
}

// GetLobby returns a Lobby that has a matching ID or no Lobby if none could
// be found.
func GetLobby(id string) *Lobby {
	for _, l := range lobbies {
		if l.ID == id {
			return l
		}
	}

	return nil
}

// RemoveLobby deletes a lobby, not allowing anyone to connect to it again.
func RemoveLobby(id string) {
	indexToDelete := -1
	for index, l := range lobbies {
		if l.ID == id {
			indexToDelete = index
		}
	}

	if indexToDelete != -1 {
		lobbies = append(lobbies[:indexToDelete], lobbies[indexToDelete+1:]...)
	}
}

func createPlayer(name string) *Player {
	return &Player{
		Name:         name,
		ID:           uuid.NewV4().String(),
		UserSession:  uuid.NewV4().String(),
		Score:        0,
		LastScore:    0,
		Rank:         1,
		votedForKick: make(map[string]bool),
		SocketMutex:  &sync.Mutex{},
		State:        Disconnected,
	}
}

func createLobby(
	drawingTime int,
	rounds int,
	maxPlayers int,
	customWords []string,
	customWordsChance int,
	clientsPerIPLimit int,
	enableVotekick bool) *Lobby {

	createDeleteMutex.Lock()

	lobby := &Lobby{
		ID:                  uuid.NewV4().String(),
		DrawingTime:         drawingTime,
		Rounds:              rounds,
		MaxPlayers:          maxPlayers,
		CustomWords:         customWords,
		CustomWordsChance:   customWordsChance,
		timeLeftTickerReset: make(chan struct{}),
		ClientsPerIPLimit:   clientsPerIPLimit,
		EnableVotekick:      enableVotekick,
		CurrentDrawing:      []*Pixel{},
	}

	if len(customWords) > 1 {
		rand.Shuffle(len(lobby.CustomWords), func(i, j int) {
			lobby.CustomWords[i], lobby.CustomWords[j] = lobby.CustomWords[j], lobby.CustomWords[i]
		})
	}

	lobbies = append(lobbies, lobby)

	createDeleteMutex.Unlock()

	return lobby
}

// JSEvent contains an eventtype and optionally any data.
type JSEvent struct {
	Type string
	Data interface{}
}

func (lobby *Lobby) HasConnectedPlayers() bool {
	for _, otherPlayer := range lobby.Players {
		if otherPlayer.Ws != nil && otherPlayer.State != Disconnected {
			return true
		}
	}

	return false
}
