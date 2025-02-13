package yamlloader

import (
	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
)

var taskSchema = Struct[manifest2.Task](
// Field[manifest2.Task, string]
)
