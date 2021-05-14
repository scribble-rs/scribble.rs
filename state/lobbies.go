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

func init() {
	//Task to clean up empty lobbies. An empty lobby is a lobby where all
	//players have been disconnected for a certain timeframe. This avoids
	//deleting lobbies when the creator of a lobby accidentally reconnects
	//or needs to refresh. Another scenario might be where the server loses
	//it's connection to all players temporarily. While unlikely, we'll be
	//able to preserve lobbies this way.
	go func() {
		lobbyCleanupTicker := time.NewTicker(90 * time.Second)
		for {
			<-lobbyCleanupTicker.C

			globalStateMutex.Lock()

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

			globalStateMutex.Unlock()
		}
	}()
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
	lobby := lobbies[indexToDelete]
	lobbies = append(lobbies[:indexToDelete], lobbies[indexToDelete+1:]...)
	log.Printf("Closing lobby %s. There are currently %d open lobbies left.\n", lobby.LobbyID, len(lobbies))
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
