package game

import (
	"testing"
)

func createLobbyWithDemoPlayers(playercount int) *Lobby {
	owner := &Player{}
	lobby := &Lobby{
		owner:   owner,
		creator: owner,
	}
	for i := 0; i < playercount; i++ {
		lobby.players = append(lobby.players, &Player{
			Connected: true,
		})
	}

	return lobby
}

func Test_CalculateVotesNeededToKick(t *testing.T) {
	t.Run("Check necessary kick vote amount for players", func(test *testing.T) {
		var expectedResults = map[int]int{
			//Kinda irrelevant since you can't kick yourself, but who cares.
			1:  1,
			2:  1,
			3:  2,
			4:  2,
			5:  3,
			6:  3,
			7:  4,
			8:  4,
			9:  5,
			10: 5,
		}

		for playerCount, expctedRequiredVotes := range expectedResults {
			lobby := createLobbyWithDemoPlayers(playerCount)
			result := calculateVotesNeededToKick(nil, lobby)
			if result != expctedRequiredVotes {
				t.Errorf("Error. Necessary vote amount was %d, but should've been %d", result, expctedRequiredVotes)
			}
		}
	})
}

func Test_RemoveAccents(t *testing.T) {
	t.Run("Check removing accented characters", func(test *testing.T) {
		var expectedResults = map[string]string{
			"é":           "e",
			"É":           "E",
			"à":           "a",
			"À":           "A",
			"ç":           "c",
			"Ç":           "C",
			"ö":           "oe",
			"Ö":           "OE",
			"œ":           "oe",
			"\n":          "\n",
			"\t":          "\t",
			"\r":          "\r",
			"":            "",
			"·":           "·",
			"?:!":         "?:!",
			"ac-ab":       "acab",
			"ac - _ab-- ": "acab",
		}

		for k, v := range expectedResults {
			result := removeAccents(k)
			if result != v {
				t.Errorf("Error. Char was %s, but should've been %s", result, v)
			}
		}
	})
}
