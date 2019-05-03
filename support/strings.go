package support

import "strings"

// EmptyString is a sample of empty string
const EmptyString = ""

// StringEmpty checks if passed string is not empty (including spaces)
func StringEmpty(str string) bool {
	return strings.TrimSpace(str) == EmptyString
}

// OneOfStrings returns a first non-empty string from set of strings
func OneOfStrings(strings ...string) string {
	for _, str := range strings {
		if str != EmptyString {
			return str
		}
	}

	return EmptyString
}
