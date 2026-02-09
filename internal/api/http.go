package api

import (
	"net/http"
	"path"
	"strings"

	"github.com/scribble-rs/scribble.rs/internal/metrics"
)

// SetupRoutes registers the /v1/ endpoints with the http package.
func (handler *V1Handler) SetupRoutes(rootPath string, register func(string, string, http.HandlerFunc)) {
	v1 := path.Join(rootPath, "v1")

	metrics.SetupRoute(func(metricsHandler http.HandlerFunc) {
		register("GET", path.Join(v1, "metrics"), metricsHandler)
	})
	register("GET", path.Join(v1, "stats"), handler.getStats)

	// These exist only for the public API. We version them in order to ensure
	// backwards compatibility as far as possible.
	register("GET", path.Join(v1, "lobby"), handler.getLobbies)
	register("POST", path.Join(v1, "lobby"), handler.postLobby)

	register("GET", path.Join(v1, "lobby", "{lobby_id}", "gallery"), handler.getGallery)
	register("PATCH", path.Join(v1, "lobby", "{lobby_id}"), handler.patchLobby)
	// We support both path parameter and cookie.
	register("PATCH", path.Join(v1, "lobby"), handler.patchLobby)

	// The websocket is shared between the public API and the official client
	register("GET", path.Join(v1, "lobby", "{lobby_id}", "ws"), handler.websocketUpgrade)
	// We support both path parameter and cookie.
	register("GET", path.Join(v1, "lobby", "ws"), handler.websocketUpgrade)

	register("POST", path.Join(v1, "lobby", "{lobby_id}", "player"), handler.postPlayer)
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

// GetIPAddressFromRequest extracts the clients IP address from the request.
// This function respects forwarding headers.
func GetIPAddressFromRequest(request *http.Request) string {
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded

	// The following logic has been implemented according to the spec, therefore please
	// refer to the spec if you have a question.

	forwardedAddress := request.Header.Get("X-Forwarded-For")
	if forwardedAddress != "" {
		// Since the field may contain multiple addresses separated by commas, we use the first
		// one, which according to the docs is supposed to be the client address.
		clientAddress := strings.TrimSpace(strings.Split(forwardedAddress, ",")[0])
		return remoteAddressToSimpleIP(clientAddress)
	}

	standardForwardedHeader := request.Header.Get("Forwarded")
	if standardForwardedHeader != "" {
		targetPrefix := "for="
		// Since forwarded can contain more than one field, we search for one specific field.
		for part := range strings.SplitSeq(standardForwardedHeader, ";") {
			trimmed := strings.TrimSpace(part)
			if after, ok := strings.CutPrefix(trimmed, targetPrefix); ok {
				// FIXME Maybe checking for a valid IP-Address would make sense here, not sure tho.
				address := remoteAddressToSimpleIP(after)
				// Since the documentation doesn't mention which quotes are used, I just remove all ;)
				return strings.NewReplacer("`", "", "'", "", "\"", "", "[", "", "]", "").Replace(address)
			}
		}
	}

	return remoteAddressToSimpleIP(request.RemoteAddr)
}

func SetDiscordCookie(
	w http.ResponseWriter,
	key, value string,
) {
	http.SetCookie(w, &http.Cookie{
		Name:        key,
		Value:       value,
		Domain:      discordDomain,
		Path:        "/",
		SameSite:    http.SameSiteNoneMode,
		Partitioned: true,
		Secure:      true,
	})
}
func SetNormalCookie(
	w http.ResponseWriter,
	key, value string,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     key,
		Value:    value,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}
