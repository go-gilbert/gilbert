package spec

import "github.com/hashicorp/hcl/v2/hclsyntax"

type Body struct {
	Vars    *Vars
	Params  Params
	Tasks   Tasks
	Unknown hclsyntax.Blocks
}

type Header struct {
	Version uint
	Imports []string
}

type Spec struct {
	Header
	Body
}
