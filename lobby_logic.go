package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

var (
	createDeleteMutex = &sync.Mutex{}
	lobbies           = []*Lobby{}
)

// Player represents a participant in a Lobby.
type Player struct {
	// UserSession uniquely identifies the player.
	UserSession string
	ws          *websocket.Conn
	// Since gorilla websockets shits it self when two calls happen at
	// the same time, we need a mutex per player, since each player has their
	// own socket.
	socketMutex *sync.Mutex

	// ID uniquely identified the Player.
	ID string
	// Name is the players displayed name
	Name string
	// Score is the points that the player got in the current Lobby.
	Score int
	// Rank is the current ranking of the player in his Lobby
	LastScore     int
	Rank          int
	State         PlayerState
	Icon          string
	votedForKick  map[string]bool
	voteKickCount int
}

type PlayerState int

const (
	Guessing     PlayerState = 0
	Drawing      PlayerState = 1
	Standby      PlayerState = 2
	Disconnected PlayerState = 3
)

// Lobby represents a game session.
type Lobby struct {
	// ID uniquely identified the Lobby.
	ID string

	// Password defines the password that participants need to enter before
	// being allowed to join the Lobby.
	Password string
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
	clientsPerIPLimit     int
}

// GetPlayer searches for a player, identifying them by usersssion.
func (lobby *Lobby) GetPlayer(userSession string) *Player {
	for _, player := range lobby.Players {
		if player.UserSession == userSession {
			return player
		}
	}

	return nil
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

func createLobby(
	password string,
	drawingTime int,
	rounds int,
	maxPlayers int,
	customWords []string,
	customWordsChance int,
	clientsPerIPLimit int) *Lobby {

	createDeleteMutex.Lock()

	lobby := &Lobby{
		ID:                  uuid.NewV4().String(),
		Password:            password,
		DrawingTime:         drawingTime,
		Rounds:              rounds,
		MaxPlayers:          maxPlayers,
		CustomWords:         customWords,
		CustomWordsChance:   customWordsChance,
		timeLeftTickerReset: make(chan struct{}),
		clientsPerIPLimit:   clientsPerIPLimit,
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

func createPlayer(name string) *Player {
	return &Player{
		Name:          name,
		ID:            uuid.NewV4().String(),
		UserSession:   uuid.NewV4().String(),
		Score:         0,
		LastScore:     0,
		Rank:          1,
		voteKickCount: 0,
		votedForKick:  make(map[string]bool),
		socketMutex:   &sync.Mutex{},
	}
}

// WriteAsJSON marshals the given input into a JSON string and sends it to the
// player using the currently established websocket connection.
func (p *Player) WriteAsJSON(object interface{}) error {
	if p.ws == nil || p.State == Disconnected {
		return errors.New("player not connected")
	}

	p.socketMutex.Lock()
	err := p.ws.WriteJSON(object)
	p.socketMutex.Unlock()

	return err
}
