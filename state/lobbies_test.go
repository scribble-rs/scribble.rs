package state

import (
	"testing"

	"github.com/scribble-rs/scribble.rs/game"
)

func TestAddAndRemove(t *testing.T) {
	if len(lobbies) > 0 {
		t.Error("Lobbies should have been empty initially.")
	}

	lobby := &game.Lobby{
		LobbyID: "WTF",
	}

	AddLobby(lobby)

	if GetLobby(lobby.LobbyID) == nil {
		t.Error("Lobby should've been found.")
	}

	RemoveLobby(lobby.LobbyID)

	if GetLobby(lobby.LobbyID) != nil {
		t.Error("Lobby shouldn't have been found.")
	}

	if len(lobbies) != 0 {
		t.Error("Lobbies should have been empty after removal.")
	}
}
