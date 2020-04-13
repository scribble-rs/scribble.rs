package communication

import (
	"html/template"
	"net/http"

	"github.com/gobuffalo/packr/v2"
)

var (
	errorPage       *template.Template
	lobbyCreatePage *template.Template
	lobbyPage       *template.Template
)

func findStringFromBox(box *packr.Box, name string) string {
	result, err := box.FindString(name)
	//Since this isn't a runtime error that should happen, we instantly panic.
	if err != nil {
		panic(errorPage)
	}

	return result
}

//In this init hook we initialize all templates that could at some point be
//needed during the server runtime. If any of the templates can't be loaded, we
//panic.
func init() {
	templates := packr.New("templates", "../templates")
	var parseError error
	errorPage, parseError = template.New("error.html").Parse(findStringFromBox(templates, "error.html"))
	if parseError != nil {
		panic(parseError)
	}
	errorPage, parseError = errorPage.New("footer.html").Parse(findStringFromBox(templates, "footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	lobbyCreatePage, parseError = template.New("lobby_create.html").Parse(findStringFromBox(templates, "lobby_create.html"))
	if parseError != nil {
		panic(parseError)
	}
	lobbyCreatePage, parseError = lobbyCreatePage.New("footer.html").Parse(findStringFromBox(templates, "footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	lobbyPage, parseError = template.New("lobby.html").Parse(findStringFromBox(templates, "lobby.html"))
	if parseError != nil {
		panic(parseError)
	}
	lobbyPage, parseError = lobbyPage.New("footer.html").Parse(findStringFromBox(templates, "footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()
}

func setupRoutes() {
	frontendRessourcesBox := packr.New("frontend", "../resources/frontend")
	//Endpoints for official webclient
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(frontendRessourcesBox)))
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
