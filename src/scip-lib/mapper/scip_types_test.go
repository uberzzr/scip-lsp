package mapper

import (
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/scip-lib/model"
	"go.lsp.dev/protocol"
)

func TestScipPackageToModelScipPackage(t *testing.T) {
	pkg := &scip.Package{
		Manager: "scip_manager",
		Name:    "scip_package",
		Version: "1.0.0",
	}

	expected := &model.ScipPackage{
		Manager: "scip_manager",
		Name:    "scip_package",
		Version: "1.0.0",
	}

	result := ScipPackageToModelScipPackage(pkg)

	assert.Equal(t, result, expected)
}

func TestScipDocumentToModelDocument(t *testing.T) {
	doc := &scip.Document{
		Occurrences: []*scip.Occurrence{
			{
				Range:       []int32{1, 0, 5},
				Symbol:      "symbol 1234 local1",
				SymbolRoles: int32(scip.SymbolRole_Definition),
				SyntaxKind:  scip.SyntaxKind_Identifier,
				Diagnostics: []*scip.Diagnostic{
					{
						Severity: scip.Severity_Error,
						Code:     "code",
						Message:  "message",
						Source:   "source",
						Tags:     []scip.DiagnosticTag{scip.DiagnosticTag_Deprecated},
					},
				},
				EnclosingRange: []int32{0, 0, 10},
			},
		},
		Language:     "go",
		RelativePath: "path/to/file.go",
		Symbols: []*scip.SymbolInformation{
			{
				Symbol:        "symbol 1234 local1",
				Documentation: []string{"documentation"},
				Relationships: []*scip.Relationship{
					{
						Symbol:           "symbol 12 local2",
						IsReference:      true,
						IsImplementation: false,
						IsTypeDefinition: false,
						IsDefinition:     true,
					},
				},
				Kind:                   scip.SymbolInformation_Kind(scip.SyntaxKind_Identifier),
				DisplayName:            "display",
				SignatureDocumentation: nil,
				EnclosingSymbol:        "enclosing",
			},
		},
		Text: "some text",
	}

	expected := &model.Document{
		Occurrences: []*model.Occurrence{
			{
				Range:          []int32{1, 0, 5},
				Symbol:         "symbol 1234 local1",
				SymbolRoles:    int32(scip.SymbolRole_Definition),
				SyntaxKind:     scip.SyntaxKind_Identifier,
				EnclosingRange: []int32{0, 0, 10},
			},
		},
		Language:     "go",
		RelativePath: "path/to/file.go",
		Symbols: []*model.SymbolInformation{
			{
				Symbol:        "symbol 1234 local1",
				Documentation: []string{"documentation"},
				Relationships: []*model.Relationship{
					{
						Symbol:           "symbol 12 local2",
						IsReference:      true,
						IsImplementation: false,
						IsTypeDefinition: false,
						IsDefinition:     true,
					},
				},
				Kind:                   scip.SymbolInformation_Kind(scip.SyntaxKind_Identifier),
				DisplayName:            "display",
				SignatureDocumentation: nil,
				EnclosingSymbol:        "enclosing",
			},
		},
		Text: "some text",
		Diagnostics: []*protocol.Diagnostic{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 0,
					},
					End: protocol.Position{
						Line:      1,
						Character: 5,
					},
				},
				Severity: protocol.DiagnosticSeverityError,
				Code:     "code",
				Source:   "source",
				Message:  "message",
				Tags:     []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
			},
		},
		SymbolMap: map[string]*model.SymbolInformation{
			"symbol 1234 local1": {
				Symbol:        "symbol 1234 local1",
				Documentation: []string{"documentation"},
				Relationships: []*model.Relationship{
					{
						Symbol:           "symbol 12 local2",
						IsReference:      true,
						IsImplementation: false,
						IsTypeDefinition: false,
						IsDefinition:     true,
					},
				},
				Kind:                   scip.SymbolInformation_Kind(scip.SyntaxKind_Identifier),
				DisplayName:            "display",
				SignatureDocumentation: nil,
				EnclosingSymbol:        "enclosing",
			},
		},
	}

	result := ScipDocumentToModelDocument(doc)
	assert.Equal(t, expected, result)
}

