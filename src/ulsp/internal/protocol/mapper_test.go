// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestPositionOffset(t *testing.T) {

	tests := []struct {
		name    string
		text    string
		pos     protocol.Position
		offset  int
		wantErr bool
	}{
		{
			name: "valid example",
			text: "sample\ncontent\n",
			pos: protocol.Position{
				Line:      1,
				Character: 0,
			},
			offset:  7,
			wantErr: false,
		},
		{
			name: "invalid line number",
			text: "sample\ncontent\n",
			pos: protocol.Position{
				Line:      15,
				Character: 0,
			},
			wantErr: true,
		},
		{
			name: "end of file, valid",
			text: "sample\ncontent\n",
			pos: protocol.Position{
				Line:      3,
				Character: 0,
			},
			wantErr: false,
			offset:  15,
		},
		{
			name: "end of file, invalid",
			text: "sample\ncontent\n",
			pos: protocol.Position{
				Line:      3,
				Character: 1,
			},
			wantErr: true,
		},
		{
			name: "column is beyond end of line",
			text: "sample\ncontent\n",
			pos: protocol.Position{
				Line:      1,
				Character: 15,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewTextOffsetMapper([]byte(tt.text))
			result, err := m.PositionOffset(tt.pos)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.offset, result)
				assert.NoError(t, err)
			}
		})
	}

}

// Test cases for OffsetPosition conversion.
// Source: https://github.com/golang/tools/blob/67ba59975e7842e66f70fd87113c9d06ae0f8e0f/gopls/internal/lsp/protocol/mapper_test.go

type testCase struct {
	content            string      // input text
	substrOrOffset     interface{} // explicit integer offset, or a substring
	wantLine, wantChar int         // expected LSP position information
}

var tests = []testCase{
	{"a𐐀b", "a", 0, 0},
	{"a𐐀b", "𐐀", 0, 1},
	{"a𐐀b", "b", 0, 3},
	{"a𐐀b\n", "\n", 0, 4},
	{"a𐐀b\r\n", "\n", 0, 4}, // \r|\n is not a valid position, so we move back to the end of the first line.
	{"a𐐀b\r\nx", "x", 1, 0},
	{"a𐐀b\r\nx\ny", "y", 2, 0},

	// Testing EOL and EOF positions
	{"", 0, 0, 0}, // 0th position of an empty buffer is (0, 0)
	{"abc", "c", 0, 2},
	{"abc", 3, 0, 3},
	{"abc\n", "\n", 0, 3},
	{"abc\n", 4, 1, 0}, // position after a newline is on the next line
}

// offset returns the test case byte offset
func (c testCase) offset() int {
	switch x := c.substrOrOffset.(type) {
	case int:
		return x
	case string:
		i := strings.Index(c.content, x)
		if i < 0 {
			panic(fmt.Sprintf("%q does not contain substring %q", c.content, x))
		}
		return i
	}
	panic("substrOrIndex must be an integer or string")
}

func TestLineChar(t *testing.T) {
	for _, test := range tests {
		m := NewTextOffsetMapper([]byte(test.content))
		offset := test.offset()
		posn, _ := m.OffsetPosition(offset)
		gotLine, gotChar := int(posn.Line), int(posn.Character)
		if gotLine != test.wantLine || gotChar != test.wantChar {
			t.Errorf("LineChar(%d) = (%d,%d), want (%d,%d)", offset, gotLine, gotChar, test.wantLine, test.wantChar)
		}
	}
}

func TestInvalidOffset(t *testing.T) {
	content := []byte("a𐐀b\r\nx\ny")
	m := NewTextOffsetMapper(content)
	for _, offset := range []int{-1, 100} {
		_, err := m.OffsetPosition(offset)
		if err == nil {
			t.Errorf("OffsetPosition(%d), want error", offset)
		}
	}
}

func TestPosition(t *testing.T) {
	for _, test := range tests {
		m := NewTextOffsetMapper([]byte(test.content))
		offset := test.offset()
		got, err := m.OffsetPosition(offset)
		if err != nil {
			t.Errorf("OffsetPosition(%d) failed: %v", offset, err)
			continue
		}
		want := protocol.Position{Line: uint32(test.wantLine), Character: uint32(test.wantChar)}
		if got != want {
			t.Errorf("Position(%d) = %v, want %v", offset, got, want)
		}
	}
}
