package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/scribble-rs/scribble.rs/internal/game"
)

var (
	ErrPlayerNotConnected = errors.New("player not connected")

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

func wsEndpoint(writer http.ResponseWriter, request *http.Request) {
	sessionCookie := GetUserSession(request)
	if sessionCookie == "" {
		// This issue can happen if you illegally request a websocket
		// connection without ever having had a usersession or your
		// client having deleted the usersession cookie.
		http.Error(writer, "you don't have access to this lobby;usersession not set", http.StatusUnauthorized)
		return
	}

	lobby, err := GetLobby(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}

	lobby.Synchronized(func() {
		player := lobby.GetPlayer(sessionCookie)
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

	for {
		messageType, data, err := socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) ||
				// This happens when the server closes the connection. It will cause 1000 retries followed by a panic.
				strings.Contains(err.Error(), "use of closed network connection") {
				// Make sure that the sockethandler is called
				lobby.OnPlayerDisconnect(player)
				// If the error is fatal, we stop listening for more messages.
				return
			}

			log.Printf("Error reading from socket: %s\n", err)
			// If the error doesn't seem fatal we attempt listening for more messages.
			continue
		}

		if messageType == websocket.TextMessage {
			received := &game.Event{}

			if err := json.Unmarshal(data, received); err != nil {
				log.Printf("Error unmarshalling message: %s\n", err)
				err := WriteJSON(player, game.Event{
					Type: "system-message",
					Data: fmt.Sprintf("error parsing message, please report this issue via Github: %s!", err),
				})
				if err != nil {
					log.Printf("Error sending errormessage: %s\n", err)
				}
				continue
			}

			if err := lobby.HandleEvent(received, player); err != nil {
				log.Printf("Error handling event: %s\n", err)
			}
		}
	}
}

// WriteJSON marshals the given input into a JSON string and sends it to the
// player using the currently established websocket connection.
func WriteJSON(player *game.Player, object any) error {
	player.GetWebsocketMutex().Lock()
	defer player.GetWebsocketMutex().Unlock()

	socket := player.GetWebsocket()
	if socket == nil || !player.Connected {
		return ErrPlayerNotConnected
	}

	return socket.WriteJSON(object)
}
