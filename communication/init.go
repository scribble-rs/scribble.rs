package communication

import (
	"embed"
	"html/template"
	"net/http"
)

var (
	//go:embed templates/*
	templateFS    embed.FS
	pageTemplates *template.Template

	//go:embed resources/*
	frontendResourcesFS embed.FS
)

//In this init hook we initialize all templates that could at some point
//be needed during the server runtime. If any of the templates can't be
//loaded, we panic.
func init() {
	var templateParseError error
	pageTemplates, templateParseError = template.ParseFS(templateFS, "templates/*")
	if templateParseError != nil {
		panic(templateParseError)
	}

	setupRoutes()
}

func setupRoutes() {
	//Endpoints for official webclient
	http.Handle("/resources/", http.FileServer(http.FS(frontendResourcesFS)))
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
