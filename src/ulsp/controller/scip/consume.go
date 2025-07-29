package scip

import (
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/uber/scip-lsp/src/scip-lib/model"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// GetSymbolDataForFile returns the document symbols for a given file
func (r *registryImpl) GetDocumentSymbolForFile(uri uri.URI) (*[]*SymbolData, error) {
	symbols := make([]*SymbolData, 0)
	doc, ok := r.Documents[uri]
	if !ok {
		return &symbols, nil
	}

	occs := doc.Document.Occurrences

	for _, occ := range occs {
		if occ.SymbolRoles&int32(scip.SymbolRole_Definition) > 0 {
			var symbolData *SymbolData
			if scip.IsLocalSymbol(occ.Symbol) {
				symbolData = doc.Locals[occ.Symbol]
			} else {
				symbolData, _ = r.GetSymbolDataForSymbol(occ.Symbol, doc.Package)
			}

			if symbolData == nil {
				continue
			}
			symbols = append(symbols, symbolData)
		}
	}
	return &symbols, nil
}

// GetSymbolForPosition returns the occurrence and symbol data for a given position
func (r *registryImpl) GetSymbolForPosition(uri uri.URI, pos protocol.Position) (*model.Occurrence, *SymbolData, error) {
	doc := r.Documents[uri]
	if doc == nil {
		return nil, nil, nil
	}

	occ := GetOccurrenceForPosition(doc.Document.Occurrences, pos)
	if occ == nil {
		return nil, nil, nil
	}

	if scip.IsLocalSymbol(occ.Symbol) && len(doc.Locals) > 0 {
		// If locals are in the doc, they won't be in the package
		return occ, doc.Locals[occ.Symbol], nil
	}

	sd, err := r.GetSymbolDataForSymbol(occ.Symbol, doc.Package)
	return occ, sd, err
}

// GetSymbolDataForSymbol returns the symbol data for a given symbol
func (r *registryImpl) GetSymbolDataForSymbol(symbol string, localPkg *PackageMeta) (*SymbolData, error) {
	if scip.IsLocalSymbol(symbol) {
		localPkg.mu.Lock()
		defer localPkg.mu.Unlock()
		return localPkg.SymbolData[symbol], nil
	}

	sy, err := model.ParseScipSymbol(symbol)
	if err != nil {
		return nil, err
	}

	pkg := r.Packages[PackageID(sy.Package.Name)]
	if pkg == nil {
		return nil, nil
	}
	pkg.mu.Lock()
	defer pkg.mu.Unlock()

	return pkg.SymbolData[symbol], nil
}

// GetFileInfo returns the FileInfo for a given URI
func (r *registryImpl) GetFileInfo(fileURI uri.URI) *FileInfo {
	return r.Documents[fileURI]
}

// GetPackageInfo returns the Package for a given package ID
func (r *registryImpl) GetPackageInfo(pkgID PackageID) *PackageMeta {
	found := make([]*PackageMeta, 0)
	for id, pkg := range r.Packages {
		if id == pkgID {
			found = append(found, pkg)
		}
	}
	return r.Packages[pkgID]
}

// GetOccurrencesForSymbol returns the occurrences for a given symbol and role
// If role is -1, it will return all occurrences for the symbol
func GetOccurrencesForSymbol(occurrences []*model.Occurrence, symbol string, role scip.SymbolRole) []*model.Occurrence {
	matchAll := true
	if role != -1 {
		matchAll = false
	}
	found := make([]*model.Occurrence, 0)
	for _, occ := range occurrences {
		if occ.Symbol == symbol && (matchAll || occ.SymbolRoles&int32(role) > 0) {
			found = append(found, occ)
		}
	}
	return found
}

// GetLocalSymbolInformation returns the symbol information for a given symbol
func GetLocalSymbolInformation(symbols []*model.SymbolInformation, symbol string) *model.SymbolInformation {
	for _, sym := range symbols {
		if sym.Symbol == symbol {
			return sym
		}
	}
	return nil
}

// GetOccurrenceForPosition returns the occurrence for a given position
func GetOccurrenceForPosition(occurrences []*model.Occurrence, pos protocol.Position) *model.Occurrence {
	// Since occurrences are sorted by their position in the document
	// we can use binary search to speed up the lookup
	low := 0
	high := len(occurrences) - 1

	for low <= high {
		mid := (low + high) / 2
		occ := occurrences[mid]
		if IsMatchingPosition(occ, pos) {
			return occ
		}
		if IsRangeBefore(occ.Range, pos) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return nil
}

// IsRangeBefore returns true if the range is before the position
func IsRangeBefore(r []int32, pos protocol.Position) bool {
	endLine := int32(0)
	endChar := int32(0)
	if len(r) == 3 {
		endLine = r[0]
		endChar = r[2]
	}
	if len(r) == 4 {
		endLine = r[2]
		endChar = r[3]
	}

	return endLine < int32(pos.Line) ||
		(endLine == int32(pos.Line) && endChar < int32(pos.Character))
}

// IsMatchingPosition returns true if the occurrence matches the position
func IsMatchingPosition(occ *model.Occurrence, pos protocol.Position) bool {
	rng, err := scip.NewRange(occ.Range)
	if err != nil {
		return false
	}
	if rng.IsSingleLine() {
		return rng.Start.Line == int32(pos.Line) &&
			rng.Start.Character <= int32(pos.Character) &&
			rng.End.Character >= int32(pos.Character)
	}

	// For multiline ranges we need a slightly more complex check:
	// 1. If the position is between the start and end line, we don't need to do a character check
	// 2. If the position is on the start line, we need to check if it's beyond the start of the range
	// 3. If the position is on the end line, we need to check if it's before the end of the range
	return (rng.Start.Line < int32(pos.Line) && rng.End.Line > int32(pos.Line)) ||
		(rng.Start.Line == int32(pos.Line) && rng.Start.Character <= int32(pos.Character)) ||
		(rng.End.Line == int32(pos.Line) && rng.End.Character >= int32(pos.Character))
}
