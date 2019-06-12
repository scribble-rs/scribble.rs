package main

import (
	"html/template"
	"log"
	"net/http"
)

var index *template.Template

func main() {
	var err error
	index, err = template.New("everything").ParseFiles("lobby.html", "footer.html")
	if err != nil {
		panic(err)
	}

	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.HandleFunc("/", homePage)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	err := index.ExecuteTemplate(w, "lobby.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
