package api

import (
	"net/http"
	"os"
)

//In this init hook we initialize all templates that could at some point
//be needed during the server runtime. If any of the templates can't be
//loaded, we panic.
func init() {
	rootPath, rootPathAvailable := os.LookupEnv("ROOT_PATH")
	if rootPathAvailable && rootPath != "" {
		CurrentBasePageConfig.RootPath = rootPath
	}
}

// SetupRoutes registers the /v1/ endpoints with the http package.
func SetupRoutes() {
	http.HandleFunc(CurrentBasePageConfig.RootPath+"/v1/stats", stats)
	//The websocket is shared between the public API and the official client
	http.HandleFunc(CurrentBasePageConfig.RootPath+"/v1/ws", wsEndpoint)

	//These exist only for the public API. We version them in order to ensure
	//backwards compatibility as far as possible.
	http.HandleFunc(CurrentBasePageConfig.RootPath+"/v1/lobby", lobbyEndpoint)
	http.HandleFunc(CurrentBasePageConfig.RootPath+"/v1/lobby/player", enterLobby)
}
