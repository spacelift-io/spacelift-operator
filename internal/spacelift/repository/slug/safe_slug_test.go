package slug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeSlug_OK(t *testing.T) {
	assert.Equal(t, "bacon-bacon", SafeSlug("bacon bacon"))
}

func TestSafeSlug_OK_Too_Long(t *testing.T) {
	var input string

	for i := 0; i < 53; i++ {
		input += "bacon "
	}

	slug := SafeSlug(input)

	assert.Equal(t, 256, len(slug))
}

func TestSafeSlug_OK_Name_Contains_Emoji(t *testing.T) {
	assert.Equal(t, "turtle-emoji-turtle", SafeSlug("turtle 🐢"))
}

func TestSafeSlug_OK_Name_Only_Contains_Emojis(t *testing.T) {
	assert.Equal(t, "emoji-turtle", SafeSlug("🐢"))
}

func TestSafeSlug_OK_Special_Characters(t *testing.T) {
	assert.Equal(t, "assdasdaccasadwsd", SafeSlug("ąśśdasdaććąsadwsd"))
}

func TestSafeSlug_OK_Special_Characters_With_Emoji(t *testing.T) {
	assert.Equal(t, "assdasdaccasadwsd-emoji-turtle", SafeSlug("ąśśdasdaććąsadwsd🐢"))
}
