package cmd

import (
	"github.com/go-gilbert/gilbert/v2/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePreRunFlags(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err, "os.Getwd error")

	cases := map[string]struct {
		expect      *config.CoreConfig
		args        []string
		expectError string
	}{
		"should apply flags to config": {
			args: []string{"--cwd", "/foo/bar", "--spec-file=foo.hcl", "--log-format", "json", "--verbose"},
			expect: &config.CoreConfig{
				WorkDir:   "/foo/bar",
				SpecFile:  "foo.hcl",
				LogFormat: "json",
				Verbose:   true,
			},
		},
		"returns default config when no flags provided": {
			args: nil,
			expect: &config.CoreConfig{
				WorkDir:   wd,
				SpecFile:  config.DefaultSpecFile,
				LogFormat: config.DefaultLogFormat,
				Verbose:   false,
			},
		},
		"should skip unknown flags": {
			args: []string{"--foo", "bar", "--bad-flag2=", "--invalid-flag", "--cwd", "/tmp", "--verbose=true", "baz", "xxx"},
			expect: &config.CoreConfig{
				WorkDir:   "/tmp",
				SpecFile:  config.DefaultSpecFile,
				LogFormat: config.DefaultLogFormat,
				Verbose:   true,
			},
		},
		"should validate missing flag value argument": {
			args:        []string{"--foo", "--spec-file", "--cwd"},
			expectError: newErrMissingFlagValue("spec-file").Error(),
		},
		"should validate missing flag value in same option": {
			args:        []string{"--foo", "--cwd=", "--spec-file", "bar"},
			expectError: newErrMissingFlagValue("cwd").Error(),
		},
		"should validate boolean flag value": {
			args:        []string{"abcd", "--verbose=qwert"},
			expectError: `invalid --verbose flag value: strconv.ParseBool: parsing "qwert": invalid syntax`,
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			got, err := ParsePreRunFlags(c.args)
			if c.expectError != "" {
				require.EqualError(t, err, c.expectError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, c.expect, got)
		})
	}
}
