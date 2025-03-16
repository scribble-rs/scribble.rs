package game

import (
	"testing"
	"time"
)

func TestOccupiedPlayerCount(t *testing.T) {
	t.Parallel()

	lobby := &Lobby{}
	if lobby.GetOccupiedPlayerSlots() != 0 {
		t.Errorf("Occupied player count expected to be 0, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	// While disconnect, there's no disconnect time, which we count as occupied.
	lobby.Players = append(lobby.Players, &Player{})
	if lobby.GetOccupiedPlayerSlots() != 1 {
		t.Errorf("Occupied player count expected to be 1, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	lobby.Players = append(lobby.Players, &Player{
		Connected: true,
	})
	if lobby.GetOccupiedPlayerSlots() != 2 {
		t.Errorf("Occupied player count expected to be 2, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	disconnectedPlayer := &Player{
		Connected: false,
	}
	lobby.Players = append(lobby.Players, disconnectedPlayer)
	if lobby.GetOccupiedPlayerSlots() != 3 {
		t.Errorf("Occupied player count expected to be 3, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	now := time.Now()
	disconnectedPlayer.disconnectTime = &now
	if lobby.GetOccupiedPlayerSlots() != 3 {
		t.Errorf("Occupied player count expected to be 3, but was %d", lobby.GetOccupiedPlayerSlots())
	}

	past := time.Now().AddDate(-1, 0, 0)
	disconnectedPlayer.disconnectTime = &past
	if lobby.GetOccupiedPlayerSlots() != 2 {
		t.Errorf("Occupied player count expected to be 2, but was %d", lobby.GetOccupiedPlayerSlots())
	}
}
