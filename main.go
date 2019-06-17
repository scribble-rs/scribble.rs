package main

import (
	"flag"
	"fmt"
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
	portHTTP := flag.Int("portHTTP", 80, "defines the port to be used for http mode")
	portHTTPS := flag.Int("portHTTPS", 443, "defines the port to be used for https mode")

	flag.Parse()

	var parseError error
	errorPage, parseError = template.New("").ParseFiles("error.html", "footer.html")
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()

	if *useHTTPS {
		//TODO Use Https
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portHTTPS), nil))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portHTTP), nil))
	}
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
}
