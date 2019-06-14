package main

import (
	"log"
	"net/http"

	_ "github.com/Bios-Marcel/scribble.rs/lobby"
)

func main() {
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
}
