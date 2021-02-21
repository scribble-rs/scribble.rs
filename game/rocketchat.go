package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type rocketChatPayload struct {
	Alias string `json:"alias"`
	Text  string `json:"text"`
}

var (
	rocketchatWebhook string
	scribbleURL       string
	rootPath          string

	netTransport *http.Transport
	netClient    *http.Client
)

func init() {
	rocketchatWebhook, _ = os.LookupEnv("ROCKETCHAT_WEBHOOK")
	scribbleURL, _ = os.LookupEnv("SCRIBBLE_URL")
	rootPath, _ = os.LookupEnv("ROOT_PATH")

	if rocketchatWebhook == "" || scribbleURL == "" {
		//If one of these wasn't set up correctly, we needn't init the additional state.
		return
	}

	// Go doesn't set timeouts by default
	netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
}

func updateRocketChat(lobby *Lobby, player *Player) {
	//This means scribble wasn't set up correctly for use with rocket chat.
	if rocketchatWebhook == "" || scribbleURL == "" {
		return
	}

	var count int
	// Only count connected players
	for _, p := range lobby.GetPlayers() {
		if p.Connected {
			count++
		}
	}

	var action string
	if !player.Connected {
		action = "disconnected"
	} else {
		action = "connected"
	}

	//FIXME Technically not correct anymore, as the lobby is only
	//closed if no player reconnects within a certain time.
	if count == 0 {
		sendRocketChatMessage(fmt.Sprintf("%v has %v. The game has ended.", player.Name, action))
	} else {
		sendRocketChatMessage(fmt.Sprintf("%v has %v. There are %v players in the game. Join [here](%v%s/ssrEnterLobby?lobby_id=%v)", player.Name, action, count, scribbleURL, rootPath, lobby.LobbyID))
	}
}

func sendRocketChatMessage(msg string) {
	payload := rocketChatPayload{
		Alias: "Scribble Bot",
		Text:  msg,
	}

	payloadByte, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
	}

	_, err = netClient.Post(rocketchatWebhook, "application/json", bytes.NewReader(payloadByte))
	if err != nil {
		log.Println(err)
	}
}
