package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	languageFile, err := os.Open(os.Args[len(os.Args)-2])
	if err != nil {
		panic(err)
	}

	lowercaser := cases.Lower(language.Make(os.Args[len(os.Args)-1]))
	reader := bufio.NewReader(languageFile)
	var words []string
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		lineAsString := string(line)

		//Remove previously edited difficulty indicators, we don't really need it anymore.
		difficultyIndicatorIndex := strings.IndexRune(lineAsString, '#')
		if difficultyIndicatorIndex != -1 {
			lineAsString = lineAsString[:difficultyIndicatorIndex]
		}

		//Lowercase and trim, to make sure we can compare them without errors
		words = append(words, strings.TrimSpace(lowercaser.String(lineAsString)))
	}

	var filteredWords []string
WORDS:
	for _, word := range words {
		for _, filteredWord := range filteredWords {
			if filteredWord == word {
				continue WORDS
			}
		}

		filteredWords = append(filteredWords, word)
	}

	//Filter for niceness
	sort.Strings(filteredWords)

	for _, word := range filteredWords {
		fmt.Println(word)
	}
}
