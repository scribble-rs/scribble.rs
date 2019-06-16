package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
)

var (
	useHTTPS  *bool
	errorPage *template.Template
)

func main() {
	useHTTPS = flag.Bool("https", false, "sets https usage to true or false")

	var parseError error
	errorPage, parseError = template.New("").ParseFiles("error.html", "footer.html")
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()

	if *useHTTPS {
		//TODO Use Https
		log.Fatal(http.ListenAndServe(":443", nil))
	} else {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
}
