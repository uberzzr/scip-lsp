package ulspdaemon

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCodeIntelMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	c := controller{
		logger:        logger.Sugar(),
		pluginMethods: sampleCodeIntelMethods(s.UUID),
		sessions:      sessionRepository,
	}

	t.Run("CodeAction", func(t *testing.T) {
		params := &protocol.CodeActionParams{}
		result, err := c.CodeAction(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("CodeLens", func(t *testing.T) {
		params := &protocol.CodeLensParams{}
		result, err := c.CodeLens(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("CodeLensRefresh", func(t *testing.T) {
		err := c.CodeLensRefresh(ctx)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("CodeLensResolve", func(t *testing.T) {
		params := &protocol.CodeLens{}
		result, err := c.CodeLensResolve(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.NotNil(t, result.Range)
	})

	t.Run("GotoDeclaration", func(t *testing.T) {
		params := &protocol.DeclarationParams{}
		result, err := c.GotoDeclaration(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("GotoDefinition", func(t *testing.T) {
		params := &protocol.DefinitionParams{}
		result, err := c.GotoDefinition(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("GotoTypeDefinition", func(t *testing.T) {
		params := &protocol.TypeDefinitionParams{}
		result, err := c.GotoTypeDefinition(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("GotoImplementation", func(t *testing.T) {
		params := &protocol.ImplementationParams{}
		result, err := c.GotoImplementation(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("References", func(t *testing.T) {
		params := &protocol.ReferenceParams{}
		result, err := c.References(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("Hover", func(t *testing.T) {
		params := &protocol.HoverParams{}
		result, err := c.Hover(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.NotNil(t, result)
	})

	t.Run("DocumentSymbol", func(t *testing.T) {
		params := &protocol.DocumentSymbolParams{}
		result, err := c.DocumentSymbol(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})
}

// sampleCodeIntelMethods a sample of RuntimePrioritizedMethods to be used for testing.
// For each method, simulates two assigned plugins: the first returns nil and the second returns an error.
func sampleCodeIntelMethods(id uuid.UUID) map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods {
	err := errors.New("sample")
	m := []*ulspplugin.Methods{
		{
			CodeAction: func(ctx context.Context, params *protocol.CodeActionParams, result *[]protocol.CodeAction) error {
				if result != nil {
					*result = append(*result, protocol.CodeAction{}, protocol.CodeAction{}, protocol.CodeAction{})
				}
				return nil
			},
			CodeLens: func(ctx context.Context, params *protocol.CodeLensParams, result *[]protocol.CodeLens) error {
				if result != nil {
					*result = append(*result, protocol.CodeLens{}, protocol.CodeLens{}, protocol.CodeLens{})
				}
				return nil
			},
			CodeLensRefresh: func(ctx context.Context) error {
				return nil
			},
			CodeLensResolve: func(ctx context.Context, params *protocol.CodeLens, result *protocol.CodeLens) error {
				if result != nil {
					result.Range = protocol.Range{}
				}
				return nil
			},
			GotoDeclaration: func(ctx context.Context, params *protocol.DeclarationParams, result *[]protocol.LocationLink) error {
				if result != nil {
					*result = append(*result, protocol.LocationLink{}, protocol.LocationLink{}, protocol.LocationLink{})
				}
				return nil
			},
			GotoDefinition: func(ctx context.Context, params *protocol.DefinitionParams, result *[]protocol.LocationLink) error {
				if result != nil {
					*result = append(*result, protocol.LocationLink{}, protocol.LocationLink{}, protocol.LocationLink{})
				}
				return nil
			}, GotoTypeDefinition: func(ctx context.Context, params *protocol.TypeDefinitionParams, result *[]protocol.LocationLink) error {
				if result != nil {
					*result = append(*result, protocol.LocationLink{}, protocol.LocationLink{}, protocol.LocationLink{})
				}
				return nil
			},
			GotoImplementation: func(ctx context.Context, params *protocol.ImplementationParams, result *[]protocol.LocationLink) error {
				if result != nil {
					*result = append(*result, protocol.LocationLink{}, protocol.LocationLink{}, protocol.LocationLink{})
				}
				return nil
			},
			References: func(ctx context.Context, params *protocol.ReferenceParams, result *[]protocol.Location) error {
				if result != nil {
					*result = append(*result, protocol.Location{}, protocol.Location{}, protocol.Location{})
				}
				return nil
			},
			Hover: func(ctx context.Context, params *protocol.HoverParams, result *protocol.Hover) error {
				if result != nil {
					result.Contents = protocol.MarkupContent{
						Kind:  protocol.PlainText,
						Value: "sample",
					}
				}
				return nil
			},
			DocumentSymbol: func(ctx context.Context, params *protocol.DocumentSymbolParams, result *[]protocol.DocumentSymbol) error {
				if result != nil {
					*result = append(*result, protocol.DocumentSymbol{}, protocol.DocumentSymbol{}, protocol.DocumentSymbol{})
				}
				return nil
			},
		},
		{
			CodeAction: func(ctx context.Context, params *protocol.CodeActionParams, result *[]protocol.CodeAction) error {
				return err
			},
			CodeLens: func(ctx context.Context, params *protocol.CodeLensParams, result *[]protocol.CodeLens) error {
				return err
			},
			CodeLensRefresh: func(ctx context.Context) error {
				return err
			},
			CodeLensResolve: func(ctx context.Context, params *protocol.CodeLens, result *protocol.CodeLens) error {
				return err
			},
			GotoDeclaration: func(ctx context.Context, params *protocol.DeclarationParams, result *[]protocol.LocationLink) error {
				return err
			},
			GotoDefinition: func(ctx context.Context, params *protocol.DefinitionParams, result *[]protocol.LocationLink) error {
				return err
			},
			GotoTypeDefinition: func(ctx context.Context, params *protocol.TypeDefinitionParams, result *[]protocol.LocationLink) error {
				return err
			},
			GotoImplementation: func(ctx context.Context, params *protocol.ImplementationParams, result *[]protocol.LocationLink) error {
				return err
			},
			References: func(ctx context.Context, params *protocol.ReferenceParams, result *[]protocol.Location) error {
				return err
			},
			Hover: func(ctx context.Context, params *protocol.HoverParams, result *protocol.Hover) error {
				return err
			},
			DocumentSymbol: func(ctx context.Context, params *protocol.DocumentSymbolParams, result *[]protocol.DocumentSymbol) error {
				return err
			},
		},
	}

	methodLists := ulspplugin.MethodLists{
		Sync:  m,
		Async: m,
	}

	result := make(ulspplugin.RuntimePrioritizedMethods)
	for _, val := range []string{
		protocol.MethodTextDocumentCodeAction,
		protocol.MethodTextDocumentCodeLens,
		protocol.MethodCodeLensRefresh,
		protocol.MethodCodeLensResolve,
		protocol.MethodTextDocumentDeclaration,
		protocol.MethodTextDocumentDefinition,
		protocol.MethodTextDocumentTypeDefinition,
		protocol.MethodTextDocumentImplementation,
		protocol.MethodTextDocumentReferences,
		protocol.MethodTextDocumentHover,
		protocol.MethodTextDocumentDocumentSymbol,
	} {
		result[val] = methodLists
	}

	return map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{id: result}

}
