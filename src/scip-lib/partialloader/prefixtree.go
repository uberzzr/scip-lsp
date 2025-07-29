package partialloader

import (
	"github.com/uber/scip-lsp/src/scip-lib/mapper"
	"github.com/uber/scip-lsp/src/scip-lib/model"
)

// SymbolPrefixTreeNode is a walkable tree that maps out all external symbols in loaded indices
// eg a symbol with path com.uber.devexp.Converter# will map down the tree as com->uber->devexp->Converter
// This tree can be used to lookup symbols by their path (and thus provide hover info), or to walk the tree for completion results.
// Leaf nodes are identified by the presence of the Info field.
type SymbolPrefixTreeNode struct {
	// Reference to the parent to make computing the identifier easier
	Parent *SymbolPrefixTreeNode
	// Preamble defines the part before the descriptor for this tree
	Preamble string
	// SymbolVersions is a map of symbol versions to their symbol information
	SymbolVersions map[string]*struct {
		// Info is the symbol information
		// (if nil, no symbol by this name exists)
		Info *model.SymbolInformation
		// DocumentPath is the path to the document that this node belongs to
		DocumentPath string
	}
	// Children defines any symbols further down the tree
	Children map[model.Descriptor]*SymbolPrefixTreeNode
	// Revision is the revision number when this node was last updated
	Revision int64
}

// SymbolPrefixTree is a tree of SymbolPrefixTreeNodes
type SymbolPrefixTree struct {
	SymbolPrefixTreeNode
}

// NewSymbolPrefixTree creates a new SymbolPrefixTree
func NewSymbolPrefixTree() *SymbolPrefixTree {
	return &SymbolPrefixTree{
		SymbolPrefixTreeNode: SymbolPrefixTreeNode{
			Children: make(map[model.Descriptor]*SymbolPrefixTreeNode),
		},
	}
}

// AddSymbol adds a symbol to the tree
func (t *SymbolPrefixTreeNode) AddSymbol(relativeDocPath string, info *model.SymbolInformation, revision int64) (node *SymbolPrefixTreeNode, isNew bool) {
	// Parse the symbol, then recursively map down the tree
	sy, err := model.ParseScipSymbol(info.Symbol)
	if err != nil {
		return nil, false
	}
	version := sy.Package.Version

	node = t
	cnt := len(sy.Descriptors)
	for i, descriptor := range mapper.ScipDescriptorsToModelDescriptors(sy.Descriptors) {
		if i == cnt-1 {
			node, isNew = node.addChild(relativeDocPath, sy.Scheme+" "+sy.Package.ID(), descriptor, info, revision, version)
		} else {
			node, isNew = node.addChild("", "", descriptor, nil, revision, version)
		}
	}

	return node, isNew
}

// GetNode returns the node for the given symbol
func (t *SymbolPrefixTreeNode) GetNode(symbol string) *SymbolPrefixTreeNode {
	sy, err := model.ParseScipSymbol(symbol)
	if err != nil {
		return nil
	}

	node := t
	for _, descriptor := range mapper.ScipDescriptorsToModelDescriptors(sy.Descriptors) {
		node = node.Children[descriptor]
		if node == nil {
			return nil
		}
	}
	return node
}

// Merge merges another tree into this one
func (t *SymbolPrefixTreeNode) Merge(other *SymbolPrefixTreeNode) {
	// Top levels should be the same, add any new children, merge already existing children
	for desc, child := range other.Children {
		if t.Children[desc] == nil {
			// Just take the child and update its parent reference
			child.Parent = t
			t.Children[desc] = child
		} else {
			// Update existing node with newer information if available
			existing := t.Children[desc]

			// Merge symbol versions
			if child.SymbolVersions != nil {
				if existing.SymbolVersions == nil {
					existing.SymbolVersions = make(map[string]*struct {
						Info         *model.SymbolInformation
						DocumentPath string
					})
				}

				// Merge the maps by adding or updating each version
				for version, versionInfo := range child.SymbolVersions {
					existing.SymbolVersions[version] = versionInfo
				}
			}

			if child.Revision > existing.Revision {
				existing.Revision = child.Revision
			}

			// Recursively merge children
			existing.Merge(child)
		}
	}
}

func (t *SymbolPrefixTreeNode) addChild(relativeDocPath string, preamble string, descriptor model.Descriptor, info *model.SymbolInformation, revision int64, version string) (*SymbolPrefixTreeNode, bool) {
	isNew := false
	if t.Children[descriptor] == nil {
		t.Children[descriptor] = &SymbolPrefixTreeNode{
			Parent:   t,
			Preamble: preamble,
			Children: make(map[model.Descriptor]*SymbolPrefixTreeNode),
			SymbolVersions: make(map[string]*struct {
				Info         *model.SymbolInformation
				DocumentPath string
			}),
			Revision: revision,
		}
		isNew = true
	} else {
		// Update existing node
		existing := t.Children[descriptor]
		existing.Revision = revision
	}

	if info != nil {
		t.Children[descriptor].SymbolVersions[version] = &struct {
			Info         *model.SymbolInformation
			DocumentPath string
		}{
			Info:         info,
			DocumentPath: relativeDocPath,
		}
	}

	return t.Children[descriptor], isNew
}

// PruneNodes removes nodes that are older than the given revision
func (t *SymbolPrefixTreeNode) PruneNodes(docPath string, revision int64) {
	for name, child := range t.Children {
		shouldDelete := false
		if child.Revision < revision {
			// Check if any symbol version contains the given document path
			for _, versionInfo := range child.SymbolVersions {
				if versionInfo.DocumentPath == docPath {
					shouldDelete = true
					break
				}
			}
		}

		if shouldDelete {
			delete(t.Children, name)
		} else {
			child.PruneNodes(docPath, revision)
		}
	}
}
