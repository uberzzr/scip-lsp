package mapper

import (
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/uber/scip-lsp/src/scip-lib/model"
	"go.lsp.dev/protocol"
)

const _localDiagnosticsPrefix = "local diagnostic_"

// ScipPackageToModelScipPackage converts a SCIP package to a model SCIP package
func ScipPackageToModelScipPackage(pkg *scip.Package) *model.ScipPackage {
	if pkg == nil {
		return nil
	}
	return &model.ScipPackage{
		Manager: pkg.Manager,
		Name:    pkg.Name,
		Version: pkg.Version,
	}
}

// ScipDocumentToModelDocument converts a SCIP document to a model document
func ScipDocumentToModelDocument(doc *scip.Document) *model.Document {
	if doc == nil {
		return nil
	}
	occurrences := make([]*model.Occurrence, 0, len(doc.Occurrences))
	symbols := make([]*model.SymbolInformation, 0, len(doc.Symbols))
	allDiagnostics := make([]*protocol.Diagnostic, 0)
	symbolMap := make(map[string]*model.SymbolInformation)

	for _, occ := range doc.Occurrences {
		// Skip occurrences with empty symbols

		occurrence, diagnostics := ScipOccurrenceToModelOccurrenceWithDiagnostics(occ)
		allDiagnostics = append(allDiagnostics, diagnostics...)
		if occ.Symbol == "" || strings.HasPrefix(occ.Symbol, _localDiagnosticsPrefix) {
			continue
		}
		occurrences = append(occurrences, occurrence)
	}

	for _, sym := range doc.Symbols {
		info := ScipSymbolInformationToModelSymbolInformation(sym)
		symbols = append(symbols, info)
		symbolMap[info.Symbol] = info
	}

	return &model.Document{
		Occurrences:  occurrences,
		Language:     doc.Language,
		RelativePath: doc.RelativePath,
		Symbols:      symbols,
		SymbolMap:    symbolMap,
		Text:         doc.Text,
		Diagnostics:  allDiagnostics,
	}
}

// ScipOccurrenceToModelOccurrence converts a SCIP occurrence to a model occurrence
func ScipOccurrenceToModelOccurrence(occ *scip.Occurrence) *model.Occurrence {
	if occ == nil {
		return nil
	}

	return &model.Occurrence{
		Range:                 occ.Range,
		Symbol:                occ.Symbol,
		SymbolRoles:           occ.SymbolRoles,
		OverrideDocumentation: occ.OverrideDocumentation,
		SyntaxKind:            occ.SyntaxKind,
		EnclosingRange:        occ.EnclosingRange,
	}
}

// ScipOccurrenceToModelOccurrenceWithDiagnostics converts a SCIP occurrence to a model occurrence and its diagnostics
func ScipOccurrenceToModelOccurrenceWithDiagnostics(occ *scip.Occurrence) (*model.Occurrence, []*protocol.Diagnostic) {
	if occ == nil {
		return nil, nil
	}
	diagnostics := make([]*protocol.Diagnostic, 0, len(occ.Diagnostics))

	diagnostics = append(diagnostics, ScipDiagnosticToModelDiagnostic(occ)...)

	// Diagnostic might include range of the occurrences, so we need to separate them
	// Also warning might be related to the missing symbol (missing occurrence)
	return &model.Occurrence{
		Range:                 occ.Range,
		Symbol:                occ.Symbol,
		SymbolRoles:           occ.SymbolRoles,
		OverrideDocumentation: occ.OverrideDocumentation,
		SyntaxKind:            occ.SyntaxKind,
		EnclosingRange:        occ.EnclosingRange,
	}, diagnostics
}

