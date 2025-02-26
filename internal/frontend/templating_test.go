package frontend

import (
	"bytes"
	"testing"

	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/scribble-rs/scribble.rs/internal/translations"
	"github.com/stretchr/testify/require"
)

func Test_templateLobbyPage(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	err := pageTemplates.ExecuteTemplate(&buffer,
		"lobby-page", &lobbyPageData{
			BasePageConfig: &BasePageConfig{
				checksums: make(map[string]string),
			},
			LobbyData: &api.LobbyData{
				SettingBounds: config.Default.LobbySettingBounds,
				GameConstants: api.GameConstantsData,
			},
			Translation: translations.DefaultTranslation,
		})
	if err != nil {
		t.Errorf("Error templating: %s", err)
	}
}

func Test_templateErrorPage(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func Test_templateIndexPage(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(&config.Config{})
	require.NoError(t, err)
	createPageData := handler.createDefaultIndexPageData()
	createPageData.Translation = translations.DefaultTranslation

	var buffer bytes.Buffer
	if err := pageTemplates.ExecuteTemplate(&buffer, "index", createPageData); err != nil {
		t.Errorf("Error templating: %s", err)
	}
}
