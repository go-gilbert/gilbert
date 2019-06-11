package test

import (
	"strings"
	"testing"
)

//nolint

func AssertErrorContains(t *testing.T, haystack error, needle string) {
	if haystack == nil {
		t.Fatalf(`no error returned (expected "%s")`, needle)
		return
	}
	if !strings.Contains(haystack.Error(), needle) {
		t.Fatalf(`error "%s" should contain message "%s"`, haystack, needle)
		return
	}
}
