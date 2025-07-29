package scanner

import (
	"bytes"
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

func TestScanDocument(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		reader  *bytes.Reader
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid",
			path:    "some/path.go",
			reader:  bytes.NewReader(decodeBase64(t, validDocumentBase64)),
			wantErr: false,
		},
		{
			name:    "invalid data",
			path:    "some/path.go",
			reader:  bytes.NewReader([]byte{0x00}),
			wantErr: true,
			errMsg:  "failed to consume field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{
				Pool:              NewBufferPool(1024, 4),
				MatchDocumentPath: func(path string) bool { return path == "some/path.go" },
				VisitDocument:     func(doc *scip.Document) {},
			}
			impl.InitBuffers()
			err := impl.ScanDocumentReader(tt.path, tt.reader)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShouldParseDocumentField_AutoCover(t *testing.T) {
	impl := &IndexScannerImpl{
		MatchSymbol:     func(_ []byte) bool { return true },
		MatchOccurrence: func(_ []byte) bool { return true },
	}
	impl.InitBuffers()

	tests := []struct {
		name          string
		fieldNumber   protowire.Number
		expectedParse bool
	}{
		{
			name:          "parse symbols field",
			fieldNumber:   documentSymbolsFieldNumber,
			expectedParse: true,
		},
		{
			name:          "parse occurrences field",
			fieldNumber:   documentOccurrencesFieldNumber,
			expectedParse: true,
		},
		{
			name:          "do not parse unknown field",
			fieldNumber:   9999,
			expectedParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldParse := impl.shouldParseDocumentField(tt.fieldNumber)
			assert.Equal(t, tt.expectedParse, shouldParse)
		})
	}
}

func TestScanDocument_InvalidFieldType_AutoCover(t *testing.T) {
	tests := []struct {
		name        string
		fieldNumber byte
		expectedErr string
	}{
		{
			name:        "invalid field type for field number 1",
			fieldNumber: 0x0D, // 0x0D represents field number 1 with an invalid wire type 5
			expectedErr: "expected bytes type for field 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{
				Pool: NewBufferPool(1024, 4),
			}
			impl.InitBuffers()

			// Prepare data with an invalid field type
			invalidData := []byte{tt.fieldNumber}

			reader := bytes.NewReader(invalidData)

			err := impl.ScanDocumentReader("test.go", reader)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestScanDocument_InvalidSymbolProto_AutoCover(t *testing.T) {
	// This test case verifies that ScanDocument returns an error when the symbol proto is invalid
	impl := &IndexScannerImpl{
		Pool:        NewBufferPool(1024, 4),
		MatchSymbol: func(symbol []byte) bool { return true },
		VisitSymbol: func(docPath string, symbol *scip.SymbolInformation) {},
	}
	impl.InitBuffers()

	// Prepare data with an invalid symbol proto
	invalidSymbolProto := []byte{0x01, 0x02, 0x03} // Invalid proto data
	symbolField := append([]byte{0x1A}, protowire.AppendVarint(nil, uint64(len(invalidSymbolProto)))...)
	symbolField = append(symbolField, invalidSymbolProto...)

	reader := bytes.NewReader(symbolField)

	err := impl.ScanDocumentReader("test.go", reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse symbol moniker")
}

func TestScanDocument_InvalidOccurrenceProto_AutoCover(t *testing.T) {
	// Verifies that ScanDocument returns an error when the occurrence proto is invalid
	tests := []struct {
		name           string
		occurrenceData []byte
		expectedError  string
	}{
		{
			name:           "Invalid occurrence proto",
			occurrenceData: []byte{0x01, 0x02, 0x03},
			expectedError:  "failed to parse symbol moniker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{
				Pool:            NewBufferPool(1024, 4),
				MatchOccurrence: func(symbol []byte) bool { return true },
				VisitOccurrence: func(docPath string, occurrence *scip.Occurrence) {},
			}
			impl.InitBuffers()

			occurrenceField := append([]byte{0x12}, protowire.AppendVarint(nil, uint64(len(tt.occurrenceData)))...)
			occurrenceField = append(occurrenceField, tt.occurrenceData...)

			reader := bytes.NewReader(occurrenceField)

			err := impl.ScanDocumentReader("test.go", reader)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestScanDocument_FailureCases(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		errMsg string
	}{
		{
			name:   "failed to consume field",
			data:   []byte{0xFF}, // Invalid tag
			errMsg: "failed to consume field",
		},
		{
			name:   "failed to consume length",
			data:   []byte{0x12, 0xFF}, // Valid tag but invalid length
			errMsg: "failed to consume length",
		},
		{
			name:   "failed to consume data",
			data:   []byte{0x12, 0x02, 0x08}, // Valid tag and length but insufficient data
			errMsg: "failed to consume",
		},
		{
			name:   "invalid symbol information data",
			data:   []byte{0x1A, 0x03, 0x0A, 0x03, 0x61, 0x62, 0x63}, // Bytes field (number 3) with length 3 and value "abc"
			errMsg: "failed to parse symbol moniker",
		},
		{
			name:   "invalid occurrence data",
			data:   []byte{0x12, 0x03, 0x0B, 0x03, 0x61, 0x62, 0x63, 0x88}, // Bytes field (number 2) with length 3 and value "abc"
			errMsg: "failed to parse symbol moniker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{
				Pool:            NewBufferPool(1024, 4),
				MatchSymbol:     func(symbol []byte) bool { return string(symbol) == "abc" },
				MatchOccurrence: func(symbol []byte) bool { return string(symbol) == "abc" },
				VisitDocument:   func(doc *scip.Document) {},
			}
			impl.InitBuffers()
			reader := bytes.NewReader(tt.data)
			err := impl.ScanDocumentReader("some/path.go", reader)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// Generated by AutoCover
func TestMatchDocument_AutoCover(t *testing.T) {
	// This test case verifies that ScanIndexReader correctly matches and visits a document
	visitedDocument := false
	visitedSymbol := false
	visitedOccurrence := false

	impl := &IndexScannerImpl{
		Pool:              NewBufferPool(1024, 4),
		MatchDocumentPath: func(path string) bool { return path == "some/path.go" },
		VisitDocument:     func(doc *scip.Document) { visitedDocument = true },
		MatchSymbol:       func(symbol []byte) bool { return string(symbol) == "test_symbol" },
		VisitSymbol:       func(docPath string, symbol *scip.SymbolInformation) { visitedSymbol = true },
		MatchOccurrence:   func(symbol []byte) bool { return string(symbol) == "test_symbol" },
		VisitOccurrence:   func(docPath string, occurrence *scip.Occurrence) { visitedOccurrence = true },
	}
	impl.InitBuffers()

	// Create a valid document with symbols and occurrences
	document := &scip.Document{
		RelativePath: "some/path.go",
		Symbols: []*scip.SymbolInformation{
			{Symbol: "test_symbol"},
		},
		Occurrences: []*scip.Occurrence{
			{Symbol: "test_symbol"},
		},
	}
	documentData, err := proto.Marshal(document)
	assert.NoError(t, err)

	// Prepare the index data
	indexData := append([]byte{0x12}, protowire.AppendVarint(nil, uint64(len(documentData)))...)
	indexData = append(indexData, documentData...)

	reader := bytes.NewReader(indexData)

	err = impl.ScanIndexReader(reader)

	assert.NoError(t, err)
	assert.True(t, visitedDocument, "Document should have been visited")
	assert.True(t, visitedSymbol, "Symbol should have been visited")
	assert.True(t, visitedOccurrence, "Occurrence should have been visited")
}
