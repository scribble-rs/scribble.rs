package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	errorPage *template.Template
	portHTTP  *int
)

func main() {
	portHTTP = flag.Int("portHTTP", 8081, "defines the port to be used for http mode")

	flag.Parse()

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	var parseError error
	errorPage, parseError = template.New("").ParseFiles("error.html", "footer.html")
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portHTTP), nil))
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
}
