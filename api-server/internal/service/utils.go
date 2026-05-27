package service

import "regexp"

var regexSpecialChars = regexp.MustCompile(`[[\]{}()*+?.\\^$|]`)

func SanitizeRegex(s string) string {
	return regexSpecialChars.ReplaceAllString(s, `\$&`)
}
