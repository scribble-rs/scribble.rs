package game

import (
	"embed"
	"io"
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

// readWordListInternal exists for testing purposes.
func readWordListInternal(
	lowercaser cases.Caser, chosenLanguage string,
	wordlistSupplier func(string) (string, error)) ([]string, error) {

	languageIdentifier := getLanguageIdentifier(chosenLanguage)
	list, available := wordListCache[languageIdentifier]
	if available {
		copiedList := make([]string, len(list))
		copy(copiedList, list)
		shuffleWordList(copiedList)
		return copiedList, nil
	}

	wordListFile, pkgerError := wordlistSupplier(languageIdentifier)
	if pkgerError != nil {
		return nil, pkgerError
	}

	tempWords := strings.Split(wordListFile, "\n")
	var words []string
	for _, word := range tempWords {
		word = strings.TrimSpace(word)

		//Newlines will just be empty strings
		if word == "" {
			continue
		}

		//Since not all words use the tag system, we can just instantly return for words that don't use it.
		lastIndexNumberSign := strings.LastIndex(word, "#")
		if lastIndexNumberSign == -1 {
			words = append(words, lowercaser.String(word))
		} else {
			//The "i" is the "impossible" tag, meaning the word was rated as undrawable / unguessable.
			if "#i" == word[lastIndexNumberSign:] {
				continue
			}
			words = append(words, lowercaser.String(word[:lastIndexNumberSign]))
		}
	}

	wordListCache[languageIdentifier] = words

	copiedList := make([]string, len(words))
	copy(copiedList, words)
	shuffleWordList(copiedList)
	return copiedList, nil
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
		for i := 0; i <= wordCount; i++ {
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
		shuffleWordList(lobby.words)
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
