package state

import (
	"log"
	"sync"
	"time"

	"github.com/scribble-rs/scribble.rs/game"
)

var (
	globalStateMutex               = &sync.Mutex{}
	lobbies          []*game.Lobby = nil
)

// LaunchCleanupRoutine starts a task to clean up empty lobbies. An empty
// lobby is a lobby where all players have been disconnected for a certain
// timeframe. This avoids deleting lobbies when the creator of a lobby
// accidentally reconnects or needs to refresh. Another scenario might be
// where the server loses it's connection to all players temporarily. While
// unlikely, we'll be able to preserve lobbies this way.
// This method shouldn't be called more than once. Initially this was part of
// this packages init method, however, in order to avoid side effects in
// tests, this has been moved into a public function that has to be called
// manually.
func LaunchCleanupRoutine() {
	go func() {
		lobbyCleanupTicker := time.NewTicker(90 * time.Second)
		for {
			<-lobbyCleanupTicker.C
			cleanupRoutineLogic()
		}
	}()
}

// cleanupRoutineLogic is an extra function in order to prevent deadlocks by
// being able to use defer mutex.Unlock().
func cleanupRoutineLogic() {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	for index := len(lobbies) - 1; index >= 0; index-- {
		lobby := lobbies[index]
		if lobby.HasConnectedPlayers() {
			continue
		}

		disconnectTime := lobby.LastPlayerDisconnectTime
		if disconnectTime == nil || time.Since(*disconnectTime) >= 75*time.Second {
			removeLobbyByIndex(index)
		}
	}
}

// AddLobby adds a lobby to the instance, making it visible for GetLobby calls.
func AddLobby(lobby *game.Lobby) {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	lobbies = append(lobbies, lobby)
}

// GetLobby returns a Lobby that has a matching ID or no Lobby if none could
// be found.
func GetLobby(id string) *game.Lobby {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	for _, l := range lobbies {
		if l.LobbyID == id {
			return l
		}
	}

	return nil
}

// ShutdownLobbiesGracefully shuts down all lobbies and removes them from the
// state, preventing reconnects to existing lobbies. New lobbies can
// technically still be added.
func ShutdownLobbiesGracefully() {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	for _, lobby := range lobbies {
		//Since a reconnect requires a lookup to the state, all attempts to
		//reconnect will end up running into the global statelock. Therefore,
		//reconnecting wouldn't be possible.
		lobby.Shutdown()
	}

	//Instead of removing one by one, we nil the array, since that's faster.
	lobbies = nil
}

// GetActiveLobbyCount indicates how many activate lobby there are. This includes
// both private and public lobbies and it doesn't matter whether the game is
// already over, hasn't even started or is still ongoing.
func GetActiveLobbyCount() int {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	return len(lobbies)
}

// GetPublicLobbies returns all lobbies with their public flag set to true.
// This implies that the lobbies can be found in the lobby browser ob the
// homepage.
func GetPublicLobbies() []*game.Lobby {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	var publicLobbies []*game.Lobby
	for _, lobby := range lobbies {
		if lobby.IsPublic() {
			publicLobbies = append(publicLobbies, lobby)
		}
	}

	return publicLobbies
}

// RemoveLobby deletes a lobby, not allowing anyone to connect to it again.
func RemoveLobby(id string) {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	removeLobby(id)
}

func removeLobby(id string) {
	indexToDelete := -1
	for index, l := range lobbies {
		if l.LobbyID == id {
			indexToDelete = index
			break
		}
	}

	if indexToDelete != -1 {
		removeLobbyByIndex(indexToDelete)
	}
}

func removeLobbyByIndex(indexToDelete int) {
	lobbyID := lobbies[indexToDelete].LobbyID

	//We delete the lobby without maintaining order, since the lobby order
	//is irrelevant. This holds true as long as there's no paging for
	//requesting lobbies via the API.
	lobbies[indexToDelete] = lobbies[len(lobbies)-1]
	//Unreference the moved item in the other slot to prevent potential
	//memory leaks.
	lobbies[len(lobbies)-1] = nil
	lobbies = lobbies[:len(lobbies)-1]

	log.Printf("Closing lobby %s. There are currently %d open lobbies left.\n", lobbyID, len(lobbies))
}

// pageStats represents dynamic information about the website.
type pageStats struct {
	ActiveLobbyCount        int    `json:"activeLobbyCount"`
	PlayersCount            uint64 `json:"playersCount"`
	OccupiedPlayerSlotCount uint64 `json:"occupiedPlayerSlotCount"`
	ConnectedPlayersCount   uint64 `json:"connectedPlayersCount"`
}

// Stats delivers information about the state of the service. Currently this
// is lobby and player counts.
func Stats() *pageStats {
	globalStateMutex.Lock()
	defer globalStateMutex.Unlock()

	var playerCount, occupiedPlayerSlotCount, connectedPlayerCount uint64
	//While one would expect locking the lobby here, it's not very
	//important to get 100% consistent results here.
	for _, lobby := range lobbies {
		playerCount += uint64(len(lobby.GetPlayers()))
		occupiedPlayerSlotCount += uint64(lobby.GetOccupiedPlayerSlots())
		connectedPlayerCount += uint64(lobby.GetConnectedPlayerCount())
	}

	return &pageStats{
		ActiveLobbyCount:        len(lobbies),
		PlayersCount:            playerCount,
		OccupiedPlayerSlotCount: occupiedPlayerSlotCount,
		ConnectedPlayersCount:   connectedPlayerCount,
	}
}
