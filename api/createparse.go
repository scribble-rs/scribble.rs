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
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the drawing time must be numeric")
	}

	if result < game.LobbySettingBounds.MinDrawingTime {
		return 0, fmt.Errorf("drawing time must not be smaller than %d", game.LobbySettingBounds.MinDrawingTime)
	}

	if result > game.LobbySettingBounds.MaxDrawingTime {
		return 0, fmt.Errorf("drawing time must not be greater than %d", game.LobbySettingBounds.MaxDrawingTime)
	}

	return int(result), nil
}

// ParseRounds checks whether the given value is an integer between
// the lower and upper bound of rounds played. All other invalid
// input, including empty strings, will return an error.
func ParseRounds(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the rounds amount must be numeric")
	}

	if result < game.LobbySettingBounds.MinRounds {
		return 0, fmt.Errorf("rounds must not be smaller than %d", game.LobbySettingBounds.MinRounds)
	}

	if result > game.LobbySettingBounds.MaxRounds {
		return 0, fmt.Errorf("rounds must not be greater than %d", game.LobbySettingBounds.MaxRounds)
	}

	return int(result), nil
}

// ParseMaxPlayers checks whether the given value is an integer between
// the lower and upper bound of maximum players per lobby. All other invalid
// input, including empty strings, will return an error.
func ParseMaxPlayers(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the max players amount must be numeric")
	}

	if result < game.LobbySettingBounds.MinMaxPlayers {
		return 0, fmt.Errorf("maximum players must not be smaller than %d", game.LobbySettingBounds.MinMaxPlayers)
	}

	if result > game.LobbySettingBounds.MaxMaxPlayers {
		return 0, fmt.Errorf("maximum players must not be greater than %d", game.LobbySettingBounds.MaxMaxPlayers)
	}

	return int(result), nil
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
// input, including empty strings, will return an error. The valueName
// parameter is only needed for the error message, since the contents of the
// error messages are meant to be shown to the client.
func ParseClientsPerIPLimit(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the clients per IP limit must be numeric")
	}

	if result < game.LobbySettingBounds.MinClientsPerIPLimit {
		return 0, fmt.Errorf("the clients per IP limit must not be lower than %d", game.LobbySettingBounds.MinClientsPerIPLimit)
	}

	if result > game.LobbySettingBounds.MaxClientsPerIPLimit {
		return 0, fmt.Errorf("the clients per IP limit must not be higher than %d", game.LobbySettingBounds.MaxClientsPerIPLimit)
	}

	return int(result), nil
}

// ParseCustomWordsChance checks whether the given value is an integer between
// 0 and 100. All other invalid input, including empty strings, will return an
// error. The valueName parameter is only needed for the error message, since
// the contents of the error messages are meant to be shown to the client.
func ParseCustomWordsChance(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("the custom word chance must be numeric")
	}

	if result < 0 {
		return 0, errors.New("custom word chance must not be lower than 0")
	}

	if result > 100 {
		return 0, errors.New("custom word chance must not be higher than 100")
	}

	return int(result), nil
}

// ParseBoolean checks whether the given value is either "true" or "false".
// The checks are case-insensitive. If an empty string is supplied, false
// is returned. All other invalid input will return an error.
// The valueName parameter is only needed for the error message, since the
// contents of the error messages are meant to be shown to the client.
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
