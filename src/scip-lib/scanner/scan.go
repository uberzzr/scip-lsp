package scanner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// IndexScanner defines the ways an index can be read.
type IndexScanner interface {
	ScanIndexFile(file string) error
	ScanIndexFolder(folder string, parallel bool) error
	ScanIndexReader(reader ScipReader) error
	ScanDocumentReader(relativeDocPath string, reader ScipReader) error
}

// ScipReader defines the minimum interface required for a reader
// to be passed into the IndexScanner
type ScipReader interface {
	io.Reader
	io.Seeker
}

// IndexScannerImpl implements the IndexScanner interface.
// A consumer can instantiate this implementation and implement
// any subset of the Match* methods, and (optionally) the matching
// Visit* method.
// After instantiating the struct, the consumer SHALL call .InitBuffers()
// Performance Considerations:
//  1. When multiple IndexScannerImpl's are expected to be instantiated at the same time,
//     a shared BufferPool can be provided to the struct's Pool property.
//     This improves performance and memory consumption.
//  2. The Match* methods use []byte's for performance reasons.
//  3. The Visit* method will only be called if: the Visit* method is defined AND
//     the Match* function returns true
//  4. Should the consumer only require the monikers (symbol IDs), using the Visit* method
//     and returning false is the recommended method.
//  5. Consider the size of the indices, there's an order of magnitude more occurrences
//     than symbols, and more symbols than documents. So if a consumer Matches & Visits every
//     occurrence, or document the performance gains over a traditional proto parser will
//     be limited.
type IndexScannerImpl struct {
	MatchDocumentPath func(string) bool
	MatchSymbol       func([]byte) bool
	MatchOccurrence   func([]byte) bool
	// TODO(IDE-1335): Implement MatchExternalSymbol
	// MatchExternalSymbol func([]byte) bool
	VisitDocument       func(*scip.Document)
	VisitSymbol         func(string, *scip.SymbolInformation)
	VisitOccurrence     func(string, *scip.Occurrence)
	VisitExternalSymbol func(*scip.SymbolInformation)
	Pool                *BufferPool
	MaxConcurrency      int
}

var _ IndexScanner = &IndexScannerImpl{}

// See https://protobuf.dev/programming-guides/encoding/#varints
const maxVarintBytes = 10

const (
	indexMetadataFieldNumber        = 1
	indexDocumentsFieldNumber       = 2
	indexExternalSymbolsFieldNumber = 3

	metadataProjectRootFieldNumber = 3

	documentRelativePathFieldNumber = 1
	documentOccurrencesFieldNumber  = 2
	documentSymbolsFieldNumber      = 3
	documentLanguageFieldNumber     = 4

	symbolInformationSymbolField = 1

	occurrenceSymbolField = 2
)

// InitBuffers initializes the relevant buffers, and pools
func (is *IndexScannerImpl) InitBuffers() {
	if is.Pool == nil {
		is.Pool = NewBufferPool(1024, 12)
	}
}

// ScanIndexReader incrementally processes an index by reading input from the io.Reader
// It skips over any other fields defined in the SCIP index.
func (is *IndexScannerImpl) ScanIndexReader(r ScipReader) error {
	if is.Pool == nil {
		is.InitBuffers()
	}
	document := scip.Document{}
	for {
		fieldNumber, fieldType, err := is.consumeBytesFieldMeta(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Join(fmt.Errorf("failed to consume field"), err)
		}
		if fieldNumber != indexDocumentsFieldNumber {
			err := is.skipField(r)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to skip %q", indexFieldName(fieldNumber)), err)
			}
			continue
		}
		dataLen, err := is.consumeLen(r, fieldType)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to consume length"), err)
		}
		dataBuf := is.Pool.Get(dataLen)
		err = is.consumeFieldData(r, dataLen, dataBuf)
		if err != nil {
			is.Pool.Put(dataBuf)
			return errors.Join(fmt.Errorf("failed to consume %q data", indexFieldName(fieldNumber)), err)
		}
		switch fieldNumber {
		case indexDocumentsFieldNumber:
			docPath, err := is.parseDocumentPath(dataBuf)
			if err != nil {
				is.Pool.Put(dataBuf)
				return errors.Join(fmt.Errorf("failed to parse document path"), err)
			}
			// Unmarshall the doc when all conditions match
			if is.MatchDocumentPath != nil && is.MatchDocumentPath(docPath) && is.VisitDocument != nil {
				if err := proto.Unmarshal(dataBuf, &document); err != nil {
					is.Pool.Put(dataBuf)
					return errors.Join(fmt.Errorf("failed to read document"), err)
				}
				is.VisitDocument(&document)
			}

			// Individually scna through the document to parse symbols/occurrences
			docReader := bytes.NewReader(dataBuf)
			if err := is.ScanDocumentReader(docPath, docReader); err != nil {
				is.Pool.Put(dataBuf)
				return errors.Join(fmt.Errorf("failed to scan document"), err)
			}
		}
		is.Pool.Put(dataBuf)
	}
}

