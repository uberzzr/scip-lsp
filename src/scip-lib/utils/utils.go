package utils

import (
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/uber/scip-lsp/src/scip-lib/model"
	"go.lsp.dev/protocol"
)

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
