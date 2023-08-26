package frontend

import (
	"bytes"
	"testing"

	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/translations"
)

func init() {
	if err := Init(); err != nil {
		panic(err)
	}
}

func Test_templateLobbyPage(t *testing.T) {
	var buffer bytes.Buffer
	err := pageTemplates.ExecuteTemplate(&buffer,
		"lobby-page", &lobbyPageData{
			BasePageConfig: &BasePageConfig{
				CacheBust: "lol",
			},
			LobbyData: &api.LobbyData{
				SettingBounds: game.LobbySettingBounds,
			},
			Translation: translations.DefaultTranslation,
		})
	if err != nil {
		t.Errorf("Error templating: %s", err)
	}
}

func Test_templateErrorPage(t *testing.T) {
	var buffer bytes.Buffer
	err := pageTemplates.ExecuteTemplate(&buffer,
		"error-page", &errorPageData{
			BasePageConfig: &BasePageConfig{},
			ErrorMessage:   "KEK",
			Translation:    translations.DefaultTranslation,
			Locale:         "en-US",
		})
	if err != nil {
		t.Errorf("Error templating: %s", err)
	}
}

func Test_templateRobotPage(t *testing.T) {
	var buffer bytes.Buffer
	err := pageTemplates.ExecuteTemplate(&buffer,
		"robot-page", &lobbyPageData{
			BasePageConfig: &BasePageConfig{},
			LobbyData: &api.LobbyData{
				EditableLobbySettings: game.EditableLobbySettings{
					MaxPlayers: 12,
					Rounds:     4,
				},
			},
			Translation: translations.DefaultTranslation,
			Locale:      "en-US",
		})
	if err != nil {
		t.Errorf("Error templating: %s", err)
	}
}

func Test_templateLobbyCreatePage(t *testing.T) {
	createPageData := createDefaultLobbyCreatePageData(config.Default.LobbySettingDefaults)
	createPageData.Translation = translations.DefaultTranslation

	var buffer bytes.Buffer
	if err := pageTemplates.ExecuteTemplate(&buffer, "lobby-create-page", createPageData); err != nil {
		t.Errorf("Error templating: %s", err)
	}
}
