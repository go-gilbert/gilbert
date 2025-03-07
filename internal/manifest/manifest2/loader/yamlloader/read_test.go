package yamlloader

import (
	"context"
	"testing"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
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
		fileDir:           "testdata",
		filePath:          "testdata/test.yml",
		dst:               dst,
		builtinNamespaces: set.From(predefinedTestNamespaces),
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

func TestLoader_Load(t *testing.T) {
	defaultCfg := LoaderConfig{
		BuiltinNamespaces: predefinedTestNamespaces,
	}

	cases := []struct {
		name    string
		file    string
		cfg     LoaderConfig
		wantErr string
	}{
		{
			name: "cyclomatic import",
			file: "testdata/cyclomatic/a.yml",
			cfg:  defaultCfg,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			loader := NewLoader(c.cfg)
			result, err := loader.Load(t.Context(), c.file)
			if c.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.wantErr)
				return
			}

			require.NoError(t, err)
			dumpDiags(t, result.Diagnostics)
			t.Logf("%#v\n", result.File)
		})
	}
}

func dumpDiags(t *testing.T, diags parsetypes.Diagnostics) {
	if len(diags) == 0 {
		return
	}

	t.Log("==== Diagnostics: ====")
	for _, diag := range diags {
		t.Log(diag)
	}
	t.Log("======== END =========")
}
