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
	lobbyPage, parseError = lobbyPage.New("footer.html").Parse(readTemplateFile("footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()
}

func setupRoutes() {
	//Endpoints for official webclient
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(pkger.Dir("/resources"))))
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ssrEnterLobby", ssrEnterLobby)
	http.HandleFunc("/ssrCreateLobby", ssrCreateLobby)

	//The websocket is shared between the public API and the official client
	http.HandleFunc("/v1/ws", wsEndpoint)

	//These exist only for the public API. We version them in order to ensure
	//backwards compatibility as far as possible.
	http.HandleFunc("/v1/lobby", createLobby)
	http.HandleFunc("/v1/lobby/player", enterLobby)
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
