package scanner

import (
	"bytes"
	b64 "encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestInitBuffers(t *testing.T) {
	impl := &IndexScannerImpl{}
	impl.InitBuffers()
	assert.NotNil(t, impl.Pool)
}

func TestConsumeLen(t *testing.T) {
	tests := []struct {
		name      string
		reader    *bytes.Reader
		fieldType protowire.Type
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "incorrect field type",
			reader:    bytes.NewReader([]byte{0x10}),
			fieldType: protowire.VarintType,
			wantErr:   true,
			errMsg:    "expected LEN type tag",
		},
		{
			name:      "insufficient data",
			reader:    bytes.NewReader([]byte{}),
			fieldType: protowire.BytesType,
			wantErr:   true,
			errMsg:    "failed to read length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{}
			impl.InitBuffers()
			_, err := impl.consumeLen(tt.reader, tt.fieldType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConsumeFieldBytesMeta_Error(t *testing.T) {
	impl := &IndexScannerImpl{}
	impl.InitBuffers()
	ctrl := gomock.NewController(t)
	mockReader := NewMockScipReader(ctrl)
	mockReader.EXPECT().Read(gomock.Any()).Return(0, errors.New("unexpected"))

	_, _, err := impl.consumeBytesFieldMeta(mockReader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read from index reader")
}

func TestConsumeFieldBytesMeta_NothingRead(t *testing.T) {
	impl := &IndexScannerImpl{}
	impl.InitBuffers()
	ctrl := gomock.NewController(t)
	mockReader := NewMockScipReader(ctrl)
	mockReader.EXPECT().Read(gomock.Any()).Return(0, nil)

	_, _, err := impl.consumeBytesFieldMeta(mockReader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read 0 bytes from index")
}

func TestConsumeFieldData(t *testing.T) {
	tests := []struct {
		name     string
		reader   *bytes.Reader
		length   int
		dataBuf  []byte
		wantErr  bool
		errMsg   string
		expected []byte
	}{
		{
			name:     "successful read",
			reader:   bytes.NewReader([]byte{0x08, 0x96, 0x01}),
			length:   2,
			dataBuf:  make([]byte, 2),
			wantErr:  false,
			expected: []byte{0x08, 0x96},
		},
		{
			name:    "insufficient data",
			reader:  bytes.NewReader([]byte{0x08}),
			length:  2,
			dataBuf: make([]byte, 2),
			wantErr: true,
			errMsg:  "expected to read 2 bytes based on LEN but read 1 bytes",
		},
		{
			name:    "empty reader",
			reader:  bytes.NewReader([]byte{}),
			length:  2,
			dataBuf: make([]byte, 2),
			wantErr: true,
			errMsg:  "failed to read data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &IndexScannerImpl{}
			impl.InitBuffers()
			err := impl.consumeFieldData(tt.reader, tt.length, tt.dataBuf)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.dataBuf)
			}
		})
	}
}

func TestSkipField(t *testing.T) {
	impl := &IndexScannerImpl{}
	impl.InitBuffers()
	reader := bytes.NewReader([]byte{0x10})
	assert.NoError(t, impl.skipField(reader))
}

func setupMock(t *testing.T, returnBytes int, returnErr error) ScipReader {
	ctrl := gomock.NewController(t)
	mockReader := NewMockScipReader(ctrl)
	mockReader.EXPECT().Read(gomock.Any()).Return(returnBytes, returnErr)
	return mockReader
}

func TestReadVarint(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) ScipReader
		wantErr   bool
		errMsg    string
	}{
		{
			name: "eof",
			setupFunc: func(t *testing.T) ScipReader {
				return bytes.NewReader([]byte{})
			},
			wantErr: true,
			errMsg:  "failed to read 0-th byte of Varint",
		},
		{
			name: "zero bytes read",
			setupFunc: func(t *testing.T) ScipReader {
				return setupMock(t, 0, nil)
			},
			wantErr: true,
			errMsg:  "failed to read 0-th byte of Varint",
		},
		{
			name: "invalid varint",
			setupFunc: func(t *testing.T) ScipReader {
				return bytes.NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}) // 10 continuation bits set
			},
			wantErr: true,
			errMsg:  "variable length integer overflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := tt.setupFunc(t)
			scratchBuf := make([]byte, 0, 10)
			_, err := readVarint(reader, &scratchBuf)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test utilities

func decodeBase64(t *testing.T, str string) []byte {
	data, err := b64.StdEncoding.DecodeString(str)
	assert.NoError(t, err)
	return data
}

func readFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	return data
}

func TestIndexFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    protowire.Number
		expected string
	}{
		{
			name:     "metadata field",
			input:    indexMetadataFieldNumber,
			expected: "metadata",
		},
		{
			name:     "documents field",
			input:    indexDocumentsFieldNumber,
			expected: "documents",
		},
		{
			name:     "external symbols field",
			input:    indexExternalSymbolsFieldNumber,
			expected: "external_symbols",
		},
		{
			name:     "unknown field",
			input:    999,
			expected: "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDocumentFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    protowire.Number
		expected string
	}{
		{
			name:     "relative path field",
			input:    documentRelativePathFieldNumber,
			expected: "relative_path",
		},
		{
			name:     "occurrences field",
			input:    documentOccurrencesFieldNumber,
			expected: "occurrences",
		},
		{
			name:     "symbols field",
			input:    documentSymbolsFieldNumber,
			expected: "symbols",
		},
		{
			name:     "language field",
			input:    documentLanguageFieldNumber,
			expected: "language",
		},
		{
			name:     "unknown field",
			input:    999,
			expected: "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := documentFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetadataFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    protowire.Number
		expected string
	}{
		{
			name:     "project root field",
			input:    metadataProjectRootFieldNumber,
			expected: "project_root",
		},
		{
			name:     "unknown field",
			input:    999,
			expected: "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metadataFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitBuffers_NilPool(t *testing.T) {
	tests := []struct {
		name string
		impl *IndexScannerImpl
	}{
		{
			name: "nil pool",
			impl: &IndexScannerImpl{Pool: nil},
		},
		{
			name: "existing pool",
			impl: &IndexScannerImpl{Pool: NewBufferPool(1024, 4)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingPool := tt.impl.Pool
			tt.impl.InitBuffers()
			assert.NotNil(t, tt.impl.Pool, "Pool should not be nil after InitBuffers")
			if existingPool != nil {
				assert.Equal(t, existingPool, tt.impl.Pool, "Existing pool should not be replaced")
			}
		})
	}
}
