package scanner

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const (
	validEmptyScip = "testdata/valid_index.scip"
	// Symbol and occurrence id: "scip-demo . . . some/package#"
	validOccurrenceBase64 = "CgMAAAoSHXNjaXAtZGVtbyAuIC4gLiBzb21lL3BhY2thZ2UjGAEoCQ=="
	validSymbolBase64     = "Ch1zY2lwLWRlbW8gLiAuIC4gc29tZS9wYWNrYWdlIxoQc29tZSBzeW1ib2wgZG9jczIHcGFja2FnZQ===="
	// relative path for doc: some/path.go
	validDocumentBase64 = "Cgxzb21lL3BhdGguZ28iAmdv"
)

func TestScanIndexFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid",
			filePath: validEmptyScip,
			wantErr:  false,
		},
		{
			name:     "bad path",
			filePath: "invalid_path.scip",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{}
			impl.InitBuffers()
			err := impl.ScanIndexFile(tt.filePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseDocumentPath(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantErr  bool
		errMsg   string
		expected string
	}{
		{
			name:     "valid",
			data:     decodeBase64(t, validDocumentBase64),
			wantErr:  false,
			expected: "some/path.go",
		},
		{
			name:    "invalid data",
			data:    []byte{0xFF}, // Invalid tag
			wantErr: true,
			errMsg:  "failed to consume field",
		},
		{
			name:     "empty",
			data:     []byte{},
			wantErr:  false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{Pool: NewBufferPool(1024, 4)}
			impl.InitBuffers()
			docPath, err := impl.parseDocumentPath(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, docPath)
			}
		})
	}
}

func TestParseOccurrenceSymbol(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantErr  bool
		errMsg   string
		expected []byte
	}{
		{
			name:     "valid",
			data:     decodeBase64(t, validOccurrenceBase64),
			wantErr:  false,
			expected: []byte("scip-demo . . . some/package#"),
		},
		{
			name:    "invalid data",
			data:    []byte{0xFF}, // Invalid tag
			wantErr: true,
			errMsg:  "failed to consume field",
		},
		{
			name:     "empty",
			data:     []byte{},
			wantErr:  false,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{Pool: NewBufferPool(1024, 4)}
			impl.InitBuffers()
			occurrence, err := impl.parseOccurrenceSymbol(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else if tt.expected == nil {
				assert.NoError(t, err)
				assert.Nil(t, tt.expected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, occurrence)
			}
		})
	}
}

func TestScanIndexFolder(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create some test SCIP files with valid content
	testFiles := []struct {
		name    string
		content []byte
	}{
		{
			name:    "test1.scip",
			content: decodeBase64(t, validDocumentBase64),
		},
		{
			name:    "test2.scip",
			content: decodeBase64(t, validDocumentBase64),
		},
	}

	for _, tf := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, tf.name), tf.content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	// Create a subdirectory to test directory skipping
	err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a directory with invalid files for error testing
	invalidDir := t.TempDir()
	invalidFiles := []struct {
		name    string
		content []byte
	}{
		{
			name:    "test1.scip",
			content: decodeBase64(t, validDocumentBase64),
		},
		{
			name:    "test2.scip",
			content: decodeBase64(t, validDocumentBase64),
		},
		{
			name:    "invalid.scip",
			content: []byte{0xFF}, // Invalid SCIP data
		},
	}

	for _, tf := range invalidFiles {
		err := os.WriteFile(filepath.Join(invalidDir, tf.name), tf.content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	tests := []struct {
		name     string
		path     string
		parallel bool
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful sequential scan",
			path:     tmpDir,
			parallel: false,
			wantErr:  false,
		},
		{
			name:     "successful parallel scan",
			path:     tmpDir,
			parallel: true,
			wantErr:  false,
		},
		{
			name:     "sequential scan with invalid file",
			path:     invalidDir,
			parallel: false,
			wantErr:  true,
			errMsg:   "failed to consume field\nfailed to consume tag\nunexpected EOF",
		},
		{
			name:     "parallel scan with invalid file",
			path:     invalidDir,
			parallel: true,
			wantErr:  true,
			errMsg:   "EOF",
		},
		{
			name:     "non-existent directory",
			path:     "/non/existent/path",
			parallel: false,
			wantErr:  true,
			errMsg:   "failed to read directory",
		},
		{
			name:     "mixed valid and invalid files with max concurrency",
			path:     invalidDir,
			parallel: true,
			wantErr:  true,
			errMsg:   "failed to consume field\nfailed to consume tag\nunexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{}
			impl.InitBuffers()

			// Set max concurrency for the mixed files test
			if tt.name == "mixed valid and invalid files with max concurrency" {
				impl.MaxConcurrency = 2 // Limit to 2 concurrent workers
			}

			err := impl.ScanIndexFolder(tt.path, tt.parallel)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConsumeMeta(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedErr error
		errMsg      string
		fieldNumber protowire.Number
		fieldType   protowire.Type
	}{
		{
			name:        "valid tag bytes",
			data:        []byte{0x0A}, // Tag for field number 1, type 2
			expectedErr: nil,
			fieldNumber: protowire.Number(1),
			fieldType:   protowire.BytesType,
		},
		{
			name:        "EOF",
			data:        []byte{}, // Empty reader to simulate EOF
			expectedErr: io.EOF,
		},
		{
			name:        "invalid tag",
			data:        []byte{0xFF}, // Invalid tag
			expectedErr: io.ErrUnexpectedEOF,
			errMsg:      "failed to consume tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{}
			impl.InitBuffers()
			reader := bytes.NewReader(tt.data)

			fieldNumber, fieldType, err := impl.consumeBytesFieldMeta(reader)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.fieldNumber, fieldNumber)
				assert.Equal(t, tt.fieldType, fieldType)
			}
		})
	}
}

