package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	var err error
	lobbyCreatePage, err = template.New("everything").ParseFiles("lobby.html", "footer.html")
	if err != nil {
		panic(err)
	}

	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.HandleFunc("/", homePage)
	http.HandleFunc("/lobby/create", createLobby)
}
