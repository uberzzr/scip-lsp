package errors

import (
	"fmt"

	"go.lsp.dev/protocol"
)

// DocumentNotFoundError indicates that a document is not found.
type DocumentNotFoundError struct {
	Document protocol.TextDocumentIdentifier
}

// Error is an implementation of the error interface.
func (n *DocumentNotFoundError) Error() string {
	return fmt.Sprintf("Document %q not found", n.Document.URI)
}

// DocumentSizeLimitError indicates that has exceeded the specified size limit
type DocumentSizeLimitError struct {
	Size int64
}

// Error is an implementation of the error interface.
func (n *DocumentSizeLimitError) Error() string {
	return fmt.Sprintf("size of %d bytes exceeds permitted limit", n.Size)
}

// DocumentOutdatedError indicates that a document is outdated.
type DocumentOutdatedError struct {
	CurrentDocument  protocol.TextDocumentItem
	OutdatedDocument protocol.TextDocumentItem
}

// Error is an implementation of the error interface.
func (n *DocumentOutdatedError) Error() string {
	return fmt.Sprintf("document %q version is outdated.  Current version: %v, Outdated version: %v", n.CurrentDocument.URI, n.CurrentDocument.Version, n.OutdatedDocument.Version)
}

// DocumentLanguageIDError indicates that a document's language ID does not match any of the expected values.
type DocumentLanguageIDError struct {
	Document            protocol.TextDocumentItem
	ExpectedLanguageIDs []protocol.LanguageIdentifier
}

// Error is an implementation of the error interface.
func (n *DocumentLanguageIDError) Error() string {
	return fmt.Sprintf("unexpected document type for %q.  expected one of %q, found %q", n.Document.URI, n.ExpectedLanguageIDs, n.Document.LanguageID)
}
