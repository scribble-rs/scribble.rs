package translations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_noObsoleteKeys(t *testing.T) {
	for language, translationMap := range translationRegistry {
		t.Run(fmt.Sprintf("obsolete_keys_check_%s", language), func(t *testing.T) {
			for key := range translationMap {
				_, englishContainsKey := DefaultTranslation[key]
				assert.True(t, englishContainsKey, "key %s is obsolete", key)
			}
		})
	}
}
