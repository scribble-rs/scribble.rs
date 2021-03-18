package translations

import "fmt"

var translations = make(map[string]Translation)

// DefaultTranslation is the fallback translation for cases where the users
// preferred language can't be found. This value is never returned by Get, but
// has to be retrieved manually if desired. Currently, this is en-US.
var DefaultTranslation Translation

// Translation represents key - value pairs of translated user interface
// strings.
type Translation map[string]string

// Get retrieves a translated string or the default string if none could
// be found.
func (translation Translation) Get(key string) string {
	value, avail := translation[key]
	if avail {
		return value
	}

	fallbackValue, fallbackAvail := DefaultTranslation[key]
	if fallbackAvail {
		return fallbackValue
	}

	panic(fmt.Sprintf("no translation value available for key '%s'", key))
}

// Get retrieves a translation pack or nil if the desired package couldn't
// be found.
func Get(language string) Translation {
	return translations[language]
}
