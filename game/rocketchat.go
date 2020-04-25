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

func updateRocketChat(lobby *Lobby, player *Player) {
	state := "connected"
	count := 0
	// Only count connected players
	for _, p := range lobby.Players {
		if p.Connected {
			count++
		}
	}
	if !player.Connected {
		state = "disconnected"
	}
	if count == 0 {
		sendRocketChatMessage(fmt.Sprintf("%v has %v. The game has ended.", player.Name, state))
		return
	}
	scribbleURL, exists := os.LookupEnv("SCRIBBLE_URL")
	if !exists {
		log.Printf("WARNING: SCRIBBLE_URL not set. Unable to send RocketChat messages")
		return
	}
	sendRocketChatMessage(fmt.Sprintf("%v has %v. There are %v players in the game. Join [here](%v/ssrEnterLobby?lobby_id=%v)", player.Name, state, count, scribbleURL, lobby.ID))
}
func sendRocketChatMessage(msg string) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	rocketchatWebhook, exists := os.LookupEnv("ROCKETCHAT_WEBHOOK")
	if !exists {
		log.Printf("WARNING: ROCKETCHAT_WEBHOOK not set. Unable to send RocketChat messages")
		return
	}
	payload := rocketChatPayload{
		Alias: "Scribble Bot",
		Text:  msg,
	}
	payloadByte, err := json.Marshal(payload)
	_, err = netClient.Post(rocketchatWebhook, "application/json", bytes.NewReader(payloadByte))
	if err != nil {
		log.Printf("%v", err)
		return
	}
}
