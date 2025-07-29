package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScipPackageID(t *testing.T) {
	pkg := &ScipPackage{
		Manager: "TestManager",
		Name:    "TestName",
		Version: "TestVersion",
	}

	expectedID := "TestManager TestName TestVersion"
	actualID := pkg.ID()

	if actualID != expectedID {
		t.Errorf("Expected ID to be %s, but got %s", expectedID, actualID)
	}
}

func TestParseScipSymbol(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantPackage  string
		wantVersion  string
		wantErr      bool
		checkPackage bool
	}{
		{
			name:         "valid symbol with package",
			input:        "semanticdb maven . . org/apache/commons/collections4/Bag#getCount().",
			wantPackage:  "org/apache/commons/collections4",
			wantVersion:  ".",
			wantErr:      false,
			checkPackage: true,
		},
		{
			name:         "valid symbol with package and version",
			input:        "semanticdb maven maven/./. org_mockito_mockito_core-4.5.1-ijar org/mockito/Mockito#when().",
			wantPackage:  "maven/./.",
			wantVersion:  "org_mockito_mockito_core-4.5.1-ijar",
			wantErr:      false,
			checkPackage: true,
		},
		{
			name:         "symbol with unset package name",
			input:        "semanticdb maven . . namespace1/namespace2/Symbol#",
			wantPackage:  "namespace1/namespace2",
			wantVersion:  ".",
			wantErr:      false,
			checkPackage: true,
		},
		{
			name:         "invalid symbol",
			input:        "invalid",
			wantErr:      true,
			checkPackage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseScipSymbol(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.checkPackage {
				assert.NotNil(t, got.Package)
				assert.Equal(t, tt.wantPackage, got.Package.Name)
				assert.Equal(t, tt.wantVersion, got.Package.Version)
			}
		})
	}
}

func TestParseScipSymbolToDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid symbol with method",
			input:    "semanticdb maven . . org/apache/commons/collections4/Bag#getCount().",
			expected: "getCount",
		},
		{
			name:     "symbol with multiple parts",
			input:    "semanticdb maven . . namespace1/namespace2/Symbol#",
			expected: "Symbol",
		},
		{
			name:     "invalid symbol",
			input:    "invalid",
			expected: "invalid",
		},
		{
			name:     "empty symbol",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseScipSymbolToDisplayName(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
