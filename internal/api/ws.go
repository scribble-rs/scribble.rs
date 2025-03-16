package api

import (
	json "encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/lxzan/gws"

	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/metrics"
	"github.com/scribble-rs/scribble.rs/internal/state"
)

var (
	ErrPlayerNotConnected = errors.New("player not connected")

	upgrader = gws.NewUpgrader(&socketHandler{}, &gws.ServerOption{
		Recovery:          gws.Recovery,
		ParallelEnabled:   true,
		PermessageDeflate: gws.PermessageDeflate{Enabled: true},
	})
)

func (handler *V1Handler) websocketUpgrade(writer http.ResponseWriter, request *http.Request) {
	userSession, err := GetUserSession(request)
	if err != nil {
		log.Printf("error getting user session: %v", err)
		http.Error(writer, "no valid usersession supplied", http.StatusBadRequest)
		return
	}

	if userSession == uuid.Nil {
		// This issue can happen if you illegally request a websocket
		// connection without ever having had a usersession or your
		// client having deleted the usersession cookie.
		http.Error(writer, "you don't have access to this lobby;usersession not set", http.StatusUnauthorized)
		return
	}

	lobbyId := GetLobbyId(request)
	if lobbyId == "" {
		http.Error(writer, "lobby id missing", http.StatusBadRequest)
		return
	}

	lobby := state.GetLobby(lobbyId)
	if lobby == nil {
		socket, err := upgrader.Upgrade(writer, request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		socket.WriteClose(1000, []byte("lobby_gone"))
		return
	}

	lobby.Synchronized(func() {
		player := lobby.GetPlayerBySession(userSession)
		if player == nil {
			http.Error(writer, "you don't have access to this lobby;usersession unknown", http.StatusUnauthorized)
			return
		}

		socket, err := upgrader.Upgrade(writer, request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		metrics.TrackPlayerConnect()

		player.SetWebsocket(socket)
		socket.Session().Store("player", player)
		socket.Session().Store("lobby", lobby)
		lobby.OnPlayerConnectUnsynchronized(player)

		go socket.ReadLoop()
	})
}

const (
	pingInterval = 10 * time.Second
	pingWait     = 5 * time.Second
)

type socketHandler struct{}

func (c *socketHandler) resetDeadline(socket *gws.Conn) {
	if err := socket.SetDeadline(time.Now().Add(pingInterval + pingWait)); err != nil {
		log.Printf("error resetting deadline: %s\n", err)
	}
}

func (c *socketHandler) OnOpen(socket *gws.Conn) {
	c.resetDeadline(socket)
}

func extract(x any, _ bool) any {
	return x
}

func (c *socketHandler) OnClose(socket *gws.Conn, _ error) {
	defer socket.Session().Delete("player")
	defer socket.Session().Delete("lobby")

	player, ok := extract(socket.Session().Load("player")).(*game.Player)
	if !ok {
		return
	}
	lobby, ok := extract(socket.Session().Load("lobby")).(*game.Lobby)
	if !ok {
		return
	}

	metrics.TrackPlayerDisconnect()
	lobby.OnPlayerDisconnect(player)
	player.SetWebsocket(nil)
}

func (c *socketHandler) OnPing(socket *gws.Conn, _ []byte) {
	c.resetDeadline(socket)
	_ = socket.WritePong(nil)
}

func (c *socketHandler) OnPong(socket *gws.Conn, _ []byte) {
	c.resetDeadline(socket)
}

func (c *socketHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	defer c.resetDeadline(socket)

	player, ok := extract(socket.Session().Load("player")).(*game.Player)
	if !ok {
		return
	}
	lobby, ok := extract(socket.Session().Load("lobby")).(*game.Lobby)
	if !ok {
		return
	}

	bytes := message.Bytes()
	handleIncommingEvent(lobby, player, bytes)
}

func handleIncommingEvent(lobby *game.Lobby, player *game.Player, data []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error occurred in incomming event listener.\n\tError: %s\n\tPlayer: %s(%s)\nStack %s\n", err, player.Name, player.ID, string(debug.Stack()))
			// FIXME Should this lead to a disconnect?
		}
	}()

	var event game.EventTypeOnly
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshalling message: %s\n", err)
		err := WriteObject(player, game.Event{
			Type: game.EventTypeSystemMessage,
			Data: fmt.Sprintf("error parsing message, please report this issue via Github: %s!", err),
		})
		if err != nil {
			log.Printf("Error sending errormessage: %s\n", err)
		}
		return
	}

	if err := lobby.HandleEvent(event.Type, data, player); err != nil {
		log.Printf("Error handling event: %s\n", err)
	}
}

func WriteObject(player *game.Player, object any) error {
	socket := player.GetWebsocket()
	if socket == nil || !player.Connected {
		return ErrPlayerNotConnected
	}

	bytes, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	// We write async, as broadcast always uses the queue. If we use write, the
	// order will become messed up, potentially causing issues in the frontend.
	socket.WriteAsync(gws.OpcodeText, bytes, func(err error) {
		if err != nil {
			log.Println("Error responding to player:", err.Error())
		}
	})
	return nil
}

func WritePreparedMessage(player *game.Player, message *gws.Broadcaster) error {
	socket := player.GetWebsocket()
	if socket == nil || !player.Connected {
		return ErrPlayerNotConnected
	}

	return message.Broadcast(socket)
}
