package expr

import (
	"errors"
	"testing"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/stretchr/testify/require"
)

func TestTokenError_Unwrap(t *testing.T) {
	err := errors.New("test")
	tokErr := &TokenError{
		Err: err,
	}

	gotErr := errors.Unwrap(tokErr)
	require.Equal(t, err, gotErr)
}

func TestTokenError_Error(t *testing.T) {
	err := &TokenError{
		Position: parsetypes.NewRange(
			parsetypes.NewPosition(1, 10),
			parsetypes.NewPosition(2, 20),
		),
		Err: errors.New("test"),
	}

	wantMsg := "test (at 1:10 - 2:20)"
	gotMsg := err.Error()
	require.Equal(t, wantMsg, gotMsg)
}
