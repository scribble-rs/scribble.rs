package game

import (
	"fmt"
	"strings"
	"testing"
)

func Test_readWordList(t *testing.T) {
	t.Run("test invalid language file", func(t *testing.T) {
		//We expect a panic for invalid language files.
		defer func() {
			err := recover()
			if err == nil {
				panic(fmt.Sprintf("Test should've failed, but returned nil error"))
			}
		}()
		_, readError := readWordList("owO")
		if readError == nil {
			t.Errorf("Reading word list didn't return an error, even though the langauge doesn't exist.")
		}
	})

	for language := range languageMap {
		t.Run(fmt.Sprintf("Testing language file for %s", language), func(t *testing.T) {
			//First run from box/drive
			testWordList(language, t)
			//Second run from in-memory cache
			testWordList(language, t)
		})
	}
}

func testWordList(language string, t *testing.T) {
	words, readError := readWordList(language)
	if readError != nil {
		t.Errorf("Error reading language %s: %s", language, readError)
	}

	if len(words) == 0 {
		t.Errorf("Wordlist for language %s was empty.", language)
	}

	for _, word := range words {
		if word == "" {
			t.Errorf("Wordlist for language %s contained empty word", language)
		}

		if strings.HasPrefix(word, " ") || strings.HasSuffix(word, " ") {
			t.Errorf("Word has surrounding spaces: %s", word)
		}
	}
}

func Test_getRandomWords(t *testing.T) {
	t.Run("Test getRandomWords with 3 words in list", func(t *testing.T) {
		lobby := &Lobby{
			CurrentWord: "",
			Words:       []string{"a", "b", "c"},
		}

		randomWords := GetRandomWords(lobby)
		for _, lobbyWord := range lobby.Words {
			if !arrayContains(randomWords, lobbyWord) {
				t.Errorf("Random words %s, didn't contain lobbyWord %s", randomWords, lobbyWord)
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 3 more in custom word list, but with 0 chance", func(t *testing.T) {
		lobby := &Lobby{
			CurrentWord:       "",
			Words:             []string{"a", "b", "c"},
			CustomWordsChance: 0,
			CustomWords:       []string{"d", "e", "f"},
		}

		randomWords := GetRandomWords(lobby)
		for _, lobbyWord := range lobby.Words {
			if !arrayContains(randomWords, lobbyWord) {
				t.Errorf("Random words %s, didn't contain lobbyWord %s", randomWords, lobbyWord)
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 100% custom word chance, but without custom words", func(t *testing.T) {
		lobby := &Lobby{
			CurrentWord:       "",
			Words:             []string{"a", "b", "c"},
			CustomWordsChance: 100,
			CustomWords:       nil,
		}

		randomWords := GetRandomWords(lobby)
		for _, lobbyWord := range lobby.Words {
			if !arrayContains(randomWords, lobbyWord) {
				t.Errorf("Random words %s, didn't contain lobbyWord %s", randomWords, lobbyWord)
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 100% custom word chance, with 3 custom words", func(t *testing.T) {
		lobby := &Lobby{
			CurrentWord:       "",
			Words:             []string{"a", "b", "c"},
			CustomWordsChance: 100,
			CustomWords:       []string{"d", "e", "f"},
		}

		for i := 0; i < 1000; i++ {
			randomWords := GetRandomWords(lobby)
			for _, customWord := range lobby.CustomWords {
				if !arrayContains(randomWords, customWord) {
					t.Errorf("Random words %s, didn't contain customWord %s", randomWords, customWord)
				}
			}
		}
	})

	t.Run("Test getRandomWords with 3 words in list and 100% custom word chance, with 3 custom words and one of them on the used list", func(t *testing.T) {
		lobby := &Lobby{
			CurrentWord:       "",
			Words:             []string{"a", "b", "c"},
			CustomWordsChance: 100,
			CustomWords:       []string{"d", "e", "f"},
			alreadyUsedWords:  []string{"f"},
		}

		for i := 0; i < 1000; i++ {
			randomWords := GetRandomWords(lobby)
			if !arrayContains(randomWords, "d") {
				t.Errorf("Random words %s, didn't contain customWord d", randomWords)
			}

			if !arrayContains(randomWords, "e") {
				t.Errorf("Random words %s, didn't contain customWord e", randomWords)
			}

			if arrayContains(randomWords, "f") {
				t.Errorf("Random words %s, contained customWord f", randomWords)
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