func TestScipDocumentWithMultilineOccToModelDocument(t *testing.T) {
	doc := &scip.Document{
		Occurrences: []*scip.Occurrence{
			{
				Range:       []int32{1, 0, 4, 5},
				Symbol:      "symbol 1234 local1",
				SymbolRoles: int32(scip.SymbolRole_Definition),
				SyntaxKind:  scip.SyntaxKind_Identifier,
				Diagnostics: []*scip.Diagnostic{
					{
						Severity: scip.Severity_Error,
						Code:     "code",
						Message:  "message",
						Source:   "source",
						Tags:     []scip.DiagnosticTag{scip.DiagnosticTag_Deprecated},
					},
				},
				EnclosingRange: []int32{0, 0, 10},
			},
		},
		Language:     "go",
		RelativePath: "path/to/file.go",
		Symbols: []*scip.SymbolInformation{
			{
				Symbol:        "symbol 1234 local1",
				Documentation: []string{"documentation"},
				Relationships: []*scip.Relationship{
					{
						Symbol:           "symbol 12 local2",
						IsReference:      true,
						IsImplementation: false,
						IsTypeDefinition: false,
						IsDefinition:     true,
					},
				},
				Kind:                   scip.SymbolInformation_Kind(scip.SyntaxKind_Identifier),
				DisplayName:            "display",
				SignatureDocumentation: nil,
				EnclosingSymbol:        "enclosing",
			},
		},
		Text: "some text",
	}

	expected := &model.Document{
		Occurrences: []*model.Occurrence{
			{
				Range:          []int32{1, 0, 4, 5},
				Symbol:         "symbol 1234 local1",
				SymbolRoles:    int32(scip.SymbolRole_Definition),
				SyntaxKind:     scip.SyntaxKind_Identifier,
				EnclosingRange: []int32{0, 0, 10},
			},
		},
		Language:     "go",
		RelativePath: "path/to/file.go",
		Symbols: []*model.SymbolInformation{
			{
				Symbol:        "symbol 1234 local1",
				Documentation: []string{"documentation"},
				Relationships: []*model.Relationship{
					{
						Symbol:           "symbol 12 local2",
						IsReference:      true,
						IsImplementation: false,
						IsTypeDefinition: false,
						IsDefinition:     true,
					},
				},
				Kind:                   scip.SymbolInformation_Kind(scip.SyntaxKind_Identifier),
				DisplayName:            "display",
				SignatureDocumentation: nil,
				EnclosingSymbol:        "enclosing",
			},
		},
		Text: "some text",
		Diagnostics: []*protocol.Diagnostic{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 0,
					},
					End: protocol.Position{
						Line:      4,
						Character: 5,
					},
				},
				Severity: protocol.DiagnosticSeverityError,
				Code:     "code",
				Source:   "source",
				Message:  "message",
				Tags:     []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
			},
		},
		SymbolMap: map[string]*model.SymbolInformation{
			"symbol 1234 local1": {
				Symbol:        "symbol 1234 local1",
				Documentation: []string{"documentation"},
				Relationships: []*model.Relationship{
					{
						Symbol:           "symbol 12 local2",
						IsReference:      true,
						IsImplementation: false,
						IsTypeDefinition: false,
						IsDefinition:     true,
					},
				},
				Kind:                   scip.SymbolInformation_Kind(scip.SyntaxKind_Identifier),
				DisplayName:            "display",
				SignatureDocumentation: nil,
				EnclosingSymbol:        "enclosing",
			},
		},
	}

	result := ScipDocumentToModelDocument(doc)
	assert.Equal(t, expected, result)
}

