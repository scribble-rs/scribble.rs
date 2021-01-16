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
	game.TriggerUpdateEvent = TriggerUpdateEvent
	game.SendDataToEveryoneExceptSender = SendDataToEveryoneExceptSender
	game.WriteAsJSON = WriteAsJSON
	game.WritePublicSystemMessage = WritePublicSystemMessage
	game.TriggerUpdatePerPlayerEvent = TriggerUpdatePerPlayerEvent
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
			log.Printf("Error occurred in wsListen.\n\tError: %s\n\tPlayer: %s(%s)\n", err, player.Name, player.ID)
			game.OnDisconnected(lobby, player)
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
				return
			}

			log.Printf("Error reading from socket: %s\n", err)
		} else if messageType == websocket.TextMessage {
			received := &game.GameEvent{}
			err := json.Unmarshal(data, received)
			if err != nil {
				log.Printf("Error unmarshalling message: %s\n", err)
				sendError := WriteAsJSON(player, game.GameEvent{Type: "system-message", Data: fmt.Sprintf("An error occurred trying to read your request, please report the error via GitHub: %s!", err)})
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

func SendDataToEveryoneExceptSender(sender *game.Player, lobby *game.Lobby, data interface{}) {
	for _, otherPlayer := range lobby.GetPlayers() {
		if otherPlayer != sender {
			WriteAsJSON(otherPlayer, data)
		}
	}
}

func TriggerUpdateEvent(eventType string, data interface{}, lobby *game.Lobby) {
	event := &game.GameEvent{Type: eventType, Data: data}
	for _, otherPlayer := range lobby.GetPlayers() {
		WriteAsJSON(otherPlayer, event)
	}
}

func TriggerUpdatePerPlayerEvent(eventType string, data func(*game.Player) interface{}, lobby *game.Lobby) {
	for _, otherPlayer := range lobby.GetPlayers() {
		WriteAsJSON(otherPlayer, &game.GameEvent{Type: eventType, Data: data(otherPlayer)})
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
	systemMessageEvent := &game.GameEvent{Type: "system-message", Data: html.EscapeString(text)}
	for _, otherPlayer := range lobby.GetPlayers() {
		//In simple message events we ignore write failures.
		WriteAsJSON(otherPlayer, systemMessageEvent)
	}
}
