package scip

import (
	"sync"
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/scip-lib/model"
)

func TestGetSymbolInformation(t *testing.T) {
	tests := []struct {
		name     string
		symbol   *SymbolData
		expected *model.SymbolInformation
	}{
		{
			name: "symbol information exists",
			symbol: &SymbolData{
				mu: &sync.Mutex{},
				Info: &model.SymbolInformation{
					Symbol:        "testSymbol",
					Documentation: []string{"This is a test symbol"},
					Kind:          scip.SymbolInformation_Function,
				},
			},
			expected: &model.SymbolInformation{
				Symbol:        "testSymbol",
				Documentation: []string{"This is a test symbol"},
				Kind:          scip.SymbolInformation_Function,
			},
		},
		{
			name: "symbol information is nil",
			symbol: &SymbolData{
				mu:   &sync.Mutex{},
				Info: nil,
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.symbol.GetSymbolInformation()
			assert.Equal(t, test.expected, actual)
		})
	}
}
