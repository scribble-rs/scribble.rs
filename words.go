package main

import (
	"io/ioutil"
	"math/rand"
	"strings"
)

var englishWords []string

func init() {
	data, err := ioutil.ReadFile("resources/words_en")
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
			englishWords = append(englishWords, word)
		} else {
			englishWords = append(englishWords, word[:len(word)-2])
		}
	}
}

func GetRandomWords() []string {
	return []string{
		englishWords[rand.Int()%len(englishWords)],
		englishWords[rand.Int()%len(englishWords)],
		englishWords[rand.Int()%len(englishWords)],
	}

}
