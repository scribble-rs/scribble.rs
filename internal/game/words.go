package game

import (
	"embed"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/text/cases"
)

var (
	wordListCache       = make(map[string][]string)
	languageIdentifiers = map[string]string{
		"english_gb": "en_gb",
		"english":    "en_us",
		"italian":    "it",
		"german":     "de",
		"french":     "fr",
		"dutch":      "nl",
	}

	//go:embed words/*
	wordFS embed.FS
)

func getLanguageIdentifier(language string) string {
	return languageIdentifiers[language]
}

// readWordListInternal exists for testing purposes, it allows passing a custom
// wordListSupplier, in order to avoid having to write tests aggainst the
// default language lists.
func readWordListInternal(
	lowercaser cases.Caser, chosenLanguage string,
	wordlistSupplier func(string) (string, error),
) ([]string, error) {
	languageIdentifier := getLanguageIdentifier(chosenLanguage)
	words, available := wordListCache[languageIdentifier]
	if !available {
		wordListFile, err := wordlistSupplier(languageIdentifier)
		if err != nil {
			return nil, fmt.Errorf("error invoking wordlistSupplier: %w", err)
		}

		// Wordlists are guaranteed not to contain any carriage returns (\r).
		words = strings.Split(lowercaser.String(wordListFile), "\n")
		wordListCache[languageIdentifier] = words
	}

	// We don't shuffle the wordList directory, as the cache isn't threadsafe.
	shuffledWords := make([]string, len(words))
	copy(shuffledWords, words)
	shuffleWordList(shuffledWords)
	return shuffledWords, nil
}

// readWordList reads the wordlist for the given language from the filesystem.
// If found, the list is cached and will be read from the cache upon next
// request. The returned slice is a safe copy and can be mutated. If the
// specified has no corresponding wordlist, an error is returned. This has been
// a panic before, however, this could enable a user to forcefully crash the
// whole application.
func readWordList(lowercaser cases.Caser, chosenLanguage string) ([]string, error) {
	return readWordListInternal(lowercaser, chosenLanguage, func(key string) (string, error) {
		wordBytes, err := wordFS.ReadFile("words/" + key)
		if err != nil {
			return "", fmt.Errorf("error reading wordfile: %w", err)
		}

		return string(wordBytes), nil
	})
}

// GetRandomWords gets a custom amount of random words for the passed Lobby.
// The words will be chosen from the custom words and the default
// dictionary, depending on the settings specified by the lobbies creator.
func GetRandomWords(wordCount int, lobby *Lobby) []string {
	return getRandomWordsCustomRng(wordCount, lobby, func() int { return rand.Intn(100) + 1 })
}

// getRandomWordsCustomRng allows passing a custom generator for random
// numbers. This can be used for predictability in unit tests.
// See GetRandomWords for functionality documentation.
func getRandomWordsCustomRng(wordCount int, lobby *Lobby, rng func() int) []string {
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		if lobby.CustomWordsChance > 0 && len(lobby.CustomWords) > 0 && rng() <= lobby.CustomWordsChance {
			words[i] = popCustomWord(lobby)
		} else {
			words[i] = popWordpackWord(lobby)
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
func popWordpackWord(lobby *Lobby) string {
	if len(lobby.words) == 0 {
		var err error
		lobby.words, err = readWordList(lobby.lowercaser, lobby.Wordpack)
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
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(wordlist), func(a, b int) {
		wordlist[a], wordlist[b] = wordlist[b], wordlist[a]
	})
}
