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
	lobby, lobbyError := getLobby(r)
	if lobbyError != nil {
		http.Error(w, lobbyError.Error(), http.StatusNotFound)
		return
	}

	//This issue can happen if you illegally request a websocket connection without ever having had
	//a usersession or your client having deleted the usersession cookie.
	sessionCookie := getUserSession(r)
	if sessionCookie == "" {
		http.Error(w, "you don't have access to this lobby;usersession not set", http.StatusUnauthorized)
		return
	}

	player := lobby.GetPlayer(sessionCookie)
	if player == nil {
		http.Error(w, "you don't have access to this lobby;usersession invalid", http.StatusUnauthorized)
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
	//Workaround to prevent crash
	defer func() {
		err := recover()
		if err != nil {
			game.OnDisconnected(lobby, player)
			log.Println("Error occurred in wsListen: ", err)
		}
	}()
	for {
		messageType, data, err := socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) ||
				//This happens when the server closes the connection. It will cause 1000 retries followed by a panic.
				strings.Contains(err.Error(), "use of closed network connection") {
				//Make sure that the sockethandler is called
				game.OnDisconnected(lobby, player)
				log.Println(player.Name + " disconnected.")
				return
			}

			log.Printf("Error reading from socket: %s\n", err)
		} else if messageType == websocket.TextMessage {
			received := &game.JSEvent{}
			err := json.Unmarshal(data, received)
			if err != nil {
				log.Printf("Error unmarshalling message: %s\n", err)
				sendError := WriteAsJSON(player, game.JSEvent{Type: "system-message", Data: fmt.Sprintf("An error occurred trying to read your request, please report the error via GitHub: %s!", err)})
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
	for _, allPlayers := range lobby.Players {
		if allPlayers != sender {
			WriteAsJSON(allPlayers, data)
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
	if socket == nil || !player.Connected {
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
