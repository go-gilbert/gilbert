package yamlloader

import (
	"context"
	"testing"

	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	ctx := newLoaderContext(context.TODO(), loaderContext{
		fileDir: "testdata",
	})

	result, diags, err := yamltree.ReadSource(ctx, jobFileSchema, yamltree.Source{
		FilePath: "testdata/test.yml",
	})

	require.NoError(t, err)
	if len(diags) != 0 {
		t.Log("== Diagnostics ==")
		for _, diag := range diags {
			t.Log(diag)
		}
		t.Log("== END ==")
	}

	t.Logf("%#v", result)
}