func TestScipOccurrenceToModelOccurrence(t *testing.T) {
	tests := []struct {
		name string
		occ  *scip.Occurrence
		want *model.Occurrence
	}{
		{
			name: "nil occurrence",
			occ:  nil,
			want: nil,
		},
		{
			name: "full occurrence",
			occ: &scip.Occurrence{
				Range:                 []int32{1, 0, 5},
				Symbol:                "test.symbol",
				SymbolRoles:           int32(scip.SymbolRole_Definition),
				OverrideDocumentation: []string{"override doc"},
				SyntaxKind:            scip.SyntaxKind_Identifier,
				EnclosingRange:        []int32{0, 0, 10},
			},
			want: &model.Occurrence{
				Range:                 []int32{1, 0, 5},
				Symbol:                "test.symbol",
				SymbolRoles:           int32(scip.SymbolRole_Definition),
				OverrideDocumentation: []string{"override doc"},
				SyntaxKind:            scip.SyntaxKind_Identifier,
				EnclosingRange:        []int32{0, 0, 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScipOccurrenceToModelOccurrence(tt.occ)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNil(t *testing.T) {
	assert.Nil(t, ScipPackageToModelScipPackage(nil))
	assert.Nil(t, ScipDocumentToModelDocument(nil))
	assert.Nil(t, ScipDiagnosticToModelDiagnostic(nil))
	assert.Nil(t, ScipSymbolInformationToModelSymbolInformation(nil))
	assert.Nil(t, ScipRelationshipToModelRelationship(nil))
	assert.Equal(t, model.Descriptor{}, ScipDescriptorToModelDescriptor(nil))
	occ, diag := ScipOccurrenceToModelOccurrenceWithDiagnostics(nil)
	assert.Nil(t, occ)
	assert.Nil(t, diag)
}

func TestScipDescriptorToModelDescriptor(t *testing.T) {
	desc := &scip.Descriptor{
		Name:          "testName",
		Suffix:        scip.Descriptor_Type,
		Disambiguator: "testDisambiguator",
	}

	expected := model.Descriptor{
		Name:          "testName",
		Suffix:        scip.Descriptor_Type,
		Disambiguator: "testDisambiguator",
	}

	result := ScipDescriptorToModelDescriptor(desc)

	assert.Equal(t, expected, result)
}

func TestScipDescriptorsToModelDescriptors(t *testing.T) {
	tests := []struct {
		name     string
		descs    []*scip.Descriptor
		expected []model.Descriptor
	}{
		{
			name: "Single descriptor",
			descs: []*scip.Descriptor{
				{
					Name:          "testName",
					Suffix:        scip.Descriptor_Type,
					Disambiguator: "testDisambiguator",
				},
			},
			expected: []model.Descriptor{
				{
					Name:          "testName",
					Suffix:        scip.Descriptor_Type,
					Disambiguator: "testDisambiguator",
				},
			},
		},
		{
			name: "Multiple descriptors",
			descs: []*scip.Descriptor{
				{
					Name:          "name1",
					Suffix:        scip.Descriptor_Type,
					Disambiguator: "dis1",
				},
				{
					Name:          "name2",
					Suffix:        scip.Descriptor_Term,
					Disambiguator: "dis2",
				},
			},
			expected: []model.Descriptor{
				{
					Name:          "name1",
					Suffix:        scip.Descriptor_Type,
					Disambiguator: "dis1",
				},
				{
					Name:          "name2",
					Suffix:        scip.Descriptor_Term,
					Disambiguator: "dis2",
				},
			},
		},
		{
			name:     "Empty descriptor",
			descs:    []*scip.Descriptor{},
			expected: []model.Descriptor{},
		},
		{
			name:     "Nil descriptors",
			descs:    nil,
			expected: []model.Descriptor{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScipDescriptorsToModelDescriptors(tt.descs)
			assert.Equal(t, tt.expected, result)
		})
	}
}
