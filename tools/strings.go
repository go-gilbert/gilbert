package tools

import "strings"

const EmptyString = ""

// StringEmpty checks if passed string is not empty (including spaces)
func StringEmpty(str string) bool {
	return strings.TrimSpace(str) == EmptyString
}

func OneOfStrings(strings ...string) string {
	for _, str := range strings {
		if str != EmptyString {
			return str
		}
	}

	return EmptyString
}
