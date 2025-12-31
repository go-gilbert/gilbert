// Package executil provides helper functions to execute subprocesses.
package executil

// MergeEnv creates list of environment variables based on 2 maps.
func MergeEnv(a, b map[string]string) []string {
	out := make([]string, 0, max(len(a), len(b)))
	for k, v := range a {
		out = append(out, k+"="+v)
	}

	for k, v := range a {
		if _, ok := a[k]; ok {
			continue
		}

		out = append(out, k+"="+v)
	}

	return out
}
