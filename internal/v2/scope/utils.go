package scope

import (
	"os"
	"strings"
)

// Env returns environment variables in key-value format usable for globals.
func Env() map[string]string {
	env := os.Environ()
	out := make(map[string]string, len(env))
	for _, str := range env {
		chunks := strings.SplitN(str, "=", 2)
		key := chunks[0]
		if len(chunks) == 1 {
			out[key] = ""
			continue
		}

		out[key] = chunks[1]
	}

	return out
}
