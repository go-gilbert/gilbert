// Package uflag provides a bare-minimum command-line flag parser.
//
// Unlike pflag, cobra and others, it skips unknown and checks only known flags.
// Main use-case for a package is to read core flags necessary to read job file and build Cobra application.
package uflag

import "strings"

const (
	longFlagPrefix     = "--"
	flagValueDelimiter = '='
)

type FlagConsumer interface {
	// IsBoolFlag returns whether flag is a boolean value.
	//
	// If flag is boolean, flag will be consumed even without a value.
	IsBoolFlag(flagName string) bool

	// IsKnownFlag returns whether a flag is known to a consumer.
	IsKnownFlag(flagName string) bool

	// ConsumeFlag consumes flag and its value.
	ConsumeFlag(flagName, value string)
}

// Parse parses command-line flags using flag consumer.
//
// Due to minimalistic approach of this package, there are multiple limitations:
//   - Unknown flags are omitted.
//   - Shorthand flags are not supported, only `--` flags are allowed.
//   - Any errors are intentionally suppressed.
func Parse(c FlagConsumer, args []string) {
	maxIndex := len(args) - 1
	for i, arg := range args {
		if !strings.HasPrefix(arg, longFlagPrefix) {
			continue
		}

		flagName := arg[len(longFlagPrefix):]
		var flagValue string

		// If flag contains a value (e.g --cwd=...), process immediately.
		// otherwise - consume next value.
		delim := strings.IndexRune(flagName, flagValueDelimiter)
		if delim != -1 {
			flagValue = flagName[delim+1:]
			flagName = flagName[0:delim]
			c.ConsumeFlag(flagName, flagValue)
			continue
		}

		if flagName == "" || !c.IsKnownFlag(flagName) {
			continue
		}

		if i == maxIndex {
			// bool flags work without a value
			if c.IsBoolFlag(flagName) {
				c.ConsumeFlag(flagName, "")
			}

			break
		}

		flagValue = args[i+1]
		if isFlag(flagValue) {
			if !c.IsBoolFlag(flagName) {
				continue
			}

			flagValue = ""
		}

		c.ConsumeFlag(flagName, flagValue)
	}
}

func isFlag(v string) bool {
	return strings.HasPrefix(v, longFlagPrefix) || strings.HasPrefix(v, "-")
}
