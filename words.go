package main

import (
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

const wordPath string = "resources/words/"

var wordList []string

func readWordList(chosenLanguage string) {
	languageMap := map[string]string{
		"English": "words_en",
		"Italian": "words_it",
		"German":  "words_de",
	}
	data, err := ioutil.ReadFile(wordPath + languageMap[chosenLanguage])
	if err != nil {
		panic(err)
	}

	tempWords := strings.Split(string(data), "\n")
	for _, word := range tempWords {
		word = strings.TrimSpace(word)
		if strings.HasSuffix(word, "#i") {
			continue
		}

		lastIndexNumberSign := strings.LastIndex(word, "#")
		if lastIndexNumberSign == -1 {
			wordList = append(wordList, word)
		} else {
			wordList = append(wordList, word[:len(word)-2])
		}
	}
}

// GetRandomWords gets 3 random words for the passed Lobby. The words will be
// chosen from the custom words and the default dictionary, depending on the
// settings specified by ther Lobby-Owner.
func GetRandomWords(lobby *Lobby) []string {
	rand.Seed(time.Now().Unix())
	wordsNotToPick := lobby.alreadyUsedWords
	wordOne := getRandomWordWithCustomWordChance(wordsNotToPick, lobby.CustomWords, lobby.CustomWordsChance)
	wordsNotToPick = append(wordsNotToPick, wordOne)
	wordTwo := getRandomWordWithCustomWordChance(wordsNotToPick, lobby.CustomWords, lobby.CustomWordsChance)
	wordsNotToPick = append(wordsNotToPick, wordTwo)
	wordThree := getRandomWordWithCustomWordChance(wordsNotToPick, lobby.CustomWords, lobby.CustomWordsChance)

	return []string{
		wordOne,
		wordTwo,
		wordThree,
	}
}

func getRandomWordWithCustomWordChance(wordsAlreadyUsed []string, customWords []string, customWordChance int) string {
	if rand.Intn(100)+1 <= customWordChance {
		return getUnusedCustomWord(wordsAlreadyUsed, customWords)
	}

	return getUnusedRandomWord(wordsAlreadyUsed)
}

func getUnusedCustomWord(wordsAlreadyUsed []string, customWords []string) string {
OUTER_LOOP:
	for _, word := range customWords {
		for _, usedWord := range wordsAlreadyUsed {
			if usedWord == word {
				continue OUTER_LOOP
			}
		}

		return word
	}

	return getUnusedRandomWord(wordsAlreadyUsed)
}

func getUnusedRandomWord(wordsAlreadyUsed []string) string {
	//We attempt to find a random word for a hundred times, afterwards we just use any.
	randomnessAttempts := 0
	var word string
OUTER_LOOP:
	for {
		word = wordList[rand.Int()%len(wordList)]
		for _, usedWord := range wordsAlreadyUsed {
			if usedWord == word {
				if randomnessAttempts == 100 {
					break OUTER_LOOP
				}

				randomnessAttempts++
				continue OUTER_LOOP
			}
		}
		break
	}

	return word
}
