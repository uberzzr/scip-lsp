package docsync

import (
	"errors"

	"github.com/sergi/go-diff/diffmatchpatch"
	protocolmapper "github.com/uber/scip-lsp/src/ulsp/internal/protocol"
	"go.lsp.dev/protocol"
)

// PositionMapper is used to map corresponding LSP protocol positions between two versions of the same document.
type PositionMapper interface {
	// MapCurrentPositionToBase maps a position in the current document to the equivalent position in the base version
	MapCurrentPositionToBase(currentPosition protocol.Position) (basePosition protocol.Position, isNew bool, err error)

	// MapBasePositionToCurrent maps a position in the base document to the equivalent position in the current version
	MapBasePositionToCurrent(basePosition protocol.Position) (currentPosition protocol.Position, err error)
}

type documentPositionMapper struct {
	modified                bool
	baseTextOffsetMapper    *protocolmapper.TextOffsetMapper // Mapper to convert base text positions to offsets
	updatedTextOffsetMapper *protocolmapper.TextOffsetMapper // Mapper to convert updated text positions to offsets
	forwardDiffs            []diffmatchpatch.Diff            // Precomputed diffs from updated to base
	reverseDiffs            []diffmatchpatch.Diff            // Precomputed diffs from base to updated
}

// NewPositionMapper creates a new position mapper using precomputed diffs
func NewPositionMapper(baseText, updatedText string) PositionMapper {
	if baseText == updatedText {
		// Skip setup for documents with identical content
		return &documentPositionMapper{
			modified: false,
		}
	}

	dmp := diffmatchpatch.New()
	return &documentPositionMapper{
		modified: true,

		// Protocol mappers for fast conversion between text positions and offsets
		baseTextOffsetMapper:    protocolmapper.NewTextOffsetMapper([]byte(baseText)),
		updatedTextOffsetMapper: protocolmapper.NewTextOffsetMapper([]byte(updatedText)),

		// Precomputed diffs to shift offsets when text is modified
		forwardDiffs: dmp.DiffMain(baseText, updatedText, false),
		reverseDiffs: dmp.DiffMain(updatedText, baseText, false),
	}
}

// MapCurrentPositionToBase takes a position in the current document text and maps it to the equivalent position
// in the base text.
func (m *documentPositionMapper) MapCurrentPositionToBase(currentPosition protocol.Position) (basePosition protocol.Position, isNew bool, err error) {
	return m.mapPosition(currentPosition, true)
}

// MapBasePositionToCurrent takes a position in the saved document text and maps it to the equivalent position
// in the current version of the document.
func (m *documentPositionMapper) MapBasePositionToCurrent(position protocol.Position) (protocol.Position, error) {
	pos, _, err := m.mapPosition(position, false)
	return pos, err
}

// mapPosition is a helper method that maps a position from source to target
// It handles the common logic between MapCurrentPositionToBase and MapBasePositionToCurrent
func (m *documentPositionMapper) mapPosition(position protocol.Position, reverse bool) (protocol.Position, bool, error) {
	if !m.modified {
		return position, false, nil
	}

	if m.baseTextOffsetMapper == nil || m.updatedTextOffsetMapper == nil {
		return position, false, errors.New("position mapper not initialized")
	}

	// Determine which mapper to use based on the direction of mapping
	sourceMapper, targetMapper, diffs := m.baseTextOffsetMapper, m.updatedTextOffsetMapper, m.forwardDiffs
	if reverse {
		sourceMapper, targetMapper, diffs = m.updatedTextOffsetMapper, m.baseTextOffsetMapper, m.reverseDiffs
	}

	// Convert protocol position into its corresponding offset in the source text
	initialOffset, err := sourceMapper.PositionOffset(position)
	if err != nil {
		return position, false, err
	}

	// Apply diffs to get the new offset and whether the position was in a section that was deleted
	shiftedOffset, deleted := diffXIndex(diffs, initialOffset)

	// Convert the shifted offset back to a protocol position in the target text
	result, err := targetMapper.OffsetPosition(shiftedOffset)
	if err != nil {
		return position, deleted, err
	}

	return result, deleted, nil
}

// DiffXIndex returns the index of the position in the target text after applying the diffs.
// Lift and shift from diffmatchpatch library, with an added boolean return value indicating
// whether the position was deleted.
// Source: https://github.com/sergi/go-diff/blob/facec63e78161d6d31a9c552a679e2287e925949/diffmatchpatch/diff.go#L1090
func diffXIndex(diffs []diffmatchpatch.Diff, loc int) (int, bool) {
	chars1 := 0
	chars2 := 0
	lastChars1 := 0
	lastChars2 := 0
	lastDiff := diffmatchpatch.Diff{}
	for i := 0; i < len(diffs); i++ {
		aDiff := diffs[i]
		if aDiff.Type != diffmatchpatch.DiffInsert {
			// Equality or deletion.
			chars1 += len(aDiff.Text)
		}
		if aDiff.Type != diffmatchpatch.DiffDelete {
			// Equality or insertion.
			chars2 += len(aDiff.Text)
		}
		if chars1 > loc {
			// Overshot the location.
			lastDiff = aDiff
			break
		}
		lastChars1 = chars1
		lastChars2 = chars2
	}
	if lastDiff.Type == diffmatchpatch.DiffDelete {
		// The location was deleted.
		return lastChars2, true
	}
	// Add the remaining character length.
	return lastChars2 + (loc - lastChars1), false
}
