package manifest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMapLazyValueCastsScalars(t *testing.T) {
	cases := []struct {
		name string
		raw  any
		spec TypeSchema
		want any
	}{
		{
			name: "string parses through schema",
			raw:  "42",
			spec: TypeSchema{Type: ValueTypeInt},
			want: int64(42),
		},
		{
			name: "non-string becomes string",
			raw:  42,
			spec: TypeSchema{Type: ValueTypeString},
			want: "42",
		},
		{
			name: "bool casts",
			raw:  1,
			spec: TypeSchema{Type: ValueTypeBool},
			want: true,
		},
		{
			name: "duration casts",
			raw:  1500,
			spec: TypeSchema{Type: ValueTypeDuration},
			want: 1500 * time.Millisecond,
		},
		{
			name: "nil scalar becomes zero value",
			raw:  nil,
			spec: TypeSchema{Type: ValueTypeString},
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := mapLazyValue(context.Background(), literalLazyValue(tc.raw), valueMapOpts{
				spec:     tc.spec,
				allowNil: true,
			})
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestMapLazyValueCastsListItems(t *testing.T) {
	got, diags := mapLazyValue(context.Background(), literalLazyValue([]int{1, 2, 3}), valueMapOpts{
		spec: TypeSchema{
			Type:  ValueTypeList,
			Items: &TypeSchema{Type: ValueTypeFloat},
		},
		allowNil: true,
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, []any{float64(1), float64(2), float64(3)}, got)
}

func TestMapLazyValueCastsDictItems(t *testing.T) {
	got, diags := mapLazyValue(context.Background(), literalLazyValue(map[string]any{
		"enabled": 1,
		"empty":   nil,
	}), valueMapOpts{
		spec: TypeSchema{
			Type:  ValueTypeDict,
			Items: &TypeSchema{Type: ValueTypeBool},
		},
		allowNil: true,
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, map[string]any{
		"enabled": true,
		"empty":   false,
	}, got)
}

func TestMapLazyValueReportsNestedTypeErrors(t *testing.T) {
	_, diags := mapLazyValue(context.Background(), literalLazyValue([]any{"nope"}), valueMapOpts{
		spec: TypeSchema{
			Type:  ValueTypeList,
			Items: &TypeSchema{Type: ValueTypeInt},
		},
		allowNil: true,
	})
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Err.Error(), "invalid value at index 0")
}

func literalLazyValue(v any) *LazyValue {
	return &LazyValue{
		Location: &ReferenceLocation{},
		Value: AnySpec{
			LiteralSpec: &LiteralSpec{
				Value: v,
			},
		},
	}
}
