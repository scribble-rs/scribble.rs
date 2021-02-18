package communication

import (
	"embed"
	"html/template"
	"net/http"
)

var (
	errorPage       *template.Template
	lobbyCreatePage *template.Template
	lobbyPage       *template.Template
)

//go:embed templates/*
var templateFS embed.FS

//In this init hook we initialize all templates that could at some point be
//needed during the server runtime. If any of the templates can't be loaded, we
//panic.
func init() {
	var parseError error

	errorPage, parseError = template.ParseFS(templateFS, "templates/error.html")
	if parseError != nil {
		panic(parseError)
	}
	errorPage, parseError = errorPage.New("footer.html").ParseFS(templateFS, "templates/footer.html")
	if parseError != nil {
		panic(parseError)
	}

	lobbyCreatePage, parseError = template.ParseFS(templateFS, "templates/lobby_create.html")
	if parseError != nil {
		panic(parseError)
	}
	lobbyCreatePage, parseError = lobbyCreatePage.New("footer.html").ParseFS(templateFS, "templates/footer.html")
	if parseError != nil {
		panic(parseError)
	}

	lobbyPage, parseError = template.ParseFS(templateFS, "templates/lobby.html")
	if parseError != nil {
		panic(parseError)
	}
	lobbyPage, parseError = lobbyPage.New("footer.html").ParseFS(templateFS, "templates/footer.html")
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()
}

//go:embed resources/*
var frontendRessources embed.FS

func setupRoutes() {
	//Endpoints for official webclient
	http.Handle("/resources/", http.FileServer(http.FS(frontendRessources)))
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ssrEnterLobby", ssrEnterLobby)
	http.HandleFunc("/ssrCreateLobby", ssrCreateLobby)

	//The websocket is shared between the public API and the official client
	http.HandleFunc("/v1/ws", wsEndpoint)

	//These exist only for the public API. We version them in order to ensure
	//backwards compatibility as far as possible.
	http.HandleFunc("/v1/lobby", lobbyEndpoint)
	http.HandleFunc("/v1/lobby/player", enterLobby)
}
