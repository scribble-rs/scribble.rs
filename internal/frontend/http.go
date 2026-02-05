package frontend

import (
	"embed"
	"encoding/hex"
	"fmt"
	"hash"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gofrs/uuid/v5"
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
	checksums map[string]string
	hash      hash.Hash

	// Version is the tagged source code version of this build. Can be empty for dev
	// builds. Untagged commits will be of format `tag-N-gSHA`.
	Version string `json:"version"`
	// Commit that was deployed, if we didn't deploy a concrete tag.
	Commit string `json:"commit"`
	// RootPath is the path directly after the domain and before the
	// scribble.rs paths. For example if you host scribblers on painting.com
	// but already host a different website, then your API paths might have to
	// look like this: painting.com/scribblers/v1.
	RootPath string `json:"rootPath"`
	// RootURL is similar to RootPath, but contains only the protocol and
	// domain. So it could be https://painting.com. This is required for some
	// non critical functionality, such as metadata tags.
	RootURL string `json:"rootUrl"`
	// CanonicalURL specifies the original domain, in case we are accessing the
	// site via some other domain, such as scribblers.fly.dev
	CanonicalURL string `json:"canonicalUrl"`
	// AllowIndexing will control whether the noindex, nofollow meta tag is
	// added to the home page.
	AllowIndexing bool `env:"ALLOW_INDEXING"`
}

var fallbackChecksum = uuid.Must(uuid.NewV4()).String()

func (baseConfig *BasePageConfig) Hash(key string, bytes []byte) error {
	_, alreadyExists := baseConfig.checksums[key]
	if alreadyExists {
		return fmt.Errorf("duplicate hash key '%s'", key)
	}
	if _, err := baseConfig.hash.Write(bytes); err != nil {
		return fmt.Errorf("error hashing '%s': %w", key, err)
	}
	baseConfig.checksums[key] = hex.EncodeToString(baseConfig.hash.Sum(nil))
	baseConfig.hash.Reset()
	return nil
}

// CacheBust is a string that is appended to all resources to prevent
// browsers from using cached data of a previous version, but still have
// long lived max age values.
func (baseConfig *BasePageConfig) withCacheBust(file string) string {
	checksum, found := baseConfig.checksums[file]
	if !found {
		// No need to crash over
		return fmt.Sprintf("%s?cache_bust=%s", file, fallbackChecksum)
	}
	return fmt.Sprintf("%s?cache_bust=%s", file, checksum)
}

func (baseConfig *BasePageConfig) WithCacheBust(file string) template.HTMLAttr {
	return template.HTMLAttr(baseConfig.withCacheBust(file))
}

func (handler *SSRHandler) cspMiddleware(handleFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Security-Policy", "base-uri 'self'; default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; img-src 'self' data:")
		handleFunc.ServeHTTP(w, r)
	}
}

// SetupRoutes registers the official webclient endpoints with the http package.
func (handler *SSRHandler) SetupRoutes(register func(string, string, http.HandlerFunc)) {
	registerWithCsp := func(s1, s2 string, hf http.HandlerFunc) {
		register(s1, s2, handler.cspMiddleware(hf))
	}
	var genericFileHandler http.HandlerFunc
	if dir := handler.cfg.ServeDirectories[""]; dir != "" {
		delete(handler.cfg.ServeDirectories, "")
		fileServer := http.FileServer(http.FS(os.DirFS(dir)))
		genericFileHandler = fileServer.ServeHTTP
	}

	for route, dir := range handler.cfg.ServeDirectories {
		fileServer := http.FileServer(http.FS(os.DirFS(dir)))
		fileHandler := http.StripPrefix(
			"/"+path.Join(handler.cfg.RootPath, route)+"/",
			http.HandlerFunc(fileServer.ServeHTTP),
		).ServeHTTP
		registerWithCsp(
			// Trailing slash means wildcard.
			"GET", path.Join(handler.cfg.RootPath, route)+"/",
			fileHandler,
		)
	}

	indexHandler := handler.cspMiddleware(handler.indexPageHandler)
	register("GET", handler.cfg.RootPath,
		http.StripPrefix(
			"/"+handler.cfg.RootPath,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "" || r.URL.Path == "/" {
					indexHandler(w, r)
					return
				}

				if genericFileHandler != nil {
					genericFileHandler.ServeHTTP(w, r)
				}
			})).ServeHTTP)

	fileServer := http.FileServer(http.FS(frontendResourcesFS))
	registerWithCsp(
		"GET", path.Join(handler.cfg.RootPath, "resources", "{file}"),
		http.StripPrefix(
			"/"+handler.cfg.RootPath,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Duration of 1 year, since we use cachebusting anyway.
				w.Header().Set("Cache-Control", "public, max-age=31536000")

				fileServer.ServeHTTP(w, r)
			}),
		).ServeHTTP,
	)
	registerWithCsp("GET", path.Join(handler.cfg.RootPath, "lobby.js"), handler.lobbyJs)
	registerWithCsp("GET", path.Join(handler.cfg.RootPath, "index.js"), handler.indexJs)
	registerWithCsp("GET", path.Join(handler.cfg.RootPath, "lobby", "{lobby_id}"), handler.ssrEnterLobby)
	registerWithCsp("POST", path.Join(handler.cfg.RootPath, "lobby"), handler.ssrCreateLobby)
}

// errorPageData represents the data that error.html requires to be displayed.
type errorPageData struct {
	*BasePageConfig
	// ErrorMessage displayed on the page.
	ErrorMessage string

	Translation *translations.Translation
	Locale      string
}

// userFacingError will return the occurred error as a custom html page to the caller.
func (handler *SSRHandler) userFacingError(w http.ResponseWriter, errorMessage string, translation *translations.Translation) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", &errorPageData{
		BasePageConfig: handler.basePageConfig,
		ErrorMessage:   errorMessage,
		Translation:    translation,
	})
	// This should never happen, but if it does, something is very wrong.
	if err != nil {
		panic(err)
	}
}

func isHumanAgent(userAgent string) bool {
	return strings.Contains(userAgent, "gecko") ||
		strings.Contains(userAgent, "chrome") ||
		strings.Contains(userAgent, "opera") ||
		strings.Contains(userAgent, "safari")
}
