package frontend

import (
	"bytes"
	"testing"

	"github.com/scribble-rs/scribble.rs/api"
	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/translations"
)

func Test_templateLobbyPage(t *testing.T) {
	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"lobby-page", &lobbyPageData{
			BasePageConfig: currentBasePageConfig,
			LobbyData: &api.LobbyData{
				EditableLobbySettings: &game.EditableLobbySettings{},
				SettingBounds:         game.LobbySettingBounds,
			},
			Translation: translations.DefaultTranslation,
		})
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}

func Test_templateErrorPage(t *testing.T) {
	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"error-page", &errorPageData{
			BasePageConfig: currentBasePageConfig,
			ErrorMessage:   "KEK",
			Translation:    translations.DefaultTranslation,
			Locale:         "en-US",
		})
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}

func Test_templateRobotPage(t *testing.T) {
	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"robot-page", &lobbyPageData{
			BasePageConfig: currentBasePageConfig,
			LobbyData: &api.LobbyData{
				EditableLobbySettings: &game.EditableLobbySettings{
					MaxPlayers: 12,
					Rounds:     4,
				},
			},
			Translation: translations.DefaultTranslation,
			Locale:      "en-US",
		})
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}

func Test_templateLobbyCreatePage(t *testing.T) {
	createPageData := createDefaultLobbyCreatePageData()
	createPageData.Translation = translations.DefaultTranslation

	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"lobby-create-page", createPageData)
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}
