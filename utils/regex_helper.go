package utils

import "regexp"

// Compiled regex expresion of String only regex pattern
var StringOnlyPattern *regexp.Regexp

func init() {
	StringOnlyPattern, _ = regexp.Compile(PATERN_STRING_ONLY)
}
