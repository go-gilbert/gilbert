package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty/gocty"
)

const maxHclVersion = 1

func Parse(data []byte, fileName string) (*Manifest, error) {
	f, err := hclsyntax.ParseConfig(data, fileName, hcl.InitialPos)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return manifestFromHcl(f)
}

func manifestFromHcl(f *hcl.File) (*Manifest, error) {
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("hcl file body is not %T (got %T)", body, f.Body)
	}

	version, err := extractVersion(body)
	if err != nil {
		return nil, err
	}

	if version > maxHclVersion {
		return nil, ErrUnsupportedVersion
	}

	return &Manifest{
		Version: version,
	}, nil
}

func extractVersion(body *hclsyntax.Body) (uint, error) {
	verAttr, ok := body.Attributes["version"]
	if !ok {
		return 0, ErrVersionMissing
	}

	ver, err := verAttr.Expr.Value(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version attribute: %w", err)
	}

	var version uint
	if err := gocty.FromCtyValue(ver, &version); err != nil {
		return 0, fmt.Errorf("invalid version attribute: %w", err)
	}

	return version, nil
}
