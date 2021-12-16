package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ParsePlayerName checks if the given value is a valid playername. Currently
// this only includes checkin whether the value is empty or only consists of
// whitespace character.
func ParsePlayerName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed, errors.New("the player name must not be empty")
	}

	return trimmed, nil
}

// ParseLanguage checks whether the given value is part of the
// game.SupportedLanguages array. The input is trimmed and lowercased.
func ParseLanguage(value string) (string, error) {
	toLower := strings.ToLower(strings.TrimSpace(value))
	for languageKey := range game.SupportedLanguages {
		if toLower == languageKey {
			return languageKey, nil
		}
	}

	return "", errors.New("the given language doesn't match any supported language")
}

// ParseDrawingTime checks whether the given value is an integer between
// the lower and upper bound of drawing time. All other invalid
// input, including empty strings, will return an error.
func ParseDrawingTime(value string) (int, error) {
	return parseIntValue(value, game.LobbySettingBounds.MinDrawingTime,
		game.LobbySettingBounds.MaxDrawingTime, "drawing time")
}

// ParseRounds checks whether the given value is an integer between
// the lower and upper bound of rounds played. All other invalid
// input, including empty strings, will return an error.
func ParseRounds(value string) (int, error) {
	return parseIntValue(value, game.LobbySettingBounds.MinRounds,
		game.LobbySettingBounds.MaxRounds, "rounds")
}

// ParseMaxPlayers checks whether the given value is an integer between
// the lower and upper bound of maximum players per lobby. All other invalid
// input, including empty strings, will return an error.
func ParseMaxPlayers(value string) (int, error) {
	return parseIntValue(value, game.LobbySettingBounds.MinMaxPlayers,
		game.LobbySettingBounds.MaxMaxPlayers, "max players amount")
}

// ParseCustomWords checks whether the given value is a string containing comma
// separated values (or a single word). Empty strings will return an empty
// (nil) array and no error. An error is only returned if there are empty words.
// For example these wouldn't parse:
//     wordone,,wordtwo
//     ,
//     wordone,
func ParseCustomWords(value string) ([]string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	lowercaser := cases.Lower(language.English)
	result := strings.Split(trimmedValue, ",")
	for index, item := range result {
		trimmedItem := lowercaser.String(strings.TrimSpace(item))
		if trimmedItem == "" {
			return nil, errors.New("custom words must not be empty")
		}
		result[index] = trimmedItem
	}

	return result, nil
}

// ParseClientsPerIPLimit checks whether the given value is an integer between
// the lower and upper bound of maximum clients per IP. All other invalid
// input, including empty strings, will return an error.
func ParseClientsPerIPLimit(value string) (int, error) {
	return parseIntValue(value, game.LobbySettingBounds.MinClientsPerIPLimit,
		game.LobbySettingBounds.MaxClientsPerIPLimit, "clients per IP limit")
}

// ParseCustomWordsChance checks whether the given value is an integer between
// 0 and 100. All other invalid input, including empty strings, will return an
// error.
func ParseCustomWordsChance(value string) (int, error) {
	return parseIntValue(value, 0, 100, "custom word chance")
}

func parseIntValue(value string, lower, upper int64, valueName string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, fmt.Errorf("%s must be numeric", valueName)
	}

	if result < lower {
		return 0, fmt.Errorf("%s must not be lower than %d", valueName, lower)
	}

	if result > upper {
		return 0, fmt.Errorf("%s must not be higher than %d", valueName, upper)
	}

	return int(result), nil
}

// ParseBoolean checks whether the given value is either "true" or "false".
// The checks are case-insensitive. If an empty string is supplied, false
// is returned. All other invalid input will return an error.
func ParseBoolean(valueName string, value string) (bool, error) {
	if value == "" {
		return false, nil
	}

	if strings.EqualFold(value, "true") {
		return true, nil
	}

	if strings.EqualFold(value, "false") {
		return false, nil
	}

	return false, fmt.Errorf("the %s value must be a boolean value ('true' or 'false)", valueName)
}
