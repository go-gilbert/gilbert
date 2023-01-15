package cmd

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/internal/config"
	"strconv"
	"strings"
)

const (
	flagCwd       = "cwd"
	flagSpecFile  = "spec-file"
	flagLogFormat = "log-format"
	flagVerbose   = "verbose"

	flagPrefix = "--"
)

// ParsePreRunFlags parses global flags that are core dependencies to build cobra application.
//
// Gilbert takes task arguments from Cobra flags, but Cobra doesn't support
// dynamic flags.
//
// It means that Cobra application should be built from spec file before parsing command line
// and core flags responsible for working directory and cache location should be parsed before
// building Cobra command.
func ParsePreRunFlags(args []string) (*config.CoreConfig, error) {
	cfg, err := config.NewCoreConfig()
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return cfg, nil
	}

	var currentFlagName string
	for _, arg := range args {
		if !strings.HasPrefix(arg, flagPrefix) {
			// Parse flag value that comes after flag name
			if currentFlagName == "" {
				continue
			}

			if err := applyStringFlag(cfg, currentFlagName, strings.TrimSpace(arg)); err != nil {
				return nil, err
			}

			currentFlagName = ""
			continue
		}

		// Capture case of two sequential flags without values
		if currentFlagName != "" {
			return nil, newErrMissingFlagValue(currentFlagName)
		}

		flagName, value, err := parseFlag(arg)
		if err != nil {
			return nil, err
		}

		if isBoolFlag(flagName) {
			if err := applyBoolFlag(cfg, flagName, value); err != nil {
				return nil, err
			}
			continue
		}

		if !isKnownFlag(flagName) {
			// Skip unknown flags
			continue
		}

		// Expect value in next argument if flag doesn't contain value.
		if value == "" {
			currentFlagName = flagName
			continue
		}

		// Apply flag with a value.
		if err := applyStringFlag(cfg, flagName, value); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// parseFlag parses flag argument.
//
// Returns flag name without prefix and value it was specified in argument after '=' character.
// Returns error if flag is not valid.
func parseFlag(arg string) (name, value string, err error) {
	flagName := strings.TrimSpace(strings.TrimPrefix(arg, flagPrefix))
	if flagName == "" {
		return name, value, newErrEmptyFlag(flagName)
	}

	// Flag value might be specified in the same argument right after '=' character.
	chunks := strings.SplitN(flagName, "=", 2)
	if len(chunks) == 1 {
		return strings.TrimSpace(chunks[0]), "", nil
	}

	name = strings.TrimSpace(chunks[0])
	value = strings.TrimSpace(chunks[1])
	return name, value, nil
}

func isKnownFlag(flagName string) bool {
	switch flagName {
	case flagCwd, flagSpecFile, flagLogFormat:
		return true
	}

	return isBoolFlag(flagName)
}

func isBoolFlag(flagName string) bool {
	return flagName == flagVerbose
}

func applyBoolFlag(cfg *config.CoreConfig, flagName, value string) (err error) {
	boolValue := true
	if value != "" {
		boolValue, err = strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid %s%s flag value: %w", flagPrefix, flagName, err)
		}
	}

	switch flagName {
	case flagVerbose:
		cfg.Verbose = boolValue
	}

	return nil
}

func applyStringFlag(cfg *config.CoreConfig, flagName, value string) error {
	if value == "" {
		return newErrMissingFlagValue(flagName)
	}

	switch flagName {
	case flagCwd:
		cfg.WorkDir = value
	case flagSpecFile:
		cfg.SpecFile = value
	case flagLogFormat:
		cfg.LogFormat = value
	default:
		// Don't throw unknown flag error since they will be passed to Cobra.
	}

	return nil
}

func newErrMissingFlagValue(flagName string) error {
	return fmt.Errorf("flag needs an argument: %s%s", flagPrefix, flagName)
}

func newErrEmptyFlag(flagName string) error {
	return fmt.Errorf("empty flag name: %q", flagName)
}
