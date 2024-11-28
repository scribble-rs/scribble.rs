package frontend

import (
	"testing"

	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
)

func TestCreateLobby(t *testing.T) {
	t.Parallel()

	data := api.CreateLobbyData(
		&config.Default,
		&game.Lobby{
			LobbyID: "TEST",
		})

	var previousSize uint8
	for _, suggestedSize := range data.SuggestedBrushSizes {
		if suggestedSize < previousSize {
			t.Error("Sorting in SuggestedBrushSizes is incorrect")
		}
	}

	for _, suggestedSize := range data.SuggestedBrushSizes {
		if suggestedSize < game.MinBrushSize {
			t.Errorf("suggested brushsize %d is below MinBrushSize %d", suggestedSize, game.MinBrushSize)
		}

		if suggestedSize > game.MaxBrushSize {
			t.Errorf("suggested brushsize %d is above MaxBrushSize %d", suggestedSize, game.MaxBrushSize)
		}
	}
}
