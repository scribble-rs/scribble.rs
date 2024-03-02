package game

import (
	"embed"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type LanguageData struct {
	LanguageCode string
	Lowercaser   func() cases.Caser
}

var (
	ErrUnknownWordList = errors.New("wordlist unknown")
	WordlistData       = map[string]LanguageData{
		"english_gb": {
			LanguageCode: "en_gb",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.BritishEnglish) },
		},
		"english": {
			LanguageCode: "en_us",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.AmericanEnglish) },
		},
		"italian": {
			LanguageCode: "it",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.Italian) },
		},
		"german": {
			LanguageCode: "de",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.German) },
		},
		"french": {
			LanguageCode: "fr",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.French) },
		},
		"dutch": {
			LanguageCode: "nl",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.Dutch) },
		},
		"ukrainian": {
			LanguageCode: "ua",
			Lowercaser:   func() cases.Caser { return cases.Lower(language.Ukrainian) },
		},
	}

	//go:embed words/*
	wordFS embed.FS
)

func getLanguageIdentifier(language string) string {
	return WordlistData[language].LanguageCode
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

const (
	EqualGuess   = 0
	CloseGuess   = 1
	DistantGuess = 2
)

// CheckGuess compares the strings with eachother. Possible results:
//   - EqualGuess (0)
//   - CloseGuess (1)
//   - DistantGuess (2)
//
// This works mostly like levensthein distance, but doesn't check further than
// to a distance of 2 and also handles transpositions where the runes are
// directly next to eachother.
func CheckGuess(a, b string) int {
	// Simplify logic later on; FIXME Explain
	if len(a) < len(b) {
		a, b = b, a
	} else if a == b {
		return EqualGuess
	}

	// We only want to indicate a close guess if:
	//   * 1 additional character is found (abc ~ abcd)
	//   * 1 character is missing (abc ~ ab)
	//   * 1 character is wrong (abc ~ adc)
	//   * 2 characters are swapped (abc ~ acb)

	if len(a)-len(b) > CloseGuess {
		return DistantGuess
	}

	var distance int
	aBytes := []byte(a)
	bBytes := []byte(b)
	for {
		aRune, aSize := utf8.DecodeRune(aBytes)
		// If a eaches the end, then so does b, as we make sure a is longer at
		// the top, therefore we can be sure no additional conflict diff occurs.
		if aRune == utf8.RuneError {
			return distance
		}
		bRune, bSize := utf8.DecodeRune(bBytes)

		// Either different runes, or b is empty, returning RuneError (65533).
		if aRune != bRune {
			// Check for transposition (abc ~ acb)
			nextARune, nextASize := utf8.DecodeRune(aBytes[aSize:])
			if nextARune == bRune {
				if nextARune != utf8.RuneError {
					nextBRune, nextBSize := utf8.DecodeRune(bBytes[bSize:])
					if nextBRune == aRune {
						distance++
						aBytes = aBytes[aSize+nextASize:]
						bBytes = bBytes[bSize+nextBSize:]
						continue
					}
				}

				// Make sure to not pop from b, so we can compare the rest, in
				// case we are only missing one character for cases such as:
				//   abc ~ bc
				//   abcde ~ abde
				bSize = 0
			} else {
				// We'd reach a diff of 2 now. Needs to happen after transposition
				// though, as transposition could still prove us wrong.
				if distance == 1 {
					return DistantGuess
				}
			}

			distance++
		}

		aBytes = aBytes[aSize:]
		bBytes = bBytes[bSize:]
	}
}
