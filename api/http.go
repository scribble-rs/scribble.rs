package api

import (
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

// GetIPADdressFromRequest extracts the clients IP address from the request.
// This function respects forwarding headers.
func GetIPAddressFromRequest(r *http.Request) string {
	//See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	//See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded

	//The following logic has been implemented according to the spec, therefore please
	//refer to the spec if you have a question.

	forwardedAddress := r.Header.Get("X-Forwarded-For")
	if forwardedAddress != "" {
		//Since the field may contain multiple addresses separated by commas, we use the first
		//one, which according to the docs is supposed to be the client address.
		clientAddress := strings.TrimSpace(strings.Split(forwardedAddress, ",")[0])
		return remoteAddressToSimpleIP(clientAddress)
	}

	standardForwardedHeader := r.Header.Get("Forwarded")
	if standardForwardedHeader != "" {
		targetPrefix := "for="
		//Since forwarded can contain more than one field, we search for one specific field.
		for _, part := range strings.Split(standardForwardedHeader, ";") {
			trimmed := strings.TrimSpace(part)
			if strings.HasPrefix(trimmed, targetPrefix) {
				//FIXME Maybe checking for a valid IP-Address would make sense here, not sure tho.
				address := remoteAddressToSimpleIP(strings.TrimPrefix(trimmed, targetPrefix))
				//Since the documentation doesn't mention which quotes are used, I just remove all ;)
				return strings.NewReplacer("`", "", "'", "", "\"", "", "[", "", "]", "").Replace(address)
			}
		}
	}

	return remoteAddressToSimpleIP(r.RemoteAddr)
}