// ScanDocumentReader scans through an individual document. This requires the path of the
// document only to pass into VisitSymbol/Occurrence methods.
func (is *IndexScannerImpl) ScanDocumentReader(docPath string, r ScipReader) error {
	symbolInfo := scip.SymbolInformation{}
	occurrence := scip.Occurrence{}
	for {
		fieldNumber, fieldType, err := is.consumeBytesFieldMeta(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Join(fmt.Errorf("failed to consume field"), err)
		}
		if !is.shouldParseDocumentField(fieldNumber) {
			err := is.skipField(r)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to skip %q", documentFieldName(fieldNumber)), err)
			}
			continue
		}
		dataLen, err := is.consumeLen(r, fieldType)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to consume length"), err)
		}
		dataBuf := is.Pool.Get(dataLen)
		err = is.consumeFieldData(r, dataLen, dataBuf)
		if err != nil {
			is.Pool.Put(dataBuf)
			return errors.Join(fmt.Errorf("failed to consume %q data", documentFieldName(fieldNumber)), err)
		}
		switch fieldNumber {
		case documentSymbolsFieldNumber:
			syByte, err := is.parseSymbolMoniker(dataBuf)
			if err != nil {
				is.Pool.Put(dataBuf)
				return errors.Join(fmt.Errorf("failed to parse symbol moniker"), err)
			}
			match := is.MatchSymbol(syByte)
			is.Pool.Put(syByte)
			if match && is.VisitSymbol != nil {
				if err := proto.Unmarshal(dataBuf, &symbolInfo); err != nil {
					is.Pool.Put(dataBuf)
					return errors.Join(fmt.Errorf("failed to read %q", documentFieldName(fieldNumber)), err)
				}
				is.VisitSymbol(docPath, &symbolInfo)
			}
		case documentOccurrencesFieldNumber:
			syByte, err := is.parseOccurrenceSymbol(dataBuf)
			if err != nil {
				is.Pool.Put(dataBuf)
				return errors.Join(fmt.Errorf("failed to parse symbol moniker"), err)
			}
			match := is.MatchOccurrence(syByte)
			is.Pool.Put(syByte)
			if match && is.VisitOccurrence != nil {
				if err := proto.Unmarshal(dataBuf, &occurrence); err != nil {
					is.Pool.Put(dataBuf)
					return errors.Join(fmt.Errorf("failed to read %q", documentFieldName(fieldNumber)), err)
				}
				is.VisitOccurrence(docPath, &occurrence)
			}
		}
		is.Pool.Put(dataBuf)
	}
}

// ScanIndexFile scans an individual index file.
func (is *IndexScannerImpl) ScanIndexFile(path string) error {
	reader, err := os.Open(path)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to open file"), err)
	}
	defer reader.Close()
	return is.ScanIndexReader(reader)
}

// ScanIndexFolder scans an entire folder's indices. Good to use when the consumer
// needs to scan through a bunch of indices for specific symbols.
func (is *IndexScannerImpl) ScanIndexFolder(path string, parallel bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to read directory"), err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(entries))

	processFile := func(filePath string) {
		defer wg.Done()
		// Skip non-SCIP files
		if filepath.Ext(filePath) != ".scip" {
			return
		}
		if err := is.ScanIndexFile(filePath); err != nil {
			errChan <- err
		}
	}

	if parallel {
		maxWorkers := runtime.NumCPU()
		if is.MaxConcurrency > 0 {
			maxWorkers = is.MaxConcurrency
		}
		sem := make(chan struct{}, maxWorkers)

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			wg.Add(1)
			fullPath := filepath.Join(path, entry.Name())

			sem <- struct{}{}
			go func(path string) {
				processFile(path)
				<-sem
			}(fullPath)
		}
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			wg.Add(1)
			fullPath := filepath.Join(path, entry.Name())
			processFile(fullPath)
		}
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		errStrings := make([]string, len(errs))
		for i, err := range errs {
			errStrings[i] = err.Error()
		}
		return errors.New(strings.Join(errStrings, "\n"))
	}

	return nil
}

