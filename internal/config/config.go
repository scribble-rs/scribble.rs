package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/subosito/gotenv"
)

type LobbySettingDefaults struct {
	Public            string `env:"PUBLIC"`
	DrawingTime       string `env:"DRAWING_TIME"`
	Rounds            string `env:"ROUNDS"`
	MaxPlayers        string `env:"MAX_PLAYERS"`
	CustomWords       string `env:"CUSTOM_WORDS"`
	CustomWordsChance string `env:"CUSTOM_WORDS_CHANCE"`
	ClientsPerIPLimit string `env:"CLIENTS_PER_IP_LIMIT"`
	EnableVotekick    string `env:"ENABLE_VOTEKICK"`
	Language          string `env:"LANGUAGE"`
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
	RootPath       string `env:"ROOT_PATH"`
	CPUProfilePath string `env:"CPU_PROFILE_PATH"`
	// LobbySettingDefaults is used for the server side rendering of the lobby
	// creation page. It doesn't affect the default values of lobbies created
	// via the API.
	LobbySettingDefaults LobbySettingDefaults `envPrefix:"LOBBY_SETTING_DEFAULTS_"`
	Port                 uint16               `env:"PORT"`
}

var Default = Config{
	Port: 8080,
	LobbySettingDefaults: LobbySettingDefaults{
		Public:            "false",
		DrawingTime:       "120",
		Rounds:            "4",
		MaxPlayers:        "12",
		CustomWordsChance: "50",
		ClientsPerIPLimit: "1",
		EnableVotekick:    "true",
		Language:          "english",
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
	config.RootPath = strings.Trim(config.RootPath, "/")

	return &config, nil
}
