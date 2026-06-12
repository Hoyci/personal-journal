package processing

import "regexp"

var (
	htmlTag    = regexp.MustCompile(`<[^>]+>`)
	multiSpace = regexp.MustCompile(`\s+`)
)

func StripHTML(s string) string {
	s = htmlTag.ReplaceAllString(s, " ")
	s = multiSpace.ReplaceAllString(s, " ")
	return s
}
