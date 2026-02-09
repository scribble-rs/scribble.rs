package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/lxzan/gws"
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
		"rounds":                4,
		"max_players":           12,
		"clients_per_ip_limit":  12,
		"custom_words_per_turn": 3,
		"public":                true,
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
	ws          *gws.Conn
}

func (s *SimPlayer) WriteJSON(value any) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.ws.WriteMessage(gws.OpcodeText, bytes)
}

func (s *SimPlayer) SendRandomStroke() {
	if err := s.WriteJSON(map[string]any{
		"fromX": rand.Float64(),
		"fromY": rand.Float64(),
		"toX":   rand.Float64(),
		"toY":   rand.Float64(),
		"color": map[string]any{
			"r": rand.Int32N(255),
			"g": rand.Int32N(255),
			"b": rand.Int32N(255),
		},
		"lineWidth": 5,
	}); err != nil {
		log.Println("error writing:", err)
	}
}

func (s *SimPlayer) SendRandomMessage() {
	if err := s.WriteJSON(map[string]any{
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

	headers := make(http.Header)
	headers.Set("usersession", session)
	wsConnection, response, err := gws.NewClient(gws.BuiltinEventHandler{}, &gws.ClientOption{
		Addr:          "ws://" + lobbyUrl + "/ws",
		RequestHeader: headers,
	})
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

	return &SimPlayer{
		Usersession: session,
		ws:          wsConnection,
	}, nil
}

func main() {
	// lobby, err := PostLobby()
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println("Lobby:", lobby.LobbyID)

	player, err := JoinPlayer("4cf284ff-8e18-4dc7-86a9-d2c2ed14227f")
	if err != nil {
		panic(err)
	}

	start := time.Now()
	for range 1_000_000 {
		player.SendRandomStroke()
		player.SendRandomMessage()
	}

	log.Println(time.Since(start).Seconds())
}
