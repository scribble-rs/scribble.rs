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

	englishWords = strings.Split(string(data), "\n")
}

func GetRandomWords() []string {
	return []string{
		englishWords[rand.Int()%len(englishWords)],
		englishWords[rand.Int()%len(englishWords)],
		englishWords[rand.Int()%len(englishWords)],
	}

}
