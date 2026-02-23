package state

import (
	"testing"

	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest //this test is very stateful
func TestAddAndRemove(t *testing.T) {
	require.Empty(t, lobbies, "Lobbies should be empty when test starts")

	createLobby := func() *game.Lobby {
		player, lobby, err := game.CreateLobby("", "player", "dutch", &game.EditableLobbySettings{
			Public:             true,
			DrawingTime:        100,
			Rounds:             10,
			MaxPlayers:         10,
			CustomWordsPerTurn: 3,
			ClientsPerIPLimit:  1,
			WordsPerTurn:       3,
		}, nil, game.ChillScoring)
		require.NoError(t, err)
		lobby.OnPlayerDisconnect(player)
		return lobby
	}
	lobbyA := createLobby()
	lobbyB := createLobby()
	lobbyC := createLobby()

	AddLobby(lobbyA)
	AddLobby(lobbyB)
	AddLobby(lobbyC)

	require.NotNil(t, GetLobby(lobbyA.LobbyID))
	require.NotNil(t, GetLobby(lobbyB.LobbyID))
	require.NotNil(t, GetLobby(lobbyC.LobbyID))

	RemoveLobby(lobbyB.LobbyID)
	require.Nil(t, GetLobby(lobbyB.LobbyID), "Lobby B should have been deleted.")

	require.NotNil(t, GetLobby(lobbyA.LobbyID), "Lobby A shouldn't have been deleted.")
	require.NotNil(t, GetLobby(lobbyC.LobbyID), "Lobby C shouldn't have been deleted.")
	require.Len(t, lobbies, 2)

	cleanupRoutineLogic(&config.LobbyCleanup{})
	require.Empty(t, lobbies)
}
