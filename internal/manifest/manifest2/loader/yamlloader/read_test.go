package yamlloader

import (
	"context"
	"testing"

	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	result, diags, err := yamltree.ReadSource(context.TODO(), jobFileSchema, yamltree.Source{
		FilePath: "testdata/test.yml",
	})

	require.NoError(t, err)
	t.Log(diags)
	t.Log(result)
}
