package communication

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/scribble-rs/scribble.rs/game"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func init() {
	game.TriggerComplexUpdateEvent = TriggerComplexUpdateEvent
	game.TriggerSimpleUpdateEvent = TriggerSimpleUpdateEvent
	game.SendDataToConnectedPlayers = SendDataToConnectedPlayers
	game.WriteAsJSON = WriteAsJSON
	game.WritePublicSystemMessage = WritePublicSystemMessage
	game.TriggerComplexUpdatePerPlayerEvent = TriggerComplexUpdatePerPlayerEvent
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	lobbyID := r.URL.Query().Get("id")
	if lobbyID == "" {
		http.Error(w, "the requested lobby doesn't exist", http.StatusNotFound)
		return
	}

	lobby := game.GetLobby(lobbyID)
	if lobby == nil {
		http.Error(w, "the requested lobby doesn't exist", http.StatusNotFound)
		return
	}

	sessionCookie, noCookieError := r.Cookie("usersession")
	//This issue can happen if you illegally request a websocket connection without ever having had
	//a usersession or your client having deleted the usersession cookie.
	if noCookieError != nil {
		http.Error(w, "the requested lobby doesn't exist", http.StatusUnauthorized)
		return
	}

	player := lobby.GetPlayer(sessionCookie.Value)
	if player == nil {
		http.Error(w, "the requested lobby doesn't exist", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(player.Name + " has connected")

	player.SetWebsocket(ws)
	game.OnConnected(lobby, player)

	ws.SetCloseHandler(func(code int, text string) error {
		game.OnDisconnected(lobby, player)
		return nil
	})

	go wsListen(lobby, player, ws)
}

func wsListen(lobby *game.Lobby, player *game.Player, socket *websocket.Conn) {
	for {
		messageType, data, err := socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) ||
				//This happens when the server closes the connection. It will cause 1000 retries followed by a panic.
				strings.Contains(err.Error(), "use of closed network connection") {
				//Make sure that the sockethandler is called
				socket.CloseHandler()
				log.Println(player.Name + " disconnected.")
				return
			} else {
				log.Printf("Error reading from socket: %s\n", err)
			}
		} else if messageType == websocket.TextMessage {
			received := &game.JSEvent{}
			err := json.Unmarshal(data, received)
			if err != nil {
				log.Printf("Error unmarshalling message: %s\n", err)
				sendError := WriteAsJSON(player, game.JSEvent{Type: "system-message", Data: fmt.Sprintf("An error occured trying to read your request, please report the error via GitHub: %s!", err)})
				if sendError != nil {
					log.Printf("Error sending errormessage: %s\n", sendError)
				}
				continue
			}

			handleError := game.HandleEvent(data, received, lobby, player)
			if handleError != nil {
				log.Printf("Error handling event: %s\n", handleError)
			}
		}
	}
}

func SendDataToConnectedPlayers(sender *game.Player, lobby *game.Lobby, data interface{}) {
	for _, otherPlayer := range lobby.Players {
		if otherPlayer != sender {
			WriteAsJSON(otherPlayer, data)
		}
	}
}

func TriggerSimpleUpdateEvent(eventType string, lobby *game.Lobby) {
	event := &game.JSEvent{Type: eventType}
	for _, otherPlayer := range lobby.Players {
		//FIXME Why did i use a goroutine here but not anywhere else?
		go func(player *game.Player) {
			WriteAsJSON(player, event)
		}(otherPlayer)
	}
}

func TriggerComplexUpdateEvent(eventType string, data interface{}, lobby *game.Lobby) {
	event := &game.JSEvent{Type: eventType, Data: data}
	for _, otherPlayer := range lobby.Players {
		WriteAsJSON(otherPlayer, event)
	}
}

func TriggerComplexUpdatePerPlayerEvent(eventType string, data func(*game.Player) interface{}, lobby *game.Lobby) {
	for _, otherPlayer := range lobby.Players {
		WriteAsJSON(otherPlayer, &game.JSEvent{Type: eventType, Data: data(otherPlayer)})
	}
}

// WriteAsJSON marshals the given input into a JSON string and sends it to the
// player using the currently established websocket connection.
func WriteAsJSON(player *game.Player, object interface{}) error {
	player.GetWebsocketMutex().Lock()
	defer player.GetWebsocketMutex().Unlock()

	socket := player.GetWebsocket()
	if socket == nil || player.State == game.Disconnected {
		return errors.New("player not connected")
	}

	return socket.WriteJSON(object)
}

func WritePublicSystemMessage(lobby *game.Lobby, text string) {
	playerHasBeenKickedMsg := &game.JSEvent{Type: "system-message", Data: html.EscapeString(text)}
	for _, otherPlayer := range lobby.Players {
		//In simple message events we ignore write failures.
		WriteAsJSON(otherPlayer, playerHasBeenKickedMsg)
	}
}