// parseDocumentPath reads the relative path of the current document from the blob
func (is *IndexScannerImpl) parseDocumentPath(docData []byte) (string, error) {
	r := bytes.NewReader(docData)
	for {
		fieldNumber, fieldType, err := is.consumeBytesFieldMeta(r)
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			return "", errors.Join(fmt.Errorf("failed to consume field"), err)
		}

		if fieldNumber != documentRelativePathFieldNumber {
			err := is.skipField(r)
			if err != nil {
				return "", errors.Join(fmt.Errorf("failed to skip %q", documentFieldName(fieldNumber)), err)
			}
			continue
		}

		dataLen, err := is.consumeLen(r, fieldType)
		if err != nil {
			return "", errors.Join(fmt.Errorf("failed to consume length"), err)
		}
		dataBuf := is.Pool.Get(dataLen)
		defer is.Pool.Put(dataBuf)
		err = is.consumeFieldData(r, dataLen, dataBuf)
		if err != nil {
			return "", errors.Join(fmt.Errorf("failed to consume %q data", documentFieldName(fieldNumber)), err)
		}
		// Return immediately after reading the document path
		dataStr := string(dataBuf)
		return dataStr, nil
	}
}

// parseSymbolMoniker reads the symbol moniker from a symbolData blob
// dataBuf is directly returned from pool and should be returned by consumer.
func (is *IndexScannerImpl) parseSymbolMoniker(symbolData []byte) ([]byte, error) {
	r := bytes.NewReader(symbolData)
	for {
		fieldNumber, fieldType, err := is.consumeBytesFieldMeta(r)
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to consume field"), err)
		}
		if fieldNumber != symbolInformationSymbolField {
			err := is.skipField(r)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("failed to skip field %d in SymbolInfo", fieldNumber), err)
			}
			continue
		}
		dataLen, err := is.consumeLen(r, fieldType)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to consume length"), err)
		}
		dataBuf := is.Pool.Get(dataLen)
		err = is.consumeFieldData(r, dataLen, dataBuf)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to consume field %d data in SymbolInfo", fieldNumber), err)
		}
		// Return immediately after reading the symbol string
		return dataBuf, nil
	}
}

// parseOccurrenceSymbol reads the symbol moniker from an occurrence blob
// dataBuf is directly returned from pool and should be returned by consumer.
func (is *IndexScannerImpl) parseOccurrenceSymbol(occData []byte) ([]byte, error) {
	r := bytes.NewReader(occData)
	for {
		fieldNumber, fieldType, err := is.consumeBytesFieldMeta(r)
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to consume field"), err)
		}
		if fieldNumber != occurrenceSymbolField {
			err := is.skipField(r)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("failed to skip field %d in Occurrence", fieldNumber), err)
			}
			continue
		}
		dataLen, err := is.consumeLen(r, fieldType)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to consume length"), err)
		}
		dataBuf := is.Pool.Get(dataLen)
		err = is.consumeFieldData(r, dataLen, dataBuf)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to consume field %d data in Occurrence", fieldNumber), err)
		}
		// Return immediately after reading the symbol string
		return dataBuf, nil
	}
}

// shouldParseDocumentField returns true if a field in a document should be parsed
func (is *IndexScannerImpl) shouldParseDocumentField(fieldNumber protowire.Number) bool {
	switch fieldNumber {
	case documentSymbolsFieldNumber:
		return is.MatchSymbol != nil
	case documentOccurrencesFieldNumber:
		return is.MatchOccurrence != nil
	}
	return false
}

