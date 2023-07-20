package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func main() {
	languageFile, err := os.Open(os.Args[len(os.Args)-1])
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(languageFile)
	if err != nil {
		panic(err)
	}
	var words map[string]json.RawMessage
	err = json.Unmarshal(data, &words)
	if err != nil {
		panic(err)
	}

	for word := range words {
		fmt.Println(word)
	}
}
