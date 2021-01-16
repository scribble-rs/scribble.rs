package state

import (
	"log"
	"sync"
	"time"

	"github.com/scribble-rs/scribble.rs/game"
)

var (
	createDeleteMutex               = &sync.Mutex{}
	lobbies           []*game.Lobby = nil
)

func init() {
	//Task to clean up lobbies.
	go func() {
		lobbyCleanupTicker := time.NewTicker(90 * time.Second)
		for {
			<-lobbyCleanupTicker.C

			createDeleteMutex.Lock()

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

			createDeleteMutex.Unlock()
		}
	}()
}

// AddLobby adds a lobby to the instance, making it visible for GetLobby calls.
func AddLobby(lobby *game.Lobby) {
	createDeleteMutex.Lock()
	defer createDeleteMutex.Unlock()

	lobbies = append(lobbies, lobby)
}

// GetLobby returns a Lobby that has a matching ID or no Lobby if none could
// be found.
func GetLobby(id string) *game.Lobby {
	createDeleteMutex.Lock()
	defer createDeleteMutex.Unlock()

	for _, l := range lobbies {
		if l.ID == id {
			return l
		}
	}

	return nil
}

// GetActiveLobbyCount indicates how many activate lobby there are. This includes
// both private and public lobbies and it doesn't matter whether the game is
// already over, hasn't even started or is still ongoing.
func GetActiveLobbyCount() int {
	createDeleteMutex.Lock()
	defer createDeleteMutex.Unlock()

	return len(lobbies)
}

// GetPublicLobbies returns all lobbies with their public flag set to true.
// This implies that the lobbies can be found in the lobby browser ob the
// homepage.
func GetPublicLobbies() []*game.Lobby {
	createDeleteMutex.Lock()
	defer createDeleteMutex.Unlock()

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
	createDeleteMutex.Lock()
	defer createDeleteMutex.Unlock()

	removeLobby(id)
}

func removeLobby(id string) {
	indexToDelete := -1
	for index, l := range lobbies {
		if l.ID == id {
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
	log.Printf("Closing lobby %s. There are currently %d open lobbies left.\n", lobby.ID, len(lobbies))
}
