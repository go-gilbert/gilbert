package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const errImportMsg = "cannot read file '%s' imported by '%s', %s"

type leafsGroup map[int][]*importNode

type importNode struct {
	parent   *importNode
	manifest *Manifest
	path     string
	level    int
	imports  []*importNode
}

func (i *importNode) mergeChildren() {
	for _, child := range i.imports {
		i.manifest.includeParent(child.manifest)
	}
}

type importTree struct {
	root  *importNode
	leafs []*importNode
	depth int
}

func newImportTree(m *Manifest) *importTree {
	return &importTree{
		root: importNodeFromManifest(m),
	}
}

func importNodeFromManifest(m *Manifest) *importNode {
	return &importNode{
		manifest: m,
		path:     filepath.Dir(m.location),
		imports:  make([]*importNode, 0, len(m.Imports)),
	}
}

func (t *importTree) result() Manifest {
	r := *t.root.manifest
	t.root = nil
	t.leafs = nil
	return r
}

// shrinkWrapImports merges all graph nodes with their parents until it gets to the root
func (t *importTree) shrinkWrapImports(leafs leafsGroup) {
	// Iterate over each leafs layer from the end
	for lvl := t.depth; lvl > 0; lvl-- {
		// Merge each layer with parent
		for _, leaf := range leafs[lvl] {
			if leaf.parent != nil {
				leaf.parent.mergeChildren()
			}
		}
	}
}

// resolveImports resolves imports and applies them on the root manifest
func (t *importTree) resolveImports() error {
	if err := t.buildImportsTree(t.root); err != nil {
		return err
	}

	// Sort leafs by level desc (optional)
	sort.Slice(t.leafs, func(i, j int) bool {
		return t.leafs[i].level > t.leafs[j].level
	})

	// group leafs by level
	leafs := make(leafsGroup)
	for _, leaf := range t.leafs {
		if leafs[leaf.level] == nil {
			// init slice if empty
			leafs[leaf.level] = make([]*importNode, 0)
		}

		leafs[leaf.level] = append(leafs[leaf.level], leaf)
	}

	t.shrinkWrapImports(leafs)
	return nil
}

// buildImportsTree resolves all imported files by the main root into import graph
func (t *importTree) buildImportsTree(n *importNode) error {
	nextLevel := n.level + 1

	// Increase global depth state
	if t.depth < nextLevel {
		t.depth = nextLevel
	}

	for _, importFile := range n.manifest.Imports {
		// Join path since import path based on parent file location
		filePath := filepath.Join(n.path, importFile)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf(errImportMsg, importFile, n.manifest.location, err)
		}

		yml, err := UnmarshalManifest(data)
		if err != nil {
			return fmt.Errorf(errImportMsg, importFile, n.manifest.location, err)
		}

		// TODO: check version and convert template expressions in template.

		yml.location = filePath
		node := importNodeFromManifest(yml)
		node.parent = n
		node.level = nextLevel

		// Process child imports if there is at least one
		if len(yml.Imports) > 0 {
			if err = t.buildImportsTree(node); err != nil {
				return fmt.Errorf("cannot resolve imports of '%s': %w", yml.location, err)
			}
		} else {
			// If node has no children, mark as leaf
			t.leafs = append(t.leafs, node)
		}

		n.imports = append(n.imports, node)
	}

	return nil
}
