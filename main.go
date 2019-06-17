package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var (
	errorPage *template.Template
)

func main() {
	portHTTP := flag.Int("portHTTP", 8080, "defines the port to be used for http mode")

	flag.Parse()

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
