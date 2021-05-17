package game

import (
	"embed"
	"io"
	"math/rand"
	"regexp"
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
	wordlistSupplier func(string) (string, error)) ([]string, error) {

	languageIdentifier := getLanguageIdentifier(chosenLanguage)
	words, available := wordListCache[languageIdentifier]
	if !available {
		wordListFile, supplierError := wordlistSupplier(languageIdentifier)
		if supplierError != nil {
			return nil, supplierError
		}

		//Due to people having git autoreplace newline characters, there
		//might be unnecessary \r characters.
		//While regex isn't super, this doesn't really matter as the word lists
		//are cached and only the first start of the first lobby will be slower.
		words = regexp.MustCompile("\r?\n").Split(wordListFile, -1)
		for wordIndex, word := range words {
			words[wordIndex] = lowercaser.String(word)
		}
		wordListCache[languageIdentifier] = words
	}

	//We don't shuffle the wordList directory, as the cache isn't threadsafe.
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
		wordFile, wordErr := wordFS.Open("words/" + key)
		if wordErr != nil {
			return "", wordErr
		}

		wordBytes, readError := io.ReadAll(wordFile)
		if readError != nil {
			return "", readError
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
	if lobby.CustomWordsChance > 0 && len(lobby.CustomWords) > 0 {
		//Always get custom words
		if lobby.CustomWordsChance == 100 {
			if len(lobby.CustomWords) >= wordCount {
				return popCustomWords(wordCount, lobby)
			}

			leftOverCustomWords := len(lobby.CustomWords)
			return append(
				popCustomWords(len(lobby.CustomWords), lobby),
				popWordpackWords(wordCount-leftOverCustomWords, lobby)...)
		}

		words := make([]string, 0, wordCount)
		for i := 0; i < wordCount; i++ {
			if len(lobby.CustomWords) >= 1 && rng() <= lobby.CustomWordsChance {
				words = append(words, popCustomWords(1, lobby)...)
			} else {
				words = append(words, popWordpackWords(1, lobby)...)
			}
		}

		return words
	}

	return popWordpackWords(wordCount, lobby)
}

func popCustomWords(wordCount int, lobby *Lobby) []string {
	wordIndex := len(lobby.CustomWords) - wordCount
	lastWords := lobby.CustomWords[wordIndex:]
	lobby.CustomWords = lobby.CustomWords[:wordIndex]
	return lastWords
}

// popWordpackWords gets X words from the wordpack. The major difference to
// popCustomWords is, that the wordlist gets reset and reshuffeled once every
// item has been popped.
func popWordpackWords(wordCount int, lobby *Lobby) []string {
	if len(lobby.words) < wordCount {
		var readError error
		lobby.words, readError = readWordList(lobby.lowercaser, lobby.Wordpack)
		if readError != nil {
			//Since this list should've been successfully read once before, we
			//can "safely" panic if this happens, assuming that there's a
			//deeper problem.
			panic(readError)
		}
	}
	wordIndex := len(lobby.words) - wordCount
	lastThreeWords := lobby.words[wordIndex:]
	lobby.words = lobby.words[:wordIndex]
	return lastThreeWords
}

func shuffleWordList(wordlist []string) {
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(wordlist), func(a, b int) {
		wordlist[a], wordlist[b] = wordlist[b], wordlist[a]
	})
}
