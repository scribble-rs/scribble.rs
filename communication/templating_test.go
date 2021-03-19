package communication

import (
	"bytes"
	"testing"

	"github.com/scribble-rs/scribble.rs/communication/translations"
	"github.com/scribble-rs/scribble.rs/game"
)

func Test_templateLobbyPage(t *testing.T) {
	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"lobby-page", &LobbyPageData{
			LobbyData: &LobbyData{
				BasePageConfig:        CurrentBasePageConfig,
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
		"error-page", &ErrorPageData{
			BasePageConfig: CurrentBasePageConfig,
			ErrorMessage:   "KEK",
		})
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}

func Test_templateRobotPage(t *testing.T) {
	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"robot-page", &LobbyData{
			BasePageConfig: CurrentBasePageConfig,
			EditableLobbySettings: &game.EditableLobbySettings{
				MaxPlayers: 12,
				Rounds:     4,
			},
		})
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}

func Test_templateLobbyCreatePage(t *testing.T) {
	var buffer bytes.Buffer
	templatingError := pageTemplates.ExecuteTemplate(&buffer,
		"lobby-create-page", createDefaultLobbyCreatePageData())
	if templatingError != nil {
		t.Errorf("Error templating: %s", templatingError)
	}
}
