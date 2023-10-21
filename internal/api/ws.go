package api

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/mailru/easyjson"

	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/state"
)

var (
	ErrPlayerNotConnected = errors.New("player not connected")

	upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin:       func(_ *http.Request) bool { return true },
		EnableCompression: true,
	}
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

	lobby := state.GetLobby(chi.URLParam(request, "lobby_id"))
	if lobby == nil {
		http.Error(writer, ErrLobbyNotExistent.Error(), http.StatusNotFound)
		return
	}

	lobby.Synchronized(func() {
		player := lobby.GetPlayer(userSession)
		if player == nil {
			http.Error(writer, "you don't have access to this lobby;usersession unknown", http.StatusUnauthorized)
			return
		}

		socket, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s(%s) has connected\n", player.Name, player.ID)

		player.SetWebsocket(socket)
		lobby.OnPlayerConnectUnsynchronized(player)

		socket.SetCloseHandler(func(code int, text string) error {
			lobby.OnPlayerDisconnect(player)
			return nil
		})

		go wsListen(lobby, player, socket)
	})
}

func wsListen(lobby *game.Lobby, player *game.Player, socket *websocket.Conn) {
	// Workaround to prevent crash, since not all kind of
	// disconnect errors are cleanly caught by gorilla websockets.
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error occurred in wsListen.\n\tError: %s\n\tPlayer: %s(%s)\nStack %s\n", err, player.Name, player.ID, string(debug.Stack()))
			lobby.OnPlayerDisconnect(player)
		}
	}()

	var event game.EventTypeOnly

	for {
		messageType, data, err := socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				lobby.OnPlayerDisconnect(player)
				return
			}

			// This way, we should catch repeated reads on closed connections
			// on both linux and windows. Previously we did this by searching
			// for certain text in the error message, which was neither
			// cross-platform nor translation aware.
			if netErr, ok := err.(*net.OpError); ok && !netErr.Temporary() {
				lobby.OnPlayerDisconnect(player)
				return
			}

			log.Printf("Error reading from socket: %s\n", err)
			// If the error doesn't seem fatal we attempt listening for more messages.
			continue
		}

		if messageType == websocket.TextMessage {
			if err := easyjson.Unmarshal(data, &event); err != nil {
				log.Printf("Error unmarshalling message: %s\n", err)
				err := WriteObject(player, game.Event{
					Type: game.EventTypeSystemMessage,
					Data: fmt.Sprintf("error parsing message, please report this issue via Github: %s!", err),
				})
				if err != nil {
					log.Printf("Error sending errormessage: %s\n", err)
				}
				continue
			}

			if err := lobby.HandleEvent(event.Type, data, player); err != nil {
				log.Printf("Error handling event: %s\n", err)
			}
		}
	}
}

func WriteObject(player *game.Player, object easyjson.Marshaler) error {
	player.GetWebsocketMutex().Lock()
	defer player.GetWebsocketMutex().Unlock()

	socket := player.GetWebsocket()
	if socket == nil || !player.Connected {
		return ErrPlayerNotConnected
	}

	bytes, err := easyjson.Marshal(object)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	return socket.WriteMessage(websocket.TextMessage, bytes)
}

func WritePreparedMessage(player *game.Player, message *websocket.PreparedMessage) error {
	player.GetWebsocketMutex().Lock()
	defer player.GetWebsocketMutex().Unlock()

	socket := player.GetWebsocket()
	if socket == nil || !player.Connected {
		return ErrPlayerNotConnected
	}

	return socket.WritePreparedMessage(message)
}
