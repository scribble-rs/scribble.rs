package game

import (
	"embed"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type languageData struct {
	languageCode string
	lowercaser   func() cases.Caser
}

var (
	ErrUnknownWordList = errors.New("wordlist unknown")
	wordlistData       = map[string]languageData{
		"english_gb": {
			languageCode: "en_gb",
			lowercaser:   func() cases.Caser { return cases.Lower(language.BritishEnglish) },
		},
		"english": {
			languageCode: "en_us",
			lowercaser:   func() cases.Caser { return cases.Lower(language.AmericanEnglish) },
		},
		"italian": {
			languageCode: "it",
			lowercaser:   func() cases.Caser { return cases.Lower(language.Italian) },
		},
		"german": {
			languageCode: "de",
			lowercaser:   func() cases.Caser { return cases.Lower(language.German) },
		},
		"french": {
			languageCode: "fr",
			lowercaser:   func() cases.Caser { return cases.Lower(language.French) },
		},
		"dutch": {
			languageCode: "nl",
			lowercaser:   func() cases.Caser { return cases.Lower(language.Dutch) },
		},
		"ukrainian": {
			languageCode: "ua",
			lowercaser:   func() cases.Caser { return cases.Lower(language.Ukrainian) },
		},
	}

	//go:embed words/*
	wordFS embed.FS
)

func getLanguageIdentifier(language string) string {
	return wordlistData[language].languageCode
}

// readWordListInternal exists for testing purposes, it allows passing a custom
// wordListSupplier, in order to avoid having to write tests aggainst the
// default language lists.
func readWordListInternal(
	lowercaser cases.Caser, chosenLanguage string,
	wordlistSupplier func(string) (string, error),
) ([]string, error) {
	languageIdentifier := getLanguageIdentifier(chosenLanguage)
	if languageIdentifier == "" {
		return nil, ErrUnknownWordList
	}

	wordListFile, err := wordlistSupplier(languageIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error invoking wordlistSupplier: %w", err)
	}

	// Wordlists are guaranteed not to contain any carriage returns (\r).
	words := strings.Split(lowercaser.String(wordListFile), "\n")
	shuffleWordList(words)
	return words, nil
}

// readDefaultWordList reads the wordlist for the given language from the filesystem.
// If found, the list is cached and will be read from the cache upon next
// request. The returned slice is a safe copy and can be mutated. If the
// specified has no corresponding wordlist, an error is returned. This has been
// a panic before, however, this could enable a user to forcefully crash the
// whole application.
func readDefaultWordList(lowercaser cases.Caser, chosenLanguage string) ([]string, error) {
	return readWordListInternal(lowercaser, chosenLanguage, func(key string) (string, error) {
		wordBytes, err := wordFS.ReadFile("words/" + key)
		if err != nil {
			return "", fmt.Errorf("error reading wordfile: %w", err)
		}

		return string(wordBytes), nil
	})
}

func reloadLobbyWords(lobby *Lobby) ([]string, error) {
	return readDefaultWordList(lobby.lowercaser, lobby.Wordpack)
}

// GetRandomWords gets a custom amount of random words for the passed Lobby.
// The words will be chosen from the custom words and the default
// dictionary, depending on the settings specified by the lobbies creator.
func GetRandomWords(wordCount int, lobby *Lobby) []string {
	return getRandomWords(wordCount, lobby, reloadLobbyWords)
}

// getRandomWords exists for test purposes, allowing to define a custom
// reloader, allowing us to specify custom wordlists in the tests without
// running into a panic on reload.
func getRandomWords(wordCount int, lobby *Lobby, reloadWords func(lobby *Lobby) ([]string, error)) []string {
	words := make([]string, wordCount)

	for customWordsLeft, i := lobby.CustomWordsPerTurn, 0; i < wordCount; i++ {
		if customWordsLeft > 0 && len(lobby.CustomWords) > 0 {
			customWordsLeft--
			words[i] = popCustomWord(lobby)
		} else {
			words[i] = popWordpackWord(lobby, reloadWords)
		}
	}

	return words
}

func popCustomWord(lobby *Lobby) string {
	lastIndex := len(lobby.CustomWords) - 1
	lastWord := lobby.CustomWords[lastIndex]
	lobby.CustomWords = lobby.CustomWords[:lastIndex]
	return lastWord
}

// popWordpackWord gets X words from the wordpack. The major difference to
// popCustomWords is, that the wordlist gets reset and reshuffeled once every
// item has been popped.
func popWordpackWord(lobby *Lobby, reloadWords func(lobby *Lobby) ([]string, error)) string {
	if len(lobby.words) == 0 {
		var err error
		lobby.words, err = reloadWords(lobby)
		if err != nil {
			// Since this list should've been successfully read once before, we
			// can "safely" panic if this happens, assuming that there's a
			// deeper problem.
			panic(err)
		}
	}
	lastIndex := len(lobby.words) - 1
	lastWord := lobby.words[lastIndex]
	lobby.words = lobby.words[:lastIndex]
	return lastWord
}

func shuffleWordList(wordlist []string) {
	rand.Shuffle(len(wordlist), func(a, b int) {
		wordlist[a], wordlist[b] = wordlist[b], wordlist[a]
	})
}