// ScipDiagnosticToModelDiagnostic converts a SCIP diagnostic to a protocol diagnostic
func ScipDiagnosticToModelDiagnostic(occ *scip.Occurrence) []*protocol.Diagnostic {
	if occ == nil {
		return nil
	}

	var positions protocol.Range
	if len(occ.Range) >= 4 {
		positions = protocol.Range{
			Start: protocol.Position{
				Line:      uint32(occ.Range[0]),
				Character: uint32(occ.Range[1]),
			},
			End: protocol.Position{
				Line:      uint32(occ.Range[2]),
				Character: uint32(occ.Range[3]),
			},
		}
	} else if len(occ.Range) >= 3 {
		positions = protocol.Range{
			Start: protocol.Position{
				Line:      uint32(occ.Range[0]),
				Character: uint32(occ.Range[1]),
			},
			End: protocol.Position{
				Line:      uint32(occ.Range[0]),
				Character: uint32(occ.Range[2]),
			},
		}
	}
	res := make([]*protocol.Diagnostic, 0, len(occ.Diagnostics))
	for _, diag := range occ.Diagnostics {
		res = append(res, &protocol.Diagnostic{
			Range:    positions,
			Severity: protocol.DiagnosticSeverity(diag.Severity),
			Code:     diag.Code,
			Message:  diag.Message,
			Source:   diag.Source,
			Tags:     convertDiagnosticTags(diag.Tags),
		})
	}

	return res
}

// convertDiagnosticTags converts SCIP diagnostic tags to protocol diagnostic tags
func convertDiagnosticTags(tags []scip.DiagnosticTag) []protocol.DiagnosticTag {
	if len(tags) == 0 {
		return nil
	}
	protocolTags := make([]protocol.DiagnosticTag, len(tags))
	for i, tag := range tags {
		protocolTags[i] = protocol.DiagnosticTag(tag)
	}
	return protocolTags
}

// ScipSymbolInformationToModelSymbolInformation converts a SCIP symbol information to a model symbol information
func ScipSymbolInformationToModelSymbolInformation(sym *scip.SymbolInformation) *model.SymbolInformation {
	if sym == nil {
		return nil
	}
	relationships := make([]*model.Relationship, 0, len(sym.Relationships))

	for _, rel := range sym.Relationships {
		relationships = append(relationships, ScipRelationshipToModelRelationship(rel))
	}

	return &model.SymbolInformation{
		Symbol:                 sym.Symbol,
		Documentation:          sym.Documentation,
		Relationships:          relationships,
		Kind:                   sym.Kind,
		DisplayName:            sym.DisplayName,
		SignatureDocumentation: ScipDocumentToModelDocument(sym.SignatureDocumentation),
		EnclosingSymbol:        sym.EnclosingSymbol,
	}
}

// ScipRelationshipToModelRelationship converts a SCIP relationship to a model relationship
func ScipRelationshipToModelRelationship(rel *scip.Relationship) *model.Relationship {
	if rel == nil {
		return nil
	}
	return &model.Relationship{
		Symbol:           rel.Symbol,
		IsReference:      rel.IsReference,
		IsImplementation: rel.IsImplementation,
		IsTypeDefinition: rel.IsTypeDefinition,
		IsDefinition:     rel.IsDefinition,
	}
}

// ScipDescriptorToModelDescriptor converts a SCIP descriptor to a model descriptor
func ScipDescriptorToModelDescriptor(desc *scip.Descriptor) model.Descriptor {
	if desc == nil {
		return model.Descriptor{}
	}
	return model.Descriptor{
		Name:          desc.Name,
		Suffix:        desc.Suffix,
		Disambiguator: desc.Disambiguator,
	}
}

// ScipDescriptorsToModelDescriptors converts a slice of SCIP descriptors to a slice of model descriptors
func ScipDescriptorsToModelDescriptors(descriptors []*scip.Descriptor) []model.Descriptor {
	if descriptors == nil {
		return []model.Descriptor{}
	}
	modelDescriptors := make([]model.Descriptor, len(descriptors))
	for i, descriptor := range descriptors {
		modelDescriptors[i] = ScipDescriptorToModelDescriptor(descriptor)
	}
	return modelDescriptors
}
