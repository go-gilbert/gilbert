package uflag

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testBoolFlag  = "bool-var"
	testOtherFlag = "other"
)

type flagRecorder struct {
	boolFlag  bool
	otherFlag string
}

func (f *flagRecorder) IsBoolFlag(flagName string) bool {
	return flagName == testBoolFlag
}

func (f *flagRecorder) IsKnownFlag(flagName string) bool {
	switch flagName {
	case testBoolFlag, testOtherFlag:
		return true
	}

	return false
}

func (f *flagRecorder) ConsumeFlag(flagName, value string) {
	switch flagName {
	case testOtherFlag:
		f.otherFlag = value
	case testBoolFlag:
		if value == "" {
			f.boolFlag = true
			return
		}

		b, _ := strconv.ParseBool(value)
		f.boolFlag = b
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		expect flagRecorder
	}{
		{
			name:   "no known flags",
			args:   []string{"--foo", "--bar"},
			expect: flagRecorder{},
		},
		{
			name: "bool flag without value",
			args: []string{"--foo", "--bool-var", "--other=bar"},
			expect: flagRecorder{
				boolFlag:  true,
				otherFlag: "bar",
			},
		},
		{
			name: "bool flag with value",
			args: []string{"--foo", "--bool-var=true"},
			expect: flagRecorder{
				boolFlag: true,
			},
		},
		{
			name:   "string flag without value",
			args:   []string{"--foo", "--other"},
			expect: flagRecorder{},
		},
		{
			name: "unterminated flag",
			args: []string{"--other", "--bool-var", "true"},
			expect: flagRecorder{
				boolFlag: true,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got flagRecorder
			Parse(&got, c.args)
			require.Equal(t, c.expect, got)
		})
	}
}
