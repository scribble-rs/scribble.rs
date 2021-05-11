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

func ParsePlayerName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed, errors.New("اسم بازیکن نمیتواند خالی باشد")
	}

	return trimmed, nil
}

func ParseLanguage(value string) (string, error) {
	toLower := strings.ToLower(strings.TrimSpace(value))
	for languageKey := range game.SupportedLanguages {
		if toLower == languageKey {
			return languageKey, nil
		}
	}

	return "", errors.New("the given language doesn't match any supported language")
}

func ParseDrawingTime(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("زمان هر نقاشی باید یک عدد باشد")
	}

	if result < game.LobbySettingBounds.MinDrawingTime {
		return 0, fmt.Errorf("زمان هر نقاشی نباید از %d کمتر باشد.", game.LobbySettingBounds.MinDrawingTime)
	}

	if result > game.LobbySettingBounds.MaxDrawingTime {
		return 0, fmt.Errorf("زمان هر نقاشی نباید از %d بیشتر باشد", game.LobbySettingBounds.MaxDrawingTime)
	}

	return int(result), nil
}

func ParseRounds(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("تعداد نوبت ها باید عدد باشد")
	}

	if result < game.LobbySettingBounds.MinRounds {
		return 0, fmt.Errorf("تعداد نوبت ها نمیتواند از %d کمتر باشد", game.LobbySettingBounds.MinRounds)
	}

	if result > game.LobbySettingBounds.MaxRounds {
		return 0, fmt.Errorf("تعداد نوبت ها نمیتواند از %d بیشتر باشد", game.LobbySettingBounds.MaxRounds)
	}

	return int(result), nil
}

func ParseMaxPlayers(value string) (int, error) {
	result, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil {
		return 0, errors.New("حداکثر تعداد بازیکنان باید عدد باشد")
	}

	if result < game.LobbySettingBounds.MinMaxPlayers {
		return 0, fmt.Errorf("حداکثر تعداد بازیکنان نباید از %d کمتر باشد", game.LobbySettingBounds.MinMaxPlayers)
	}

	if result > game.LobbySettingBounds.MaxMaxPlayers {
		return 0, fmt.Errorf("حداکثر تعداد بازیکنان نباید از %d بیشتر باشد", game.LobbySettingBounds.MaxMaxPlayers)
	}

	return int(result), nil
}

func ParseCustomWords(value string) ([]string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	lowercaser := cases.Lower(language.English)
	result := strings.Split(trimmedValue, "-")
	for index, item := range result {
		trimmedItem := lowercaser.String(strings.TrimSpace(item))
		if trimmedItem == "" {
			return nil, errors.New("هیچ کلمه خودی ای نباید بی حرف باشد")
		}
		result[index] = trimmedItem
	}

	return result, nil
}

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

func ParseBoolean(valueName string, value string) (bool, error) {
	if strings.EqualFold(value, "true") {
		return true, nil
	}

	if strings.EqualFold(value, "false") {
		return false, nil
	}

	if value == "" {
		return false, nil
	}

	return false, fmt.Errorf("the %s value must be a boolean value ('true' or 'false)", valueName)
}
