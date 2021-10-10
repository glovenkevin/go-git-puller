package utils

import "regexp"

var StringOnlyPattern *regexp.Regexp

func init() {
	StringOnlyPattern, _ = regexp.Compile(PATERN_STRING_ONLY)
}
