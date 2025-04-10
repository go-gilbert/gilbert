package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopCommandProcessor_EvalCommand(t *testing.T) {
	proc := NoopCommandProcessor{}
	v, err := proc.EvalCommand(t.Context(), "")
	require.Error(t, err)
	require.Empty(t, v)
}

func TestNoopValueResolver_Values(t *testing.T) {
	r := NoopValueResolver{}
	v, err := r.Values()
	require.Error(t, err)
	require.Empty(t, v)
}
