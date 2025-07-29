package docsync

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestNewPositionMapper(t *testing.T) {
	tests := []struct {
		name         string
		baseText     string
		updatedText  string
		wantModified bool
	}{
		{
			name:         "identical texts",
			baseText:     "Hello\nWorld",
			updatedText:  "Hello\nWorld",
			wantModified: false,
		},
		{
			name:         "different texts",
			baseText:     "Hello\nWorld",
			updatedText:  "Hello\nNew World",
			wantModified: true,
		},
		{
			name:         "empty to non-empty",
			baseText:     "",
			updatedText:  "Hello world",
			wantModified: true,
		},
		{
			name:         "non-empty to empty",
			baseText:     "Hello world",
			updatedText:  "",
			wantModified: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewPositionMapper(tt.baseText, tt.updatedText)
			assert.Equal(t, tt.wantModified, mapper.(*documentPositionMapper).modified)
		})
	}

	t.Run("uninitialized mapper", func(t *testing.T) {
		mapper := &documentPositionMapper{
			modified: true,
		}
		_, _, err := mapper.MapCurrentPositionToBase(protocol.Position{Line: 0, Character: 0})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "position mapper not initialized")
	})
}

func TestMapCurrentPositionToBase(t *testing.T) {
	tests := []struct {
		name            string
		baseText        string
		updatedText     string
		position        protocol.Position
		want            protocol.Position
		wantErrContains string
		wantIsNew       bool
	}{
		{
			name:        "identical texts",
			baseText:    "Hello\nWorld",
			updatedText: "Hello\nWorld",
			position:    protocol.Position{Line: 1, Character: 2},
			want:        protocol.Position{Line: 1, Character: 2},
		},
		{
			name:        "text with deletion - position before deletion",
			baseText:    "Line 1\nLine 2\nLine 3",
			updatedText: "Line 1\nLine 3",
			position:    protocol.Position{Line: 1, Character: 0},
			want:        protocol.Position{Line: 1, Character: 0},
		},
		{
			name:        "text with deletion - position after deletion",
			baseText:    "Line 1\nLine 2\nLine 3",
			updatedText: "Line 1\nLine 3",
			position:    protocol.Position{Line: 1, Character: 5},
			want:        protocol.Position{Line: 2, Character: 5},
		},
		{
			name:        "empty text modifications - empty to non-empty",
			baseText:    "",
			updatedText: "Hello\nWorld",
			position:    protocol.Position{Line: 1, Character: 1},
			want:        protocol.Position{Line: 0, Character: 0},
			wantIsNew:   true,
		},
		{
			name:        "empty text modifications - non-empty to empty",
			baseText:    "Hello\nWorld",
			updatedText: "",
			position:    protocol.Position{Line: 0, Character: 0},
			want:        protocol.Position{Line: 1, Character: 5},
		},
		{
			name:        "large text - position at start",
			baseText:    strings.Repeat("a\n", 1000),
			updatedText: strings.Repeat("b\n", 1000),
			position:    protocol.Position{Line: 0, Character: 0},
			want:        protocol.Position{Line: 0, Character: 0},
			wantIsNew:   true,
		},
		{
			name:        "large text - position near end",
			baseText:    strings.Repeat("a\n", 1000),
			updatedText: strings.Repeat("b\n", 1000),
			position:    protocol.Position{Line: 998, Character: 0},
			want:        protocol.Position{Line: 998, Character: 0},
			wantIsNew:   true,
		},
		{
			name:        "large text with significant differences",
			baseText:    "Line 1\n" + strings.Repeat("Same line\n", 150) + "Last line",
			updatedText: "Line 1\n" + strings.Repeat("Modified line\n", 100) + "Last line",
			position:    protocol.Position{Line: 50, Character: 0},
			want:        protocol.Position{Line: 50, Character: 0},
			wantIsNew:   true,
		},
		{
			name:            "invalid position - beyond end of file",
			baseText:        "Line 1\nLine 2",
			updatedText:     "Line 1\nLine 2\nLine 3",
			position:        protocol.Position{Line: 10, Character: 0},
			wantErrContains: "line number 10 out of range",
		},
		{
			name:            "invalid position - beyond end of large file",
			baseText:        strings.Repeat("a\n", 1000),
			updatedText:     strings.Repeat("b\n", 1000),
			position:        protocol.Position{Line: 2000, Character: 0},
			wantErrContains: "line number 2000 out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewPositionMapper(tt.baseText, tt.updatedText)
			got, isNew, err := mapper.MapCurrentPositionToBase(tt.position)
			if tt.wantErrContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "Mapped possition does not match expected")
			assert.Equal(t, tt.wantIsNew, isNew, "IsNew value does not match expected")
		})
	}
}

