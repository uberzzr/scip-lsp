package scip

import (
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/scip-lib/model"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

const (
	goSrcPkg     = "gomod github.com/golang/go/src 1.20.3"
	goSrcPkgName = "github.com/golang/go/src"
	getEnvID     = "scip-go gomod github.com/golang/go/src 1.20.3 os/Getenv()."

	tracingPkg              = "gomod code.uber.internal/devexp/test_management/tracing 0f67d80e60274b77875a241c43ef980bc9ffe0d8"
	tracingPkgName          = "code.uber.internal/devexp/test_management/tracing"
	tracingUUIDKey          = "scip-go gomod code.uber.internal/devexp/test_management/tracing 0f67d80e60274b77875a241c43ef980bc9ffe0d8 `code.uber.internal/devexp/test_management/tracing`/PipelineUUIDTagKey."
	tracingNoDisplayNameKey = "scip-go gomod code.uber.internal/devexp/test_management/tracing 0f67 `code.uber.internal/devexp/test_management/tracing`/NoDisplayName."
	tracingFile             = "src/code.uber.internal/devexp/test_management/tracing/span.go"

	fievelEntitiesPkg     = "maven com/uber/hcv/route/generation/entities "
	fievelEntitiesPkgName = "com/uber/hcv/route/generation/entities"
	fievelCentroidKey     = "semanticdb maven . . com/uber/hcv/route/generation/entities/CentroidPath#"
	javaUtilListKey       = "semanticdb maven jdk 21 java/util/List#"

	ulspUnknownPkg     = "ulsp unknown 0.0.1"
	ulspUnknownPkgName = "unknown"
	localPkg           = "local"
	local0             = "local 0"
)

func TestIsMatchingPosition(t *testing.T) {
	tests := []struct {
		name     string
		position protocol.Position
		rangePos []int32
		expected bool
	}{
		{
			name:     "position is within range",
			position: protocol.Position{Line: 0, Character: 1},
			rangePos: []int32{0, 0, 2},
			expected: true,
		},
		{
			name:     "position is before range",
			position: protocol.Position{Line: 0, Character: 0},
			rangePos: []int32{1, 1, 2},
			expected: false,
		},
		{
			name:     "position is after range",
			position: protocol.Position{Line: 2, Character: 2},
			rangePos: []int32{1, 1, 2},
			expected: false,
		},
		{
			name:     "range is multiline, position in range",
			position: protocol.Position{Line: 1, Character: 0},
			rangePos: []int32{0, 0, 2, 10},
			expected: true,
		},
		{
			name:     "range is multiline, position out of range",
			position: protocol.Position{Line: 0, Character: 3},
			rangePos: []int32{1, 1, 2, 2},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			occ := model.Occurrence{
				Range: test.rangePos,
			}
			actual := IsMatchingPosition(&occ, test.position)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetOccurrenceForPosition(t *testing.T) {
	tests := []struct {
		name     string
		position protocol.Position
		occs     []*model.Occurrence
		expected *model.Occurrence
	}{
		{
			name:     "occurrence exists",
			position: protocol.Position{Line: 0, Character: 1},
			occs: []*model.Occurrence{
				{Range: []int32{0, 0, 2}},
				{Range: []int32{0, 3, 4}},
			},
			expected: &model.Occurrence{Range: []int32{0, 0, 2}},
		},
		{
			name:     "occurrence is at the end",
			position: protocol.Position{Line: 5, Character: 4},
			occs: []*model.Occurrence{
				{Range: []int32{0, 0, 2}},
				{Range: []int32{0, 3, 4}},
				{Range: []int32{1, 3, 4}},
				{Range: []int32{2, 3, 4}},
				{Range: []int32{3, 3, 4}},
				{Range: []int32{5, 3, 4}},
			},
			expected: &model.Occurrence{Range: []int32{5, 3, 4}},
		},
		{
			name:     "occurrence doesn't exist",
			position: protocol.Position{Line: 20, Character: 4},
			occs: []*model.Occurrence{
				{Range: []int32{0, 0, 1, 1}},
				{Range: []int32{1, 2, 2}},
				{Range: []int32{1, 3, 4}},
				{Range: []int32{2, 3, 4}},
				{Range: []int32{3, 3, 4}},
				{Range: []int32{5, 3, 4}},
			},
			expected: nil,
		},
		{
			name:     "matching occurrence is multiline",
			position: protocol.Position{Line: 20, Character: 4},
			occs: []*model.Occurrence{
				{Range: []int32{0, 0, 2}},
				{Range: []int32{0, 3, 4}},
				{Range: []int32{1, 3, 4}},
				{Range: []int32{2, 3, 4}},
				{Range: []int32{3, 3, 4}},
				{Range: []int32{18, 3, 21, 4}},
			},
			expected: &model.Occurrence{Range: []int32{18, 3, 21, 4}},
		},
		{
			name:     "matching occurrence is early in the list",
			position: protocol.Position{Line: 0, Character: 0},
			occs: []*model.Occurrence{
				{Range: []int32{0, 0, 2}},
				{Range: []int32{0, 3, 4}},
				{Range: []int32{1, 3, 4}},
				{Range: []int32{2, 3, 4}},
				{Range: []int32{3, 3, 4}},
				{Range: []int32{18, 3, 21}},
			},
			expected: &model.Occurrence{Range: []int32{0, 0, 2}},
		},
		{
			name:     "all occurrences are multiline",
			position: protocol.Position{Line: 0, Character: 0},
			occs: []*model.Occurrence{
				{Range: []int32{0, 0, 1, 2}},
				{Range: []int32{1, 3, 2, 4}},
				{Range: []int32{2, 4, 3, 1}},
				{Range: []int32{3, 2, 4, 4}},
				{Range: []int32{5, 3, 7, 4}},
				{Range: []int32{18, 3, 21, 1}},
			},
			expected: &model.Occurrence{Range: []int32{0, 0, 1, 2}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetOccurrenceForPosition(test.occs, test.position)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetSymbolForPosition(t *testing.T) {
	tmpDir := bazel.TestTmpDir()
	reg := getFakeRegistryImpl(tmpDir)

	tests := []struct {
		name        string
		file        uri.URI
		position    protocol.Position
		expected    *SymbolData
		expectedOcc *model.Occurrence
	}{
		{
			name:        "symbol exists",
			file:        uri.File("tracing.go"),
			position:    protocol.Position{Line: 0, Character: 0},
			expected:    reg.Packages[goSrcPkgName].SymbolData[getEnvID],
			expectedOcc: reg.Documents[uri.File("tracing.go")].Document.Occurrences[0],
		},
		{
			name:        "symbol doesn't exist",
			file:        uri.File("tracing.go"),
			position:    protocol.Position{Line: 20, Character: 4},
			expected:    nil,
			expectedOcc: nil,
		},
		{
			name:        "symbol is local",
			file:        uri.File("tracing.go"),
			position:    protocol.Position{Line: 3, Character: 2},
			expected:    reg.Packages[goSrcPkgName].SymbolData[local0],
			expectedOcc: reg.Documents[uri.File("tracing.go")].Document.Occurrences[1],
		},
		{
			name:        "no package",
			file:        uri.File("tracing.go"),
			position:    protocol.Position{Line: 3, Character: 2},
			expected:    reg.Packages[goSrcPkgName].SymbolData[local0],
			expectedOcc: reg.Documents[uri.File("tracing.go")].Document.Occurrences[1],
		},
		{
			name:        "no file",
			file:        uri.File("tracing.go"),
			position:    protocol.Position{Line: 10, Character: 2},
			expected:    nil,
			expectedOcc: reg.Documents[uri.File("tracing.go")].Document.Occurrences[2],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			occ, actual, err := reg.GetSymbolForPosition(test.file, test.position)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedOcc, occ)
		})
	}

	t.Run("unindexed file", func(t *testing.T) {
		occ, actual, err := reg.GetSymbolForPosition(uri.File("notindexed.go"), protocol.Position{Line: 3, Character: 2})
		assert.NoError(t, err)
		assert.Nil(t, occ)
		assert.Nil(t, actual)
	})
}

func TestGetDocumentSymbolForFile(t *testing.T) {
	tmpDir := bazel.TestTmpDir()
	reg := getFakeRegistryImpl(tmpDir)

	tests := []struct {
		name     string
		file     uri.URI
		expected *[]*SymbolData
	}{
		{
			name:     "symbol doesn't exist",
			file:     uri.File("random.go"),
			expected: &[]*SymbolData{},
		},
		{
			name: "symbol exist",
			file: uri.File("tracing.go"),
			expected: &[]*SymbolData{
				reg.Packages[goSrcPkgName].SymbolData[getEnvID],
				reg.Documents[uri.File("tracing.go")].Locals[local0],
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := reg.GetDocumentSymbolForFile(test.file)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFileInfo(t *testing.T) {
	tmpDir := bazel.TestTmpDir()
	reg := getFakeRegistryImpl(tmpDir)

	tests := []struct {
		name     string
		file     uri.URI
		expected *FileInfo
	}{
		{
			name:     "file exists",
			file:     uri.File("tracing.go"),
			expected: reg.Documents[uri.File("tracing.go")],
		},
		{
			name:     "file does not exist",
			file:     uri.File("nonexistent.go"),
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := reg.GetFileInfo(test.file)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPackageInfo(t *testing.T) {
	tmpDir := bazel.TestTmpDir()
	reg := getFakeRegistryImpl(tmpDir)

	tests := []struct {
		name     string
		pkgID    PackageID
		expected *PackageMeta
	}{
		{
			name:     "package exists",
			pkgID:    PackageID(goSrcPkgName),
			expected: reg.Packages[PackageID(goSrcPkgName)],
		},
		{
			name:     "package does not exist",
			pkgID:    PackageID("nonexistent-package"),
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := reg.GetPackageInfo(test.pkgID)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetOccurrencesForSymbol(t *testing.T) {
	tests := []struct {
		name        string
		occurrences []*model.Occurrence
		symbol      string
		role        scip.SymbolRole
		expected    []*model.Occurrence
	}{
		{
			name: "unspecified role matches all roles",
			occurrences: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_WriteAccess),
				},
				{
					Symbol:      "other.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
			},
			symbol: "test.symbol",
			role:   -1,
			expected: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_WriteAccess),
				},
			},
		},
		{
			name: "matching symbol and role",
			occurrences: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition | scip.SymbolRole_WriteAccess),
				},
				{
					Symbol:      "other.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
			},
			symbol: "test.symbol",
			role:   scip.SymbolRole_Definition,
			expected: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition | scip.SymbolRole_WriteAccess),
				},
			},
		},
		{
			name: "multiple matches with read access role",
			occurrences: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				},
			},
			symbol: "test.symbol",
			role:   scip.SymbolRole_ReadAccess,
			expected: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				},
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				},
			},
		},
		{
			name: "no matches",
			occurrences: []*model.Occurrence{
				{
					Symbol:      "test.symbol",
					SymbolRoles: int32(scip.SymbolRole_Definition),
				},
			},
			symbol:   "other.symbol",
			role:     scip.SymbolRole_ReadAccess,
			expected: []*model.Occurrence{},
		},
		{
			name:        "empty occurrences",
			occurrences: []*model.Occurrence{},
			symbol:      "test.symbol",
			role:        scip.SymbolRole_ReadAccess,
			expected:    []*model.Occurrence{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetOccurrencesForSymbol(test.occurrences, test.symbol, test.role)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetLocalSymbolInformation(t *testing.T) {
	tests := []struct {
		name     string
		symbols  []*model.SymbolInformation
		symbol   string
		expected *model.SymbolInformation
	}{
		{
			name: "symbol exists",
			symbols: []*model.SymbolInformation{
				{
					Symbol: "local.symbol1",
				},
				{
					Symbol: "local.symbol2",
				},
			},
			symbol: "local.symbol1",
			expected: &model.SymbolInformation{
				Symbol: "local.symbol1",
			},
		},
		{
			name: "symbol does not exist",
			symbols: []*model.SymbolInformation{
				{
					Symbol: "local.symbol1",
				},
			},
			symbol:   "local.symbol2",
			expected: nil,
		},
		{
			name:     "empty symbols list",
			symbols:  []*model.SymbolInformation{},
			symbol:   "local.symbol1",
			expected: nil,
		},
		{
			name: "last symbol in list",
			symbols: []*model.SymbolInformation{
				{
					Symbol: "local.symbol1",
				},
				{
					Symbol: "local.symbol2",
				},
			},
			symbol: "local.symbol2",
			expected: &model.SymbolInformation{
				Symbol: "local.symbol2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetLocalSymbolInformation(test.symbols, test.symbol)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func getFakeRegistryImpl(tmpDir string) *registryImpl {
	srcPkg := &model.ScipPackage{
		Manager: "gomod",
		Name:    "github.com/golang/go/src",
		Version: "1.20.3",
	}
	reg := &registryImpl{
		WorkspaceRoot: tmpDir,
		Packages:      make(map[PackageID]*PackageMeta),
		Documents:     make(map[uri.URI]*FileInfo),
	}
	p := reg.getOrCreatePackage(srcPkg)

	reg.Documents[uri.File("tracing.go")] = &FileInfo{
		URI: uri.File("tracing.go"),
		Document: &model.Document{
			RelativePath: "tracing.go",
			Occurrences: []*model.Occurrence{
				{
					Symbol:      getEnvID,
					SymbolRoles: int32(scip.SymbolRole_Definition),
					Range:       []int32{0, 0, 1, 2},
				},
				{
					Symbol:      local0,
					SymbolRoles: int32(scip.SymbolRole_Definition),
					Range:       []int32{3, 0, 4},
				},
				{
					Symbol:      "scip-go gomod github.com/nopackage/go 1.20.3 os/Dontgetenv().",
					SymbolRoles: int32(scip.SymbolRole_ReadAccess),
					Range:       []int32{10, 0, 5},
				},
			},
		},
		Package: p,
		Locals:  map[string]*SymbolData{},
	}

	sd := p.getOrCreateSymbolData(getEnvID)
	sd.Definition = &model.Occurrence{
		Symbol:      getEnvID,
		SymbolRoles: int32(scip.SymbolRole_Definition),
		Range:       []int32{0, 0, 1, 2},
	}
	sd.Info = &model.SymbolInformation{
		Symbol:        getEnvID,
		Documentation: []string{},
		Kind:          scip.SymbolInformation_Function,
	}
	sd.Location = &protocol.LocationLink{
		TargetRange: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 1, Character: 2},
		},
		TargetURI: uri.File("github.com/golang/go/src/os/os.go"),
	}
	sd.References = map[PackageID][]*FileOccurences{
		PackageID(tracingPkgName): {
			{
				file: uri.File("tracing.go"),
				Occurrences: []*model.Occurrence{
					{
						Symbol:      getEnvID,
						SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						Range:       []int32{0, 0, 1, 2},
					},
				},
			},
		},
	}

	p.storeLocalSymbols(map[string]*model.SymbolInformation{
		local0: {
			Symbol:        local0,
			Documentation: []string{"hello there"},
		},
	}, map[string][]*model.Occurrence{
		local0: {
			&model.Occurrence{
				Symbol:      local0,
				SymbolRoles: int32(scip.SymbolRole_Definition),
				Range:       []int32{3, 0, 4},
			},
			&model.Occurrence{
				Symbol:      local0,
				SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				Range:       []int32{5, 0, 4},
			},
		},
	}, uri.File("tracing.go"))

	reg.Documents[uri.File("tracing.go")].storeLocalSymbols(map[string]*model.SymbolInformation{
		local0: {
			Symbol:        local0,
			Documentation: []string{"hello there"},
		},
	}, map[string][]*model.Occurrence{
		local0: {
			&model.Occurrence{
				Symbol:      local0,
				SymbolRoles: int32(scip.SymbolRole_Definition),
				Range:       []int32{3, 0, 4},
			},
			&model.Occurrence{
				Symbol:      local0,
				SymbolRoles: int32(scip.SymbolRole_ReadAccess),
				Range:       []int32{5, 0, 4},
			},
		},
	})

	return reg
}
