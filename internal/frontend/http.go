package frontend

import (
	"crypto/md5"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/translations"
)

var (
	//go:embed templates/*
	templateFS    embed.FS
	pageTemplates *template.Template

	//go:embed resources/*
	frontendResourcesFS embed.FS
)

func Init() error {
	var err error
	pageTemplates, err = template.ParseFS(templateFS, "templates/*")
	if err != nil {
		return fmt.Errorf("error loading templates: %w", err)
	}

	entries, err := frontendResourcesFS.ReadDir("resources")
	if err != nil {
		return fmt.Errorf("error reading resource directory: %w", err)
	}

	hash := md5.New()
	for _, entry := range entries {
		bytes, err := frontendResourcesFS.ReadFile("resources/" + entry.Name())
		if err != nil {
			return fmt.Errorf("error reading resource %s: %w", entry.Name(), err)
		}

		if _, err := hash.Write(bytes); err != nil {
			return fmt.Errorf("error hashing resource %s: %w", entry.Name(), err)
		}
	}

	currentBasePageConfig.CacheBust = fmt.Sprintf("%x", hash.Sum(nil))
	return nil
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
	// CacheBust is a string that is appended to all resources to prevent
	// browsers from using cached data of a previous version, but still have
	// long lived max age values.
	CacheBust string `json:"cacheBust"`
}

// SetupRoutes registers the official webclient endpoints with the http package.
func SetupRoutes(config *config.Config, router chi.Router) {
	router.Get("/"+config.RootPath, newHomePageHandler(config.LobbySettingDefaults))

	fileServer := http.FileServer(http.FS(frontendResourcesFS))
	router.Get(
		"/"+path.Join(config.RootPath, "resources/*"),
		http.StripPrefix(
			"/"+config.RootPath,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Duration of 1 year, since we use cachebusting anyway.
				w.Header().Set("Cache-Control", "public, max-age=31536000")

				fileServer.ServeHTTP(w, r)
			}),
		).ServeHTTP,
	)
	router.Get("/"+path.Join(config.RootPath, "ssrEnterLobby/{lobby_id}"), ssrEnterLobby)
	router.Post("/"+path.Join(config.RootPath, "ssrCreateLobby"), ssrCreateLobby)
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
