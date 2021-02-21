package communication

import (
	"fmt"
	"net/http"
	"strings"
)

var CurrentBasePageConfig = &BasePageConfig{}

// BasePageConfig is data that all pages require to function correctly, no matter
// whether error page or lobby page.
type BasePageConfig struct {
	// RootPath is the path directly after the domain and before the
	// scribble.rs paths. For example if you host scribblers on painting.com
	// but already host a different website, then your API paths might have to
	// look like this: painting.com/scribblers/v1.
	RootPath string `json:"rootPath"`
}

// ErrorPageData represents the data that error.html requires to be displayed.
type ErrorPageData struct {
	*BasePageConfig
	// ErrorMessage displayed on the page.
	ErrorMessage string
}

//userFacingError will return the occurred error as a custom html page to the caller.
func userFacingError(w http.ResponseWriter, errorMessage string) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", &ErrorPageData{
		BasePageConfig: CurrentBasePageConfig,
		ErrorMessage:   errorMessage,
	})
	//This should never happen, but if it does, something is very wrong.
	if err != nil {
		panic(err)
	}
}

// remoteAddressToSimpleIP removes unnecessary clutter from the input,
// reducing it to a simple IPv4. We expect two different formats here.
// One being http.Request#RemoteAddr (127.0.0.1:12345) and the other
// being forward headers, which contain brackets, as there's no proper
// API, but just a string that needs to be parsed.
func remoteAddressToSimpleIP(input string) string {
	address := input
	lastIndexOfDoubleColon := strings.LastIndex(address, ":")
	if lastIndexOfDoubleColon != -1 {
		address = address[:lastIndexOfDoubleColon]
	}

	return strings.TrimSuffix(strings.TrimPrefix(address, "["), "]")
}

// Serve will start an HTTP server listening on the given port.
// This is a blocking call-
func Serve(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
