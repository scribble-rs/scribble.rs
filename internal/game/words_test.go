package game

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Test_wordListsContainNoCarriageReturns(t *testing.T) {
	t.Parallel()

	for _, entry := range WordlistData {
		fileName := entry.LanguageCode
		fileBytes, err := wordFS.ReadFile("words/" + fileName)
		if err != nil {
			t.Errorf("language file '%s' could not be read: %s", fileName, err)
		} else if bytes.ContainsRune(fileBytes, '\r') {
			t.Errorf("language file '%s' contains a carriage return", fileName)
		}
	}
}

func Test_readWordList(t *testing.T) {
	t.Parallel()

	t.Run("test invalid language file", func(t *testing.T) {
		t.Parallel()

		words, err := readDefaultWordList(cases.Lower(language.English), "nonexistent")
		assert.ErrorIs(t, err, ErrUnknownWordList)
		assert.Empty(t, words)
	})

	for language := range WordlistData {
		language := language
		t.Run(language, func(t *testing.T) {
			t.Parallel()

			testWordList(t, language)
			testWordList(t, language)
		})
	}
}

func testWordList(t *testing.T, chosenLanguage string) {
	t.Helper()

	lowercaser := WordlistData[chosenLanguage].Lowercaser()
	words, err := readDefaultWordList(lowercaser, chosenLanguage)
	if err != nil {
		t.Errorf("Error reading language %s: %s", chosenLanguage, err)
	}

	if len(words) == 0 {
		t.Errorf("Wordlist for language %s was empty.", chosenLanguage)
	}

	for _, word := range words {
		if word == "" {
			// We can't print the faulty line, since we are shuffling
			// the words in order to avoid predictability.
			t.Errorf("Wordlist for language %s contained empty word", chosenLanguage)
		}

		if strings.TrimSpace(word) != word {
			t.Errorf("Word has surrounding whitespace characters: '%s'", word)
		}

		if lowercaser.String(word) != word {
			t.Errorf("Word hasn't been lowercased: '%s'", word)
		}
	}
}

func Test_getRandomWords(t *testing.T) {
	t.Parallel()

	t.Run("Test getRandomWords with 3 words in list", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			CurrentWord: "",
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 0,
			},
			words: []string{"a", "b", "c"},
			mutex: &sync.Mutex{},
		}

		randomWords := GetRandomWords(3, lobby)
		for _, lobbyWord := range lobby.words {
			if !arrayContains(randomWords, lobbyWord) {
				t.Errorf("Random words %s, didn't contain lobbyWord %s", randomWords, lobbyWord)
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 3 more in custom word list, but with 0 chance", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			CurrentWord: "",
			words:       []string{"a", "b", "c"},
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 0,
			},

			CustomWords: []string{"d", "e", "f"},
			mutex:       &sync.Mutex{},
		}

		randomWords := GetRandomWords(3, lobby)
		for _, lobbyWord := range lobby.words {
			if !arrayContains(randomWords, lobbyWord) {
				t.Errorf("Random words %s, didn't contain lobbyWord %s", randomWords, lobbyWord)
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 100% custom word chance, but without custom words", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			CurrentWord: "",
			words:       []string{"a", "b", "c"},
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 3,
			},
			CustomWords: nil,
			mutex:       &sync.Mutex{},
		}

		randomWords := GetRandomWords(3, lobby)
		for _, lobbyWord := range lobby.words {
			if !arrayContains(randomWords, lobbyWord) {
				t.Errorf("Random words %s, didn't contain lobbyWord %s", randomWords, lobbyWord)
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 100% custom word chance, with 3 custom words", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			CurrentWord: "",
			words:       []string{"a", "b", "c"},
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 3,
			},
			CustomWords: []string{"d", "e", "f"},
			mutex:       &sync.Mutex{},
		}

		randomWords := GetRandomWords(3, lobby)
		for _, customWord := range lobby.CustomWords {
			if !arrayContains(randomWords, customWord) {
				t.Errorf("Random words %s, didn't contain customWord %s", randomWords, customWord)
			}
		}
	})
}

