// This file includes a selection of byte offset conversion methods from the gopls "protocol" package.
// Based on the following: https://github.com/golang/tools/blob/67d73b2960c82b2c8db0b9d0694c66a789a1db11/gopls/internal/lsp/protocol/mapper.go

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// License Revision: https://github.com/golang/tools/blob/67d73b2960c82b2c8db0b9d0694c66a789a1db11/LICENSE

package protocol

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"unicode/utf8"

	"go.lsp.dev/protocol"
)

// TextOffsetMapper is used for conversions related to text offsets.
type TextOffsetMapper struct {
	Content []byte

	// Line-number information is requested only for a tiny
	// fraction of Mappers, so we compute it lazily.
	// Call initLines() before accessing fields below.
	linesOnce sync.Once
	lineStart []int // byte offset of start of ith line (0-based); last=EOF iff \n-terminated
	nonASCII  bool

	// TODO(adonovan): adding an extra lineStart entry for EOF
	// might simplify every method that accesses it. Try it out.
}

// NewTextOffsetMapper creates a new mapper for the given URI and content.
func NewTextOffsetMapper(content []byte) *TextOffsetMapper {
	return &TextOffsetMapper{Content: content}
}

// initLines populates the lineStart table.
func (m *TextOffsetMapper) initLines() {
	m.linesOnce.Do(func() {
		nlines := bytes.Count(m.Content, []byte("\n"))
		m.lineStart = make([]int, 1, nlines+1) // initially []int{0}
		for offset, b := range m.Content {
			if b == '\n' {
				m.lineStart = append(m.lineStart, offset+1)
			}
			if b >= utf8.RuneSelf {
				m.nonASCII = true
			}
		}
	})
}

// PositionOffset converts a protocol (UTF-16) position to a byte offset.
func (m *TextOffsetMapper) PositionOffset(p protocol.Position) (int, error) {
	m.initLines()

	// Validate line number.
	if p.Line > uint32(len(m.lineStart)) {
		return 0, fmt.Errorf("line number %d out of range 0-%d", p.Line, len(m.lineStart))
	} else if p.Line == uint32(len(m.lineStart)) {
		if p.Character == 0 {
			return len(m.Content), nil // EOF
		}
		return 0, fmt.Errorf("column is beyond end of file")
	}

	offset := m.lineStart[p.Line]
	content := m.Content[offset:] // rest of file from start of enclosing line

	// Advance bytes up to the required number of UTF-16 codes.
	col8 := 0
	for col16 := 0; col16 < int(p.Character); col16++ {
		r, sz := utf8.DecodeRune(content)
		if sz == 0 {
			return 0, fmt.Errorf("column is beyond end of file")
		}
		if r == '\n' {
			return 0, fmt.Errorf("column is beyond end of line")
		}
		if sz == 1 && r == utf8.RuneError {
			return 0, fmt.Errorf("buffer contains invalid UTF-8 text")
		}
		content = content[sz:]

		if r >= 0x10000 {
			col16++ // rune was encoded by a pair of surrogate UTF-16 codes

			if col16 == int(p.Character) {
				break // requested position is in the middle of a rune
			}
		}
		col8 += sz
	}
	return offset + col8, nil
}

// OffsetPosition converts a byte offset to a protocol (UTF-16) position.
func (m *TextOffsetMapper) OffsetPosition(offset int) (protocol.Position, error) {
	if !(0 <= offset && offset <= len(m.Content)) {
		return protocol.Position{}, fmt.Errorf("invalid offset %d (want 0-%d)", offset, len(m.Content))
	}
	// No error may be returned after this point,
	// even if the offset does not fall at a rune boundary.
	// (See panic in MappedRange.Range reachable.)

	line, col16 := m.lineCol16(offset)
	return protocol.Position{Line: uint32(line), Character: uint32(col16)}, nil
}

// lineCol16 converts a valid byte offset to line and UTF-16 column numbers, both 0-based.
func (m *TextOffsetMapper) lineCol16(offset int) (int, int) {
	line, start, cr := m.line(offset)
	var col16 int
	if m.nonASCII {
		col16 = UTF16Len(m.Content[start:offset])
	} else {
		col16 = offset - start
	}
	if cr {
		col16-- // retreat from \r at line end
	}
	return line, col16
}

// line returns:
// - the 0-based index of the line that encloses the (valid) byte offset;
// - the start offset of that line; and
// - whether the offset denotes a carriage return (\r) at line end.
func (m *TextOffsetMapper) line(offset int) (int, int, bool) {
	m.initLines()
	// In effect, binary search returns a 1-based result.
	line := sort.Search(len(m.lineStart), func(i int) bool {
		return offset < m.lineStart[i]
	})

	// Adjustment for line-endings: \r|\n is the same as |\r\n.
	var eol int
	if line == len(m.lineStart) {
		eol = len(m.Content) // EOF
	} else {
		eol = m.lineStart[line] - 1
	}
	cr := offset == eol && offset > 0 && m.Content[offset-1] == '\r'

	line-- // 0-based

	return line, m.lineStart[line], cr
}

// UTF16Len returns the number of codes in the UTF-16 transcoding of s.
func UTF16Len(s []byte) int {
	var n int
	for len(s) > 0 {
		n++

		// Fast path for ASCII.
		if s[0] < 0x80 {
			s = s[1:]
			continue
		}

		r, size := utf8.DecodeRune(s)
		if r >= 0x10000 {
			n++ // surrogate pair
		}
		s = s[size:]
	}
	return n
}
