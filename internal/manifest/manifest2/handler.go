package manifest2

import (
	"errors"
	"fmt"
	"strings"
)

const namespaceDelimiter = "/"

type JobHandlerRef struct {
	// Namespace is import name of a plugin.
	Namespace string

	// Name is action or mixin name to use to run job.
	Name string
}

// SplitActionName splits string representation of action name and namespace.
//
// For example:
//
//	"foo/bar" -> namespace: "foo", name: "bar"
func SplitActionName(str string) (JobHandlerRef, error) {
	var dst JobHandlerRef
	chunks := strings.SplitN(str, namespaceDelimiter, 2)
	if len(chunks) != 2 {
		return dst, errors.New("action name should be in format {namespace}/{action}")
	}

	dst.Namespace = chunks[0]
	dst.Name = chunks[1]
	if dst.Name == "" {
		return dst, errors.New("empty action name")
	}

	if dst.Namespace == "" {
		return dst, errors.New("empty namespace")
	}

	return dst, nil
}

// ValidateTaskName checks whether task or mixin name is correct.
func ValidateTaskName(str string) error {
	if ActionNameHasNamespace(str) {
		return fmt.Errorf("task and mixin names shouldn't contain %q character", namespaceDelimiter)
	}

	return nil
}

// ActionNameHasNamespace checks whether action name string contains '/' symbol
// which is import namespace delimiter.
func ActionNameHasNamespace(str string) bool {
	return strings.Contains(str, namespaceDelimiter)
}
