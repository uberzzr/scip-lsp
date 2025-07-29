package mapper

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"go.lsp.dev/protocol"
)

func TestPluginInfoToRuntimePrioritizedMethods(t *testing.T) {

	methodExamples := make([]*ulspplugin.Methods, 3)
	for i := 0; i < 3; i++ {
		methodExamples[i] = &ulspplugin.Methods{
			PluginNameKey: fmt.Sprintf("test-plugin-%v", i),
			DidOpen: func(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
				return nil
			},
		}
	}

	t.Run("valid plugins", func(t *testing.T) {
		allPluginInfo := []ulspplugin.PluginInfo{
			{
				Priorities: map[string]ulspplugin.Priority{
					protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityRegular,
				},
				Methods: methodExamples[0],
				NameKey: methodExamples[0].PluginNameKey,
			},
			{
				Priorities: map[string]ulspplugin.Priority{
					protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityAsync,
				},
				Methods: methodExamples[1],
				NameKey: methodExamples[1].PluginNameKey,
			},
			{
				Priorities: map[string]ulspplugin.Priority{
					protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityHigh,
				},
				Methods: methodExamples[2],
				NameKey: methodExamples[2].PluginNameKey,
			},
		}

		result, err := PluginInfoToRuntimePrioritizedMethods(allPluginInfo)
		assert.NoError(t, err)

		// Ensure that plugins are split between sync and async, and sorted in correct order.
		assert.Equal(t, 2, len(result[protocol.MethodTextDocumentDidOpen].Sync))
		assert.Equal(t, 1, len(result[protocol.MethodTextDocumentDidOpen].Async))
		sort.Slice(allPluginInfo, func(p, q int) bool {
			return allPluginInfo[p].Priorities[protocol.MethodTextDocumentDidOpen] < allPluginInfo[q].Priorities[protocol.MethodTextDocumentDidOpen]
		})
		for i := 0; i < len(allPluginInfo)-len(result[protocol.MethodTextDocumentDidOpen].Async); i++ {
			assert.Equal(t, result[protocol.MethodTextDocumentDidOpen].Sync[i], allPluginInfo[i].Methods)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		allPluginInfo := []ulspplugin.PluginInfo{
			{
				Priorities: map[string]ulspplugin.Priority{
					protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityRegular,
				},
			},
			{
				Priorities: map[string]ulspplugin.Priority{
					protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityAsync,
				},
				Methods: methodExamples[1],
			},
			{
				Priorities: map[string]ulspplugin.Priority{
					protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityHigh,
				},
				Methods: methodExamples[2],
			},
		}

		_, err := PluginInfoToRuntimePrioritizedMethods(allPluginInfo)
		assert.Error(t, err)
	})
}
