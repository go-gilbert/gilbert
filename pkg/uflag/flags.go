package uflag

// WriteBoolFlag decodes boolean flag value and writes to a specified bool reference.
func WriteBoolFlag(dst *bool, val string) {
	switch val {
	case "1", "", "true":
		*dst = true
	case "0", "false":
		*dst = false
	default:
		return
	}
}
