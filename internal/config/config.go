package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/scribble-rs/scribble.rs/internal/game"
	"github.com/subosito/gotenv"
)

type LobbySettingDefaults struct {
	Public             string `env:"PUBLIC"`
	DrawingTime        string `env:"DRAWING_TIME"`
	Rounds             string `env:"ROUNDS"`
	MaxPlayers         string `env:"MAX_PLAYERS"`
	CustomWords        string `env:"CUSTOM_WORDS"`
	CustomWordsPerTurn string `env:"CUSTOM_WORDS_PER_TURN"`
	ClientsPerIPLimit  string `env:"CLIENTS_PER_IP_LIMIT"`
	Language           string `env:"LANGUAGE"`
	ScoreCalculation   string `env:"SCORE_CALCULATION"`
}

type CORS struct {
	AllowedOrigins   []string `env:"ALLOWED_ORIGINS"`
	AllowCredentials bool     `env:"ALLOW_CREDENTIALS"`
}

type LobbyCleanup struct {
	// Interval is the interval in which the cleanup routine will run. If set
	// to `0`, the cleanup routine will be disabled.
	Interval time.Duration `env:"INTERVAL"`
	// PlayerInactivityThreshold is the time after which a player counts as
	// inactivity and won't keep the lobby up. Note that cleaning up a lobby can
	// therefore take up to Interval + PlayerInactivityThreshold.
	PlayerInactivityThreshold time.Duration `env:"PLAYER_INACTIVITY_THRESHOLD"`
}

type Config struct {
	// NetworkAddress is empty by default, since that implies listening on
	// all interfaces. For development usecases, on windows for example, this
	// is very annoying, as windows will nag you with firewall prompts.
	NetworkAddress string `env:"NETWORK_ADDRESS"`
	// RootPath is the path directly after the domain and before the
	// scribblers paths. For example if you host scribblers on painting.com
	// but already host a different website on that domain, then your API paths
	// might have to look like this: painting.com/scribblers/v1
	RootPath string `env:"ROOT_PATH"`
	// RootURL is similar to RootPath, but contains only the protocol and
	// domain. So it could be https://painting.com. This is required for some
	// non critical functionality, such as metadata tags.
	RootURL string `env:"ROOT_URL"`
	// CanonicalURL specifies the original domain, in case we are accessing the
	// site via some other domain, such as scribblers.fly.dev
	CanonicalURL string `env:"CANONICAL_URL"`
	// AllowIndexing will control whether the noindex, nofollow meta tag is
	// added to the home page.
	AllowIndexing bool `env:"ALLOW_INDEXING"`
	// ServeDirectories is a map of `path` to `directory`. All directories are
	// served under the given path.
	ServeDirectories map[string]string `env:"SERVE_DIRECTORIES"`
	CPUProfilePath   string            `env:"CPU_PROFILE_PATH"`
	// LobbySettingDefaults is used for the server side rendering of the lobby
	// creation page. It doesn't affect the default values of lobbies created
	// via the API.
	LobbySettingDefaults LobbySettingDefaults `envPrefix:"LOBBY_SETTING_DEFAULTS_"`
	LobbySettingBounds   game.SettingBounds   `envPrefix:"LOBBY_SETTING_BOUNDS_"`
	Port                 uint16               `env:"PORT"`
	CORS                 CORS                 `envPrefix:"CORS_"`
	LobbyCleanup         LobbyCleanup         `envPrefix:"LOBBY_CLEANUP_"`
}

var Default = Config{
	Port: 8080,
	LobbySettingDefaults: LobbySettingDefaults{
		Public:             "false",
		DrawingTime:        "120",
		Rounds:             "4",
		MaxPlayers:         "24",
		CustomWordsPerTurn: "3",
		ClientsPerIPLimit:  "2",
		Language:           "english",
		ScoreCalculation:   "chill",
	},
	LobbySettingBounds: game.SettingBounds{
		MinDrawingTime:        60,
		MaxDrawingTime:        300,
		MinRounds:             1,
		MaxRounds:             20,
		MinMaxPlayers:         2,
		MaxMaxPlayers:         24,
		MinClientsPerIPLimit:  1,
		MaxClientsPerIPLimit:  24,
		MinCustomWordsPerTurn: 1,
		MaxCustomWordsPerTurn: 3,
	},
	CORS: CORS{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: false,
	},
	LobbyCleanup: LobbyCleanup{
		Interval:                  90 * time.Second,
		PlayerInactivityThreshold: 75 * time.Second,
	},
}

// Load loads the configuration from the environment. If a .env file is
// available, it will be loaded as well. Values found in the environment
// will overwrite whatever is load from the .env file.
func Load() (*Config, error) {
	envVars := make(map[string]string)

	dotEnvPath := ".env"
	if _, err := os.Stat(dotEnvPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error checking for existence of .env file: %w", err)
		}
	} else {
		envFileContent, err := gotenv.Read(dotEnvPath)
		if err != nil {
			return nil, fmt.Errorf("error reading .env file: %w", err)
		}
		for key, value := range envFileContent {
			envVars[key] = value
		}
	}

	// Add local environment variables to EnvVars map
	for _, keyValuePair := range os.Environ() {
		pair := strings.SplitN(keyValuePair, "=", 2)
		// For some reason, gitbash can contain the variable `=::=::\` which
		// gives us a pair where the first entry is empty.
		if pair[0] == "" {
			continue
		}
		envVars[pair[0]] = pair[1]
	}

	config := Default
	if err := env.ParseWithOptions(&config, env.Options{
		Environment: envVars,
		OnSet: func(key string, value any, isDefault bool) {
			if !reflect.ValueOf(value).IsZero() {
				log.Printf("Setting '%s' to '%v' (isDefault: %v)\n", key, value, isDefault)
			}
		},
	}); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	// Prevent user error and let the code decide when we need slashes.
	config.RootURL = strings.TrimSuffix(config.RootURL, "/")
	if config.CanonicalURL == "" {
		config.CanonicalURL = config.RootURL
	}
	config.RootPath = strings.Trim(config.RootPath, "/")

	return &config, nil
}
