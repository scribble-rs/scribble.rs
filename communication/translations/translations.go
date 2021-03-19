package translations

import (
	"fmt"
)

// init initializes all localization packs. Each new package has to be added
// to this function.
func init() {
	//We are making sure to add english first, since it's the default.
	DefaultTranslation = initDefaultTranslation()
	initGermanTranslation()
}

var translationRegistry = make(map[string]Translation)

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

// Put adds a new key to the translation. If the key already exists, the
// server panics. This happens on startup, therefore it's safe.
func (translation Translation) Put(key, value string) {
	_, avail := translation[key]
	if avail {
		panic(fmt.Sprintf("Duplicate key '%s'", key))
	}

	translation[key] = value
}

// GetLanguage retrieves a translation pack or nil if the desired package
// couldn't be found.
func GetLanguage(language string) Translation {
	return translationRegistry[language]
}

// RegisterTranslation makes adds a language to the registry and makes
// it available via Get. If the language is already registered, the server
// panics. This happens on startup, therefore it's safe.
func RegisterTranslation(language string, translation Translation) {
	_, avail := translationRegistry[language]
	if avail {
		panic(fmt.Sprintf("Language '%s' has been registered multiple times", language))
	}

	if DefaultTranslation != nil {
		for key := range translation {
			_, fallbackValueAvail := DefaultTranslation[key]
			if !fallbackValueAvail {
				panic(fmt.Sprintf("Language key '%s' in language '%s' has no default translation value in 'en_US'", key, language))
			}
		}
	}

	translationRegistry[language] = translation
}

func createTranslation() Translation {
	return make(map[string]string)
}
