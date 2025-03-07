package yamlloader

import (
	"context"
	"testing"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/hashicorp/go-set/v3"
	"github.com/stretchr/testify/require"
)

var predefinedTestNamespaces = []string{
	"go",
	"fs",
}

func TestRead(t *testing.T) {
	dst := &manifest2.JobFile{}
	ctx := newLoaderContext(context.TODO(), loaderContext{
		fileDir:         "testdata",
		filePath:        "testdata/test.yml",
		dst:             dst,
		knownNamespaces: set.From(predefinedTestNamespaces),
	})

	result, diags, err := yamltree.ReadSource(ctx, jobFileSchema, yamltree.Source{
		FilePath: "testdata/test.yml",
	}, yamltree.WithUnknownFieldAction(yamltree.UnknownFieldActionWarn))

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