func Test_getRandomWordsReloading(t *testing.T) {
	t.Parallel()

	loadWordList := func() []string { return []string{"a", "b", "c"} }
	reloadWordList := func(_ *Lobby) ([]string, error) { return loadWordList(), nil }
	wordList := loadWordList()

	t.Run("test reload with 3 words and 0 custom words and 0 chance", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			words: wordList,
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 0,
			},
			CustomWords: nil,
			mutex:       &sync.Mutex{},
		}

		// Running this 10 times, expecting it to get 3 words each time, even
		// though our pool has only got a size of 3.
		for i := 0; i < 10; i++ {
			words := getRandomWords(3, lobby, reloadWordList)
			if len(words) != 3 {
				t.Errorf("Test failed, incorrect wordcount: %d", len(words))
			}
		}
	})

	t.Run("test reload with 3 words and 0 custom words and 100 chance", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			words: wordList,
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 3,
			},
			CustomWords: nil,
			mutex:       &sync.Mutex{},
		}

		// Running this 10 times, expecting it to get 3 words each time, even
		// though our pool has only got a size of 3.
		for i := 0; i < 10; i++ {
			words := getRandomWords(3, lobby, reloadWordList)
			if len(words) != 3 {
				t.Errorf("Test failed, incorrect wordcount: %d", len(words))
			}
		}
	})

	t.Run("test reload with 3 words and 1 custom words and 0 chance", func(t *testing.T) {
		t.Parallel()

		lobby := &Lobby{
			words: wordList,
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsPerTurn: 3,
			},
			CustomWords: []string{"a"},
			mutex:       &sync.Mutex{},
		}

		// Running this 10 times, expecting it to get 3 words each time, even
		// though our pool has only got a size of 3.
		for i := 0; i < 10; i++ {
			words := getRandomWords(3, lobby, reloadWordList)
			if len(words) != 3 {
				t.Errorf("Test failed, incorrect wordcount: %d", len(words))
			}
		}
	})
}

func arrayContains(array []string, item string) bool {
	for _, arrayItem := range array {
		if arrayItem == item {
			return true
		}
	}

	return false
}

var poximityBenchCases = [][]string{
	{"", ""},
	{"a", "a"},
	{"ab", "ab"},
	{"abc", "abc"},
	{"abc", "abcde"},
	{"cde", "abcde"},
	{"a", "abcdefghijklmnop"},
	{"cde", "abcde"},
	{"cheese", "wheel"},
	{"abcdefg", "bcdefgh"},
}

func Benchmark_proximity_custom(b *testing.B) {
	for _, benchCase := range poximityBenchCases {
		b.Run(fmt.Sprint(benchCase[0], " ", benchCase[1]), func(b *testing.B) {
			var sink int
			for i := 0; i < b.N; i++ {
				sink = CheckGuess(benchCase[0], benchCase[1])
			}
			_ = sink
		})
	}
}

// We've replaced levensthein with the implementation from proximity_custom
// func Benchmark_proximity_levensthein(b *testing.B) {
// 	for _, benchCase := range poximityBenchCases {
// 		b.Run(fmt.Sprint(benchCase[0], " ", benchCase[1]), func(b *testing.B) {
// 			var sink int
// 			for i := 0; i < b.N; i++ {
// 				sink = levenshtein.ComputeDistance(benchCase[0], benchCase[1])
// 			}
// 			_ = sink
// 		})
// 	}
// }

func Test_CheckGuess_Negative(t *testing.T) {
	t.Parallel()

	type testCase struct {
		a, b string
	}

	cases := []testCase{
		{
			a: "abc",
			b: "abcde",
		},
		{
			a: "abc",
			b: "01abc",
		},
		{
			a: "abc",
			b: "a",
		},
		{
			a: "abc",
			b: "c",
		},
		{
			a: "abc",
			b: "c",
		},
		{
			a: "hallo",
			b: "welt",
		},
		{
			a: "abcd",
			b: "badc",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s ~ %s", c.a, c.b), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, 2, CheckGuess(c.a, c.b))
		})
	}
}

func Test_CheckGuess_Positive(t *testing.T) {
	t.Parallel()

	type testCase struct {
		a, b string
		dist int
	}

	cases := []testCase{
		{
			a:    "abc",
			b:    "abc",
			dist: EqualGuess,
		},
		{
			a:    "abc",
			b:    "abcd",
			dist: CloseGuess,
		},
		{
			a:    "abc",
			b:    "ab",
			dist: CloseGuess,
		},
		{
			a:    "abc",
			b:    "bc",
			dist: CloseGuess,
		},
		{
			a:    "abcde",
			b:    "abde",
			dist: CloseGuess,
		},
		{
			a:    "abc",
			b:    "adc",
			dist: CloseGuess,
		},
		{
			a:    "abc",
			b:    "acb",
			dist: CloseGuess,
		},
		{
			a:    "abcd",
			b:    "acbd",
			dist: CloseGuess,
		},
		{
			a:    "abcd",
			b:    "bacd",
			dist: CloseGuess,
		},
		{
			a:    "cheese",
			b:    "wheel",
			dist: DistantGuess,
		},
		{
			a:    "a",
			b:    "bcdefg",
			dist: DistantGuess,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s ~ %s", c.a, c.b), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.dist, CheckGuess(c.a, c.b))
		})
	}
}
