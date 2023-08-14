package frontend

import (
	"embed"
	"html/template"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/scribble-rs/scribble.rs/internal/translations"
)

var (
	//go:embed templates/*
	templateFS    embed.FS
	pageTemplates *template.Template

	//go:embed resources/*
	frontendResourcesFS embed.FS
)

// In this init hook we initialize all templates that could at some point
// be needed during the server runtime. If any of the templates can't be
// loaded, we panic.
func init() {
	var templateParseError error
	pageTemplates, templateParseError = template.ParseFS(templateFS, "templates/*")
	if templateParseError != nil {
		panic(templateParseError)
	}
}

// FIXME Delete global state.
var currentBasePageConfig = &BasePageConfig{}

func SetRootPath(rootPath string) {
	if rootPath != "" {
		currentBasePageConfig.RootPath = "/" + rootPath
	}
}

// BasePageConfig is data that all pages require to function correctly, no matter
// whether error page or lobby page.
type BasePageConfig struct {
	// RootPath is the path directly after the domain and before the
	// scribble.rs paths. For example if you host scribblers on painting.com
	// but already host a different website, then your API paths might have to
	// look like this: painting.com/scribblers/v1.
	RootPath string `json:"rootPath"`
}

// SetupRoutes registers the official webclient endpoints with the http package.
func SetupRoutes(rootPath string, router chi.Router) {
	router.Get("/"+rootPath, homePage)
	router.Get(
		"/"+path.Join(rootPath, "resources/*"),
		http.StripPrefix(
			"/"+rootPath,
			http.FileServer(http.FS(frontendResourcesFS)),
		).ServeHTTP,
	)
	router.Get("/"+path.Join(rootPath, "ssrEnterLobby/{lobby_id}"), ssrEnterLobby)
	router.Post("/"+path.Join(rootPath, "ssrCreateLobby"), ssrCreateLobby)
}

// errorPageData represents the data that error.html requires to be displayed.
type errorPageData struct {
	*BasePageConfig
	// ErrorMessage displayed on the page.
	ErrorMessage string

	Translation translations.Translation
	Locale      string
}

// userFacingError will return the occurred error as a custom html page to the caller.
func userFacingError(w http.ResponseWriter, errorMessage string) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", &errorPageData{
		BasePageConfig: currentBasePageConfig,
		ErrorMessage:   errorMessage,
	})
	// This should never happen, but if it does, something is very wrong.
	if err != nil {
		panic(err)
	}
}
