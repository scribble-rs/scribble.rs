package frontend

import (
	"embed"
	"fmt"
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

func init() {
	var err error
	pageTemplates, err = template.ParseFS(templateFS, "templates/*")
	if err != nil {
		panic(fmt.Errorf("error loading templates: %w", err))
	}
}

// BasePageConfig is data that all pages require to function correctly, no matter
// whether error page or lobby page.
type BasePageConfig struct {
	// Version is the source code version of this build.
	Version string `json:"version"`
	// RootPath is the path directly after the domain and before the
	// scribble.rs paths. For example if you host scribblers on painting.com
	// but already host a different website, then your API paths might have to
	// look like this: painting.com/scribblers/v1.
	RootPath string `json:"rootPath"`
	// CacheBust is a string that is appended to all resources to prevent
	// browsers from using cached data of a previous version, but still have
	// long lived max age values.
	CacheBust string `json:"cacheBust"`
}

// SetupRoutes registers the official webclient endpoints with the http package.
func (handler *SSRHandler) SetupRoutes(router chi.Router) {
	router.Get("/"+handler.cfg.RootPath, handler.homePageHandler)

	fileServer := http.FileServer(http.FS(frontendResourcesFS))
	router.Get(
		"/"+path.Join(handler.cfg.RootPath, "resources/*"),
		http.StripPrefix(
			"/"+handler.cfg.RootPath,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Duration of 1 year, since we use cachebusting anyway.
				w.Header().Set("Cache-Control", "public, max-age=31536000")

				fileServer.ServeHTTP(w, r)
			}),
		).ServeHTTP,
	)
	router.Get("/"+path.Join(handler.cfg.RootPath, "ssrEnterLobby/{lobby_id}"), handler.ssrEnterLobby)
	router.Post("/"+path.Join(handler.cfg.RootPath, "ssrCreateLobby"), handler.ssrCreateLobby)
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
func (handler *SSRHandler) userFacingError(w http.ResponseWriter, errorMessage string) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", &errorPageData{
		BasePageConfig: handler.basePageConfig,
		ErrorMessage:   errorMessage,
	})
	// This should never happen, but if it does, something is very wrong.
	if err != nil {
		panic(err)
	}
}
