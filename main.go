package main

import (
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	errorPage *template.Template
	portHTTP  *int
)

func main() {
	portHTTP = flag.Int("portHTTP", 8080, "defines the port to be used for http mode")

	flag.Parse()

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	var parseError error
	errorPage, parseError = template.New("").ParseFiles("error.html", "footer.html")
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()

	// "When you deploy an app through heroku, it does not allow you to specify the port number"
	// https://stackoverflow.com/a/51344239/3927431
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
}
