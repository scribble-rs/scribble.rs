package game

import (
	"sync"
	"testing"
	"time"
)

func TestOccupiedPlayerCount(t *testing.T) {
	lobby := &Lobby{
		mutex: &sync.Mutex{},
	}
	if lobby.GetOccupiedPlayerSlots() != 0 {
		t.Errorf("Occupied player count expected to be 0, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	//While disconnect, there's no disconnect time, which we count as occupied.
	lobby.players = append(lobby.players, &Player{})
	if lobby.GetOccupiedPlayerSlots() != 1 {
		t.Errorf("Occupied player count expected to be 1, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	lobby.players = append(lobby.players, &Player{
		Connected: true,
	})
	if lobby.GetOccupiedPlayerSlots() != 2 {
		t.Errorf("Occupied player count expected to be 2, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	disconnectedPlayer := &Player{
		Connected: false,
	}
	lobby.players = append(lobby.players, disconnectedPlayer)
	if lobby.GetOccupiedPlayerSlots() != 3 {
		t.Errorf("Occupied player count expected to be 3, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	now := time.Now()
	disconnectedPlayer.disconnectTime = &now
	if lobby.GetOccupiedPlayerSlots() != 3 {
		t.Errorf("Occupied player count expected to be 3, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	past := time.Now().Local().AddDate(-1, 0, 0)
	disconnectedPlayer.disconnectTime = &past
	if lobby.GetOccupiedPlayerSlots() != 2 {
		t.Errorf("Occupied player count expected to be 2, but was %d", lobby.GetOccupiedPlayerSlots())
	}

}
