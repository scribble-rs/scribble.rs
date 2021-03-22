package frontend

import (
	"net/http"

	"github.com/scribble-rs/scribble.rs/api"
)

// ErrorPageData represents the data that error.html requires to be displayed.
type ErrorPageData struct {
	*api.BasePageConfig
	// ErrorMessage displayed on the page.
	ErrorMessage string
}

//userFacingError will return the occurred error as a custom html page to the caller.
func userFacingError(w http.ResponseWriter, errorMessage string) {
	err := pageTemplates.ExecuteTemplate(w, "error-page", &ErrorPageData{
		BasePageConfig: api.CurrentBasePageConfig,
		ErrorMessage:   errorMessage,
	})
	//This should never happen, but if it does, something is very wrong.
	if err != nil {
		panic(err)
	}
}
