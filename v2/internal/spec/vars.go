package spec

import "github.com/hashicorp/hcl/v2"

type Vars struct {
	Range      hcl.Range
	Attributes hcl.Attributes
}

func ParseVars(b *hcl.Block) (*Vars, hcl.Diagnostics) {
	attrs, err := b.Body.JustAttributes()
	if err != nil {
		return nil, err
	}

	return &Vars{
		Range:      b.DefRange,
		Attributes: attrs,
	}, nil
}
