package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/scribble-rs/scribble.rs/internal/api"
)

type body struct {
	contentType string
	data        io.Reader
}

func request(method, url string, body *body, queryParameters map[string]any) (*http.Response, error) {
	var data io.Reader
	if body != nil {
		data = body.data
	}
	request, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, fmt.Errorf("error perparing request: %w", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", body.contentType)
	}
	query := request.URL.Query()
	for k, v := range queryParameters {
		query.Set(k, fmt.Sprint(v))
	}
	request.URL.RawQuery = query.Encode()

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error during http call: %w", err)
	}

	return response, nil
}

func PostLobby() (*api.LobbyData, error) {
	response, err := request(http.MethodPost, "http://localhost:8080/v1/lobby", nil, map[string]any{
		"language":              "english",
		"drawing_time":          120,
		"word_select_count":     5,
		"rounds":                4,
		"max_players":           12,
		"clients_per_ip_limit":  12,
		"custom_words_per_turn": 3,
		"public":                true,
		"timerStart":            false,
	})
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		fmt.Println(string(bytes))
	}
	var lobby api.LobbyData
	if err := json.NewDecoder(response.Body).Decode(&lobby); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &lobby, err
}

type SimPlayer struct {
	Id          string
	Name        string
	Usersession string
	ws          *websocket.Conn
	rand        *rand.Rand
}

func (s *SimPlayer) SendRandomStroke() {
	if err := s.ws.WriteJSON(map[string]any{
		"fromX": rand.Float64(),
		"fromY": rand.Float64(),
		"toX":   rand.Float64(),
		"toY":   rand.Float64(),
		"color": map[string]any{
			"r": rand.Intn(255),
			"g": rand.Intn(255),
			"b": rand.Intn(255),
		},
		"lineWidth": 5,
	}); err != nil {
		log.Println("error writing:", err)
	}
}

func (s *SimPlayer) SendRandomMessage() {
	if err := s.ws.WriteJSON(map[string]any{
		"type": "message",
		"data": uuid.Must(uuid.NewV4()).String(),
	}); err != nil {
		log.Println("error writing:", err)
	}
}

func JoinPlayer(lobbyId string) (*SimPlayer, error) {
	lobbyUrl := "localhost:8080/v1/lobby/" + lobbyId
	response, err := request(http.MethodPost, "http://"+lobbyUrl+"/player", nil, map[string]any{
		"username": uuid.Must(uuid.NewV4()).String(),
	})
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(bytes))
	}

	var session string
	for _, cookie := range response.Cookies() {
		if strings.EqualFold(cookie.Name, "Usersession") {
			session = cookie.Value
			break
		}
	}

	if session == "" {
		return nil, errors.New("no usersession")
	}

	dialer := *websocket.DefaultDialer
	dialer.Subprotocols = []string{"json"}
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	headers := make(http.Header)
	headers.Set("usersession", session)
	wsConnection, response, err := dialer.Dial("ws://"+lobbyUrl+"/ws", headers)
	if response != nil && response.StatusCode != http.StatusSwitchingProtocols {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(bytes))
	}
	if err != nil {
		return nil, fmt.Errorf("error establishing websocket connection: %w", err)
	}

	// Sink messages
	go func() {
		m := make(map[string]any)
		for {
			if err := wsConnection.ReadJSON(&m); err != nil {
				log.Println("error receiving")
			}
		}
	}()

	return &SimPlayer{
		Usersession: session,
		ws:          wsConnection,
		rand:        rand.New(rand.NewSource(time.Now().Unix())),
	}, nil
}

func main() {
	for i := 0; i < 10; i++ {
		lobby, err := PostLobby()
		if err != nil {
			panic(err)
		}
		log.Println("Lobby:", lobby.LobbyID)
	}

	// player, err := JoinPlayer(lobby.LobbyID)
	// if err != nil {
	// 	panic(err)
	// }

	// start := time.Now()
	// for i := 0; i < 1_000_000; i++ {
	// 	player.SendRandomStroke()
	// 	player.SendRandomMessage()
	// }

	// log.Println(time.Since(start).Seconds())
}
