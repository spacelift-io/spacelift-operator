package slug

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	emoji "github.com/tmdvs/Go-Emoji-Utils"
)

const maxSlugSize = 256

// SafeSlug turns input into a slug with a safe predefined length.
func SafeSlug(input string) string {
	input = replaceEmojis(input)
	ret := slug.Make(input)

	if len(ret) > maxSlugSize {
		ret = ret[0:maxSlugSize]
	}

	return ret
}

func replaceEmojis(s string) string {
	emojis := emoji.FindAll(s)

	for _, i := range emojis {
		emo := i.Match
		s = strings.ReplaceAll(s, emo.Value, fmt.Sprintf("-emoji-%s-", emo.Descriptor))
	}

	return s
}
