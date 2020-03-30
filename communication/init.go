package communication

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/markbates/pkger"
)

var (
	errorPage       *template.Template
	lobbyCreatePage *template.Template
	lobbyPage       *template.Template
)

//In this init hook we initialize all templates that could at some point be
//needed during the server runtime. If any of the templates can't be loaded, we
//panic.
func init() {
	var parseError error
	errorPage, parseError = template.New("error.html").Parse(readTemplateFile("error.html"))
	if parseError != nil {
		panic(parseError)
	}
	errorPage, parseError = errorPage.New("footer.html").Parse(readTemplateFile("footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	lobbyCreatePage, parseError = template.New("lobby_create.html").Parse(readTemplateFile("lobby_create.html"))
	if parseError != nil {
		panic(parseError)
	}
	lobbyCreatePage, parseError = lobbyCreatePage.New("footer.html").Parse(readTemplateFile("footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	lobbyPage, parseError = template.New("lobby.html").Parse(readTemplateFile("lobby.html"))
	if parseError != nil {
		panic(parseError)
	}
	lobbyPage, parseError = lobbyPage.New("lobby_players.html").Parse(readTemplateFile("lobby_players.html"))
	if parseError != nil {
		panic(parseError)
	}
	lobbyPage, parseError = lobbyPage.New("lobby_word.html").Parse(readTemplateFile("lobby_word.html"))
	if parseError != nil {
		panic(parseError)
	}
	lobbyPage, parseError = lobbyPage.New("footer.html").Parse(readTemplateFile("footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(pkger.Dir("/resources"))))

	http.HandleFunc("/", homePage)
	http.HandleFunc("/lobby", ShowLobby)
	http.HandleFunc("/lobby/create", createLobby)
	http.HandleFunc("/lobby/players", GetPlayers)
	http.HandleFunc("/lobby/wordhint", GetWordHint)
	http.HandleFunc("/lobby/rounds", GetRounds)
	http.HandleFunc("/ws", wsEndpoint)
}

func readTemplateFile(name string) string {
	templateHandle, pkgerError := pkger.Open("/templates/" + name)
	if pkgerError != nil {
		panic(pkgerError)
	}
	defer templateHandle.Close()

	bytes, readError := ioutil.ReadAll(templateHandle)
	if readError != nil {
		panic(readError)
	}

	return string(bytes)
}
