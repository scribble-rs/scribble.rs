package game

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/gofrs/uuid"
	"github.com/lxzan/gws"
	easyjson "github.com/mailru/easyjson"
	"github.com/scribble-rs/scribble.rs/internal/sanitize"
)

func createLobbyWithDemoPlayers(playercount int) *Lobby {
	owner := &Player{}
	lobby := &Lobby{
		Owner:   owner,
		creator: owner,
		mutex:   &sync.Mutex{},
	}
	for i := 0; i < playercount; i++ {
		lobby.players = append(lobby.players, &Player{
			Connected: true,
		})
	}

	return lobby
}

func noOpWriteObject(_ *Player, _ easyjson.Marshaler) error {
	return nil
}

func noOpWritePreparedMessage(_ *Player, _ *gws.Broadcaster) error {
	return nil
}

func Test_CalculateVotesNeededToKick(t *testing.T) {
	t.Run("Check necessary kick vote amount for players", func(test *testing.T) {
		expectedResults := map[int]int{
			// Kinda irrelevant since you can't kick yourself, but who cares.
			1:  2,
			2:  2,
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
		expectedResults := map[string]string{
			"é":     "e",
			"É":     "E",
			"à":     "a",
			"À":     "A",
			"ç":     "c",
			"Ç":     "C",
			"ö":     "oe",
			"Ö":     "OE",
			"œ":     "oe",
			"\n":    "\n",
			"\t":    "\t",
			"\r":    "\r",
			"":      "",
			"·":     "·",
			"?:!":   "?:!",
			"ac-ab": "acab",
			//nolint:gocritic
			"ac - _ab-- ": "acab",
		}

		for k, v := range expectedResults {
			result := sanitize.CleanText(k)
			if result != v {
				t.Errorf("Error. Char was %s, but should've been %s", result, v)
			}
		}
	})
}

func Test_simplifyText(t *testing.T) {
	// We only test the replacement we do ourselves. We won't test
	// the "sanitize", or furthermore our expectations of it for now.

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "dash",
			input: "-",
			want:  "",
		},
		{
			name:  "underscore",
			input: "_",
			want:  "",
		},
		{
			name:  "space",
			input: " ",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitize.CleanText(tt.input); got != tt.want {
				t.Errorf("simplifyText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_recalculateRanks(t *testing.T) {
	lobby := &Lobby{
		mutex: &sync.Mutex{},
	}
	lobby.players = append(lobby.players, &Player{
		ID:        uuid.Must(uuid.NewV4()),
		Score:     1,
		Connected: true,
	})
	lobby.players = append(lobby.players, &Player{
		ID:        uuid.Must(uuid.NewV4()),
		Score:     1,
		Connected: true,
	})
	recalculateRanks(lobby)

	rankPlayerA := lobby.players[0].Rank
	rankPlayerB := lobby.players[1].Rank
	if rankPlayerA != 1 || rankPlayerB != 1 {
		t.Errorf("With equal score, ranks should be equal. (A: %d; B: %d)",
			rankPlayerA, rankPlayerB)
	}

	lobby.players = append(lobby.players, &Player{
		ID:        uuid.Must(uuid.NewV4()),
		Score:     0,
		Connected: true,
	})
	recalculateRanks(lobby)

	rankPlayerA = lobby.players[0].Rank
	rankPlayerB = lobby.players[1].Rank
	if rankPlayerA != 1 || rankPlayerB != 1 {
		t.Errorf("With equal score, ranks should be equal. (A: %d; B: %d)",
			rankPlayerA, rankPlayerB)
	}

	rankPlayerC := lobby.players[2].Rank
	if rankPlayerC != 2 {
		t.Errorf("new player should be rank 2, since the previous two players had the same rank. (C: %d)", rankPlayerC)
	}
}

func Test_calculateGuesserScore(t *testing.T) {
	lastScore := calculateGuesserScore(0, 0, 115, 120)
	if lastScore >= maxBaseScore {
		t.Errorf("Score should have declined, but was bigger than or "+
			"equal to the baseScore. (LastScore: %d; BaseScore: %d)", lastScore, maxBaseScore)
	}

	lastDecline := -1
	for secondsLeft := 105; secondsLeft >= 5; secondsLeft -= 10 {
		newScore := calculateGuesserScore(0, 0, secondsLeft, 120)
		if newScore > lastScore {
			t.Errorf("Score with more time taken should be lower. (LastScore: %d; NewScore: %d)", lastScore, newScore)
		}
		newDecline := lastScore - newScore
		if lastDecline != -1 && newDecline > lastDecline {
			t.Errorf("Decline should get lower with time taken. (LastDecline: %d; NewDecline: %d)\n", lastDecline, newDecline)
		}
		lastScore = newScore
		lastDecline = newDecline
	}
}

func Test_handleNameChangeEvent(t *testing.T) {
	lobby := &Lobby{}
	lobby.WriteObject = noOpWriteObject
	lobby.WritePreparedMessage = noOpWritePreparedMessage
	player := lobby.JoinPlayer("Kevin")

	handleNameChangeEvent(player, lobby, "Jim")

	expectedName := "Jim"
	if player.Name != expectedName {
		t.Errorf("playername didn't change; Expected %s, but was %s", expectedName, player.Name)
	}
}

func getUnexportedField(field reflect.Value) any {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func Test_wordSelectionEvent(t *testing.T) {
	lobby := &Lobby{
		mutex: &sync.Mutex{},
		EditableLobbySettings: EditableLobbySettings{
			DrawingTime:     10,
			WordSelectCount: 3,
			Rounds:          10,
		},
		words: []string{"abc", "def", "ghi"},
	}
	wordHintEvents := make(map[uuid.UUID][]*WordHint)
	var wordChoice []string
	lobby.WriteObject = func(player *Player, message easyjson.Marshaler) error {
		event, ok := message.(*Event)
		if ok {
			if event.Type == EventTypeYourTurn {
				wordChoice = event.Data.([]string)
			}
		}

		return nil
	}
	lobby.WritePreparedMessage = func(player *Player, message *gws.Broadcaster) error {
		data := getUnexportedField(reflect.ValueOf(message).Elem().FieldByName("payload")).([]byte)
		type event struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		var e event
		if err := json.Unmarshal(data, &e); err != nil {
			t.Fatal("error unmarshalling message", err)
		}

		t.Log(e.Type)
		if e.Type == "update-wordhint" {
			var wordHints []*WordHint
			if err := json.Unmarshal(e.Data, &wordHints); err != nil {
				t.Fatal("error unmarshalling word hints:", err)
			}
			wordHintEvents[player.ID] = wordHints
		}
		return nil
	}

	drawer := lobby.JoinPlayer("Drawer")
	drawer.Connected = true
	lobby.Owner = drawer
	lobby.creator = drawer

	if err := lobby.HandleEvent(EventTypeStart, nil, drawer); err != nil {
		t.Errorf("Couldn't start lobby: %s", err)
	}

	guesser := lobby.JoinPlayer("Guesser")
	guesser.Connected = true

	err := lobby.HandleEvent(EventTypeChooseWord, []byte(`{"data": 0}`), drawer)
	if err != nil {
		t.Errorf("Couldn't choose word: %s", err)
	}

	wordHintsForDrawer := wordHintEvents[drawer.ID]
	if len(wordHintsForDrawer) != 3 {
		t.Errorf("Word hints for drawer were of incorrect length; %d != %d", len(wordHintsForDrawer), 3)
	}

	for index, wordHint := range wordHintsForDrawer {
		if wordHint.Character == 0 {
			t.Error("Word hints for drawer contained invisible character")
		}

		if !wordHint.Underline {
			t.Error("Word hints for drawer contained not underlined character")
		}

		expectedRune := rune(wordChoice[0][index])
		if wordHint.Character != expectedRune {
			t.Errorf("Character at index %d was %c instead of %c", index, wordHint.Character, expectedRune)
		}
	}

	wordHintsForGuesser := wordHintEvents[guesser.ID]
	if len(wordHintsForGuesser) != 3 {
		t.Errorf("Word hints for guesser were of incorrect length; %d != %d", len(wordHintsForGuesser), 3)
	}

	for _, wordHint := range wordHintsForGuesser {
		if wordHint.Character != 0 {
			t.Error("Word hints for guesser contained visible character")
		}

		if !wordHint.Underline {
			t.Error("Word hints for guesser contained not underlined character")
		}
	}
}

func Test_kickDrawer(t *testing.T) {
	lobby := &Lobby{
		mutex: &sync.Mutex{},
		EditableLobbySettings: EditableLobbySettings{
			DrawingTime:     10,
			wordSelectCount: 3,
			Rounds:          10,
		},
		words: []string{"a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a"},
	}
	lobby.WriteObject = noOpWriteObject
	lobby.WritePreparedMessage = noOpWritePreparedMessage

	marcel := lobby.JoinPlayer("marcel")
	marcel.Connected = true
	lobby.Owner = marcel
	lobby.creator = marcel

	kevin := lobby.JoinPlayer("kevin")
	kevin.Connected = true
	chantal := lobby.JoinPlayer("chantal")
	chantal.Connected = true

	if err := lobby.HandleEvent(EventTypeStart, nil, marcel); err != nil {
		t.Errorf("Couldn't start lobby: %s", err)
	}

	if lobby.Drawer() == nil {
		t.Error("Drawer should've been a, but was nil")
	}

	if lobby.Drawer() != marcel {
		t.Errorf("Drawer should've been a, but was %s", lobby.Drawer().Name)
	}

	lobby.Synchronized(func() {
		advanceLobby(lobby)
	})

	if lobby.Drawer() == nil {
		t.Error("Drawer should've been b, but was nil")
	}

	if lobby.Drawer() != kevin {
		t.Errorf("Drawer should've been b, but was %s", lobby.Drawer().Name)
	}

	lobby.Synchronized(func() {
		kickPlayer(lobby, kevin, 1)
	})

	if lobby.Drawer() == nil {
		t.Error("Drawer should've been c, but was nil")
	}

	if lobby.Drawer() != chantal {
		t.Errorf("Drawer should've been c, but was %s", lobby.Drawer().Name)
	}
}