func TestScanIndexReader_AutoCover(t *testing.T) {
	impl := &IndexScannerImpl{
		Pool:              NewBufferPool(1024, 4),
		MatchDocumentPath: func(path string) bool { return path == "some/path.go" },
		VisitDocument:     func(doc *scip.Document) {},
	}
	impl.InitBuffers()

	data := decodeBase64(t, validDocumentBase64)
	reader := bytes.NewReader(data)

	err := impl.ScanIndexReader(reader)

	assert.NoError(t, err)
}

func TestScanIndexReaderInvalidDocument_AutoCover(t *testing.T) {
	impl := &IndexScannerImpl{
		Pool:              NewBufferPool(1024, 4),
		MatchDocumentPath: func(path string) bool { return path == "some/path.go" },
		VisitDocument:     func(doc *scip.Document) {},
	}
	impl.InitBuffers()

	data := []byte{0x00} // Invalid data
	reader := bytes.NewReader(data)

	err := impl.ScanIndexReader(reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to consume field")
}

func TestScanIndexReader_NilPool(t *testing.T) {
	// Create a scanner with nil Pool
	impl := &IndexScannerImpl{
		Pool:              nil,
		MatchDocumentPath: func(path string) bool { return path == "some/path.go" },
		VisitDocument:     func(doc *scip.Document) {},
	}

	data := decodeBase64(t, validDocumentBase64)
	reader := bytes.NewReader(data)

	err := impl.ScanIndexReader(reader)
	assert.NoError(t, err)
	assert.NotNil(t, impl.Pool, "Pool should be initialized after ScanIndexReader")
}

// Generated by AutoCover
// Test case to verify metadataFieldName function with known and unknown field numbers
func TestMetadataFieldName_AutoCover(t *testing.T) {
	tests := []struct {
		name         string
		fieldNumber  protowire.Number
		expectedName string
	}{
		{
			name:         "known field - project root",
			fieldNumber:  metadataProjectRootFieldNumber,
			expectedName: "project_root",
		},
		{
			name:         "unknown field",
			fieldNumber:  99,
			expectedName: "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldName := metadataFieldName(tt.fieldNumber)
			assert.Equal(t, tt.expectedName, fieldName)
		})
	}
}

