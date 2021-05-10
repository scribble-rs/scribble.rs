package frontend

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/scribble-rs/scribble.rs/api"
	"github.com/scribble-rs/scribble.rs/translations"
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
}

var currentBasePageConfig = &BasePageConfig{
	RootPath: api.RootPath,
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
func SetupRoutes() {
	http.Handle(api.RootPath+"/resources/",
		http.StripPrefix(api.RootPath,
			http.FileServer(http.FS(frontendResourcesFS))))
	http.HandleFunc(api.RootPath+"/", homePage)
	http.HandleFunc(api.RootPath+"/ssrEnterLobby", ssrEnterLobby)
	http.HandleFunc(api.RootPath+"/ssrCreateLobby", ssrCreateLobby)
}

// errorPageData represents the data that error.html requires to be displayed.
type errorPageData struct {
	*BasePageConfig
	// ErrorMessage displayed on the page.
	ErrorMessage string

	Translation translations.Translation
	Locale      string
}

//userFacingError will return the occurred error as a custom html page to the caller.
func userFacingError(w http.ResponseWriter, errorMessage string) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", &errorPageData{
		BasePageConfig: currentBasePageConfig,
		ErrorMessage:   errorMessage,
	})
	//This should never happen, but if it does, something is very wrong.
	if err != nil {
		panic(err)
	}
}
