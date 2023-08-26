package game

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Test_wordListsContainNoCarriageReturns(t *testing.T) {
	for _, fileName := range languageIdentifiers {
		fileBytes, err := wordFS.ReadFile("words/" + fileName)
		if err != nil {
			t.Errorf("language file '%s' could not be read: %s", fileName, err)
		} else if bytes.ContainsRune(fileBytes, '\r') {
			t.Errorf("language file '%s' contains a carriage return", fileName)
		}
	}
}

func Test_readWordList(t *testing.T) {
	t.Run("test invalid language file", func(t *testing.T) {
		_, err := readWordList(cases.Lower(language.English), "owO")
		if err == nil {
			t.Errorf("Reading word list didn't return an error, even though the language doesn't exist.")
		}
	})

	for language := range languageIdentifiers {
		t.Run(fmt.Sprintf("Testing language file from embedded data for %s", language), func(t *testing.T) {
			testWordList(t, language)
		})
		t.Run(fmt.Sprintf("Testing language file from in-memory cached data for %s", language), func(t *testing.T) {
			testWordList(t, language)
		})
	}
}

func testWordList(t *testing.T, chosenLanguage string) {
	t.Helper()

	lowercaser := cases.Lower(language.English)
	words, err := readWordList(lowercaser, chosenLanguage)
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
	t.Run("Test getRandomWords with 3 words in list", func(t *testing.T) {
		lobby := &Lobby{
			CurrentWord: "",
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 0,
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
		lobby := &Lobby{
			CurrentWord: "",
			words:       []string{"a", "b", "c"},
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 0,
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
		lobby := &Lobby{
			CurrentWord: "",
			words:       []string{"a", "b", "c"},
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 100,
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
		lobby := &Lobby{
			CurrentWord: "",
			words:       []string{"a", "b", "c"},
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 100,
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

func Test_regressionGetRandomWords_singleCustomWord(t *testing.T) {
	lobby := &Lobby{
		CurrentWord: "",
		EditableLobbySettings: EditableLobbySettings{
			CustomWordsChance: 99,
		},
		CustomWords: []string{"custom"},
		mutex:       &sync.Mutex{},
	}

	words := make([]string, 99)
	for i := 0; i < 99; i++ {
		words[i] = strconv.FormatInt(int64(i), 10)
	}
	lobby.words = words

	// The implementation detail is, that the chance has to be smaller than
	// or equal to our random number. So the only number possible to return
	// our only custom word, is 1.
	if getRandomWordsCustomRng(1, lobby, func() int { return 1 })[0] != "custom" {
		t.Error("Custom should've been found, but wasn't.")
	}

	// Now furthermore, we expect 2 - 100 to give us non-custom words and not panic.
	for i := 2; i <= 100; i++ {
		if getRandomWordsCustomRng(1, lobby, func() int { return i })[0] == "custom" {
			t.Error("Custom word was found but shouldn't have.")
		}
	}
}

func Test_getRandomWordsReloading(t *testing.T) {
	wordList, err := readWordListInternal(cases.Lower(language.English), "test", func(language string) (string, error) {
		return "a\nb\nc", nil
	})
	if err != nil {
		panic(err)
	}

	t.Run("test reload with 3 words and 0 custom words and 0 chance", func(t2 *testing.T) {
		lobby := &Lobby{
			words: wordList,
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 0,
			},
			CustomWords: nil,
			mutex:       &sync.Mutex{},
		}

		// Running this 10 times, expecting it to get 3 words each time, even
		// though our pool has only got a size of 3.
		for i := 0; i < 10; i++ {
			words := GetRandomWords(3, lobby)
			if len(words) != 3 {
				t.Errorf("Test failed, incorrect wordcount: %d", len(words))
			}
		}
	})

	t.Run("test reload with 3 words and 0 custom words and 100 chance", func(t2 *testing.T) {
		lobby := &Lobby{
			words: wordList,
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 100,
			},
			CustomWords: nil,
			mutex:       &sync.Mutex{},
		}

		// Running this 10 times, expecting it to get 3 words each time, even
		// though our pool has only got a size of 3.
		for i := 0; i < 10; i++ {
			words := GetRandomWords(3, lobby)
			if len(words) != 3 {
				t.Errorf("Test failed, incorrect wordcount: %d", len(words))
			}
		}
	})

	t.Run("test reload with 3 words and 1 custom words and 0 chance", func(t2 *testing.T) {
		lobby := &Lobby{
			words: wordList,
			EditableLobbySettings: EditableLobbySettings{
				CustomWordsChance: 100,
			},
			CustomWords: []string{"a"},
			mutex:       &sync.Mutex{},
		}

		// Running this 10 times, expecting it to get 3 words each time, even
		// though our pool has only got a size of 3.
		for i := 0; i < 10; i++ {
			words := GetRandomWords(3, lobby)
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