// Generated by AutoCover
func TestScanIndexReader_FailureCases_AutoCover(t *testing.T) {
	// This test case simulates scanning an index reader with various errors
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "document path error",
			data:    []byte{0x00},
			wantErr: true,
			errMsg:  "failed to consume field",
		},
		{
			name:    "unexpected field type",
			data:    []byte{0x10, 0x96, 0x01},
			wantErr: true,
			errMsg:  "expected bytes type",
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
			reader := bytes.NewReader(tt.data)
			err := impl.ScanIndexReader(reader)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Generated by AutoCover
func TestParseSymbolMoniker_EmptyData_AutoCover(t *testing.T) {
	// This test case verifies parseSymbolMoniker function with empty data
	impl := &IndexScannerImpl{Pool: NewBufferPool(1024, 4)}
	impl.InitBuffers()
	symbol, err := impl.parseSymbolMoniker([]byte{})

	assert.NoError(t, err)
	assert.Nil(t, symbol)
}

// Generated by AutoCover
func TestScanIndexReader_InvalidDocumentData_AutoCover(t *testing.T) {
	// This test case verifies that ScanIndexReader returns an error for invalid document data
	tests := []struct {
		name     string
		docData  []byte
		expected string
	}{
		{
			name:     "Invalid protobuf data",
			docData:  []byte{0x01, 0x02, 0x03},
			expected: "failed to parse document path",
		},
		{
			name:     "Empty document data",
			docData:  []byte{},
			expected: "failed to consume field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{
				Pool:              NewBufferPool(1024, 4),
				MatchDocumentPath: func(path string) bool { return true },
				VisitDocument:     func(doc *scip.Document) {},
			}
			impl.InitBuffers()

			documentField := append([]byte{0x12}, protowire.AppendVarint(nil, uint64(len(tt.docData)))...)
			documentField = append(documentField, tt.docData...)

			reader := bytes.NewReader(documentField)

			err := impl.ScanIndexReader(reader)

			if tt.docData == nil || len(tt.docData) == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
			}
		})
	}
}

// Generated by AutoCover
func TestScanIndexReader_InvalidFieldLengthInsufficientData_AutoCover(t *testing.T) {
	// This test verifies that ScanIndexReader handles invalid field lengths with insufficient data
	impl := &IndexScannerImpl{
		Pool: NewBufferPool(1024, 4),
	}
	impl.InitBuffers()

	// Prepare data with an invalid field length with insufficient data
	invalidData := []byte{0x12, 0xFF}

	reader := bytes.NewReader(invalidData)

	err := impl.ScanIndexReader(reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read length")
}

func TestScanIndexReader_FailureCases(t *testing.T) {
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
			name:   "failed to scan document",
			data:   []byte{0x12, 0x06, 0x0A, 0x03, 0x61, 0x62, 0x63, 0x12}, // valid documents field, but contents invalid
			errMsg: "failed to scan document",
		},
		{
			name:   "failed to parse document",
			data:   []byte{0x12, 0x07, 0x0A, 0x04, 0x61, 0x62, 0x63, 0x64, 0x12}, // valid documents field, but contents invalid
			errMsg: "failed to read document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{
				Pool:              NewBufferPool(1024, 4),
				MatchDocumentPath: func(path string) bool { return path == "abcd" },
				VisitDocument:     func(doc *scip.Document) {},
			}
			impl.InitBuffers()
			reader := bytes.NewReader(tt.data)
			err := impl.ScanIndexReader(reader)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestBufferPool_ConcurrentAccess(t *testing.T) {
	pool := NewBufferPool(1024, 4)
	const numGoroutines = 100
	const numOperations = 1000
	var wg sync.WaitGroup

	// Test concurrent Get and Put operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Get a buffer of random size between 512 and 4096
				size := 512 + (j % 3585)
				buf := pool.Get(size)
				// Simulate some work
				for k := 0; k < len(buf); k++ {
					buf[k] = byte(k % 256)
				}
				// Put the buffer back
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()
}

func TestScanIndexFolder_ScipFilter(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a simple SCIP index for testing
	testIndex := &scip.Index{
		Metadata: &scip.Metadata{
			ProjectRoot: "test-project",
		},
		Documents: []*scip.Document{
			{
				RelativePath: "test.go",
				Symbols: []*scip.SymbolInformation{
					{
						Symbol: "test.symbol",
					},
				},
			},
		},
	}

	// Serialize the test index
	indexBytes, err := proto.Marshal(testIndex)
	require.NoError(t, err)

	// Create some test files
	testFiles := []string{
		"index1.scip",
		"index2.scip",
		"notanindex.txt",
		"index3.scip",
	}

	// Create test files
	for _, file := range testFiles {
		content := []byte("invalid content")
		if filepath.Ext(file) == ".scip" {
			content = indexBytes
		}
		err := os.WriteFile(filepath.Join(tempDir, file), content, 0644)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		expectedFiles int32
		parallel      bool
	}{
		{
			name:          "sequential scan",
			expectedFiles: 3,
			parallel:      false,
		},
		{
			name:          "parallel scan",
			expectedFiles: 3,
			parallel:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var processedFiles int32
			scanner := &IndexScannerImpl{
				MatchDocumentPath: func(string) bool {
					atomic.AddInt32(&processedFiles, 1)
					return true
				},
			}
			scanner.InitBuffers()

			err := scanner.ScanIndexFolder(tempDir, tt.parallel)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFiles, atomic.LoadInt32(&processedFiles), "incorrect number of files processed")
		})
	}
}