func TestMapBasePositionToCurrent(t *testing.T) {
	tests := []struct {
		name            string
		baseText        string
		updatedText     string
		position        protocol.Position
		want            protocol.Position
		wantErrContains string
	}{
		{
			name:        "identical texts",
			baseText:    "Hello\nWorld",
			updatedText: "Hello\nWorld",
			position:    protocol.Position{Line: 1, Character: 2},
			want:        protocol.Position{Line: 1, Character: 2},
		},
		{
			name:        "text with addition - position before addition",
			baseText:    "Line 1\nLine 2\nLine 3",
			updatedText: "Line 1\nNew Line\nLine 2\nLine 3",
			position:    protocol.Position{Line: 1, Character: 0},
			want:        protocol.Position{Line: 2, Character: 0},
		},
		{
			name:        "text with addition - position at addition point",
			baseText:    "Line 1\nLine 2\nLine 3",
			updatedText: "Line 1\nNew Line\nLine 2\nLine 3",
			position:    protocol.Position{Line: 1, Character: 0},
			want:        protocol.Position{Line: 2, Character: 0},
		},
		{
			name:        "text with addition - position after addition",
			baseText:    "Line 1\nLine 2\nLine 3",
			updatedText: "Line 1\nNew Line\nLine 2\nLine 3",
			position:    protocol.Position{Line: 2, Character: 0},
			want:        protocol.Position{Line: 3, Character: 0},
		},
		{
			name:        "unicode text - position before unicode",
			baseText:    "Line 1\n你好\nLine 3",
			updatedText: "Line 1\n你好世界\nLine 3",
			position:    protocol.Position{Line: 1, Character: 0},
			want:        protocol.Position{Line: 1, Character: 0},
		},
		{
			name:        "empty text modifications - empty to non-empty",
			baseText:    "",
			updatedText: "Hello\nWorld",
			position:    protocol.Position{Line: 0, Character: 0},
			want:        protocol.Position{Line: 1, Character: 5},
		},
		{
			name:        "empty text modifications - non-empty to empty",
			baseText:    "Hello\nWorld",
			updatedText: "",
			position:    protocol.Position{Line: 0, Character: 0},
			want:        protocol.Position{Line: 0, Character: 0},
		},
		{
			name:        "large text - position at start",
			baseText:    strings.Repeat("a\n", 1000),
			updatedText: strings.Repeat("b\n", 1000),
			position:    protocol.Position{Line: 0, Character: 0},
			want:        protocol.Position{Line: 0, Character: 0},
		},
		{
			name:        "large text - position near end",
			baseText:    strings.Repeat("a\n", 1000),
			updatedText: strings.Repeat("b\n", 1000),
			position:    protocol.Position{Line: 998, Character: 0},
			want:        protocol.Position{Line: 998, Character: 0},
		},
		{
			name:        "large text with significant differences",
			baseText:    "Line 1\n" + strings.Repeat("Same line\n", 100) + "Last line",
			updatedText: "Line 1\n" + strings.Repeat("Modified line\n", 100) + "Last line",
			position:    protocol.Position{Line: 50, Character: 0},
			want:        protocol.Position{Line: 50, Character: 0},
		},
		{
			name:            "invalid position - beyond end of file",
			baseText:        "Line 1\nLine 2",
			updatedText:     "Line 1\nLine 2\nLine 3",
			position:        protocol.Position{Line: 10, Character: 0},
			wantErrContains: "line number 10 out of range",
		},
		{
			name:            "invalid position - beyond end of large file",
			baseText:        strings.Repeat("a\n", 1000),
			updatedText:     strings.Repeat("b\n", 1000),
			position:        protocol.Position{Line: 2000, Character: 0},
			wantErrContains: "line number 2000 out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewPositionMapper(tt.baseText, tt.updatedText)
			got, err := mapper.MapBasePositionToCurrent(tt.position)
			if tt.wantErrContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "Mapped position does not match expected")
		})
	}
}
