package communication

import (
	"fmt"
	"net/http"
	"strings"
)

//userFacingError will return the occurred error as a custom html page to the caller.
func userFacingError(w http.ResponseWriter, errorMessage string) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", errorMessage)
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