// consumeBytesFieldMeta reads a field from the reader and returns the field number and type.
func (is *IndexScannerImpl) consumeBytesFieldMeta(r ScipReader) (protowire.Number, protowire.Type, error) {
	tagBuf := is.Pool.Get(1)
	defer is.Pool.Put(tagBuf)
	numRead, err := r.Read(tagBuf)
	if err == io.EOF {
		return 0, 0, io.EOF
	}
	if err != nil {
		return 0, 0, errors.Join(fmt.Errorf("failed to read from index reader"), err)
	}
	if numRead == 0 {
		return 0, 0, errors.New("read 0 bytes from index")
	}
	fieldNumber, fieldType, errCode := protowire.ConsumeTag(tagBuf)
	if errCode < 0 {
		return 0, 0, errors.Join(fmt.Errorf("failed to consume tag"), protowire.ParseError(errCode))
	}
	if fieldType != protowire.BytesType {
		return 0, 0, fmt.Errorf("expected bytes type for field %d", fieldNumber)
	}
	return fieldNumber, fieldType, nil
}

// skipField skips over a field in the reader
func (is *IndexScannerImpl) skipField(r ScipReader) error {
	lenBuf := is.Pool.Get(maxVarintBytes)[:0]
	dataLen, err := readVarint(r, &lenBuf)
	is.Pool.Put(lenBuf)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to read length"), err)
	}
	_, err = r.Seek(int64(dataLen), io.SeekCurrent)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to skip field"), err)
	}

	return nil
}

// consumeLen reads the length of a field from the reader
func (is *IndexScannerImpl) consumeLen(r ScipReader, fieldType protowire.Type) (int, error) {
	if fieldType != protowire.BytesType {
		return 0, errors.New("expected LEN type tag")
	}
	lenBuf := is.Pool.Get(maxVarintBytes)[:0]
	dataLen, err := readVarint(r, &lenBuf)
	is.Pool.Put(lenBuf)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("failed to read length"), err)
	}
	return int(dataLen), err
}

// readAndParseField reads a field from the reader and parses it.
func (is *IndexScannerImpl) consumeFieldData(r ScipReader, dataLen int, dataBuf []byte) error {
	if dataLen > 0 {
		numRead, err := r.Read(dataBuf)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to read data"), err)
		}
		if numRead != dataLen {
			return fmt.Errorf("expected to read %d bytes based on LEN but read %d bytes", dataLen, numRead)
		}
	}
	return nil
}

// readVarint attempts to read a varint, using scratchBuf for temporary storage
// scratchBuf should be able to accommodate any varint size
// based on its capacity, and be cleared before readVarint is called.
// Varints < 128 fit in 1 byte, which means 4 bits are available for field
// numbers. The SCIP types have less than 15 fields, so the tag will fit in 1 byte.
func readVarint(r ScipReader, scratchBuf *[]byte) (uint64, error) {
	nextByteBuf := make([]byte, 1, 1)
	for i := 0; i < cap(*scratchBuf); i++ {
		numRead, err := r.Read(nextByteBuf)
		if err != nil {
			return 0, errors.Join(fmt.Errorf("failed to read %d-th byte of Varint", i), err)
		}
		if numRead == 0 {
			return 0, fmt.Errorf("failed to read %d-th byte of Varint", i)
		}
		nextByte := nextByteBuf[0]
		*scratchBuf = append(*scratchBuf, nextByte)
		if nextByte <= 127 { // https://protobuf.dev/programming-guides/encoding/#varints
			// Continuation bit is not set, so Varint must've ended
			break
		}
	}
	value, errCode := protowire.ConsumeVarint(*scratchBuf)
	if errCode < 0 {
		return value, protowire.ParseError(errCode)
	}
	return value, nil
}

func indexFieldName(i protowire.Number) string {
	if i == indexMetadataFieldNumber {
		return "metadata"
	} else if i == indexDocumentsFieldNumber {
		return "documents"
	} else if i == indexExternalSymbolsFieldNumber {
		return "external_symbols"
	}
	return "<unknown>"
}

func documentFieldName(i protowire.Number) string {
	if i == documentRelativePathFieldNumber {
		return "relative_path"
	} else if i == documentOccurrencesFieldNumber {
		return "occurrences"
	} else if i == documentSymbolsFieldNumber {
		return "symbols"
	} else if i == documentLanguageFieldNumber {
		return "language"
	}
	return "<unknown>"
}

func metadataFieldName(i protowire.Number) string {
	if i == metadataProjectRootFieldNumber {
		return "project_root"
	}
	return "<unknown>"
}
