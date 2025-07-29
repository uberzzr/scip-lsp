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

func TestDocumentMethods(t *testing.T) {
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
		pluginMethods: sampleDocumentMethods(s.UUID),
		sessions:      sessionRepository,
	}

	t.Run("DidChangeWatchedFiles", func(t *testing.T) {
		params := &protocol.DidChangeWatchedFilesParams{}
		err := c.DidChangeWatchedFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("DidChange", func(t *testing.T) {
		params := &protocol.DidChangeTextDocumentParams{}
		err := c.DidChange(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("DidOpen", func(t *testing.T) {
		params := &protocol.DidOpenTextDocumentParams{}
		err := c.DidOpen(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("DidClose", func(t *testing.T) {
		params := &protocol.DidCloseTextDocumentParams{}
		err := c.DidClose(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("WillSave", func(t *testing.T) {
		params := &protocol.WillSaveTextDocumentParams{}
		err := c.WillSave(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("WillSaveWaitUntil", func(t *testing.T) {
		params := &protocol.WillSaveTextDocumentParams{}
		result, err := c.WillSaveWaitUntil(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result))
	})

	t.Run("DidSave", func(t *testing.T) {
		params := &protocol.DidSaveTextDocumentParams{}
		err := c.DidSave(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("WillRenameFiles", func(t *testing.T) {
		params := &protocol.RenameFilesParams{}
		result, err := c.WillRenameFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result.DocumentChanges))
	})

	t.Run("DidRenameFiles", func(t *testing.T) {
		params := &protocol.RenameFilesParams{}
		err := c.DidRenameFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("WillCreateFiles", func(t *testing.T) {
		params := &protocol.CreateFilesParams{}
		result, err := c.WillCreateFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result.DocumentChanges))
	})

	t.Run("DidCreateFiles", func(t *testing.T) {
		params := &protocol.CreateFilesParams{}
		err := c.DidCreateFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})

	t.Run("WillDeleteFiles", func(t *testing.T) {
		params := &protocol.DeleteFilesParams{}
		result, err := c.WillDeleteFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Equal(t, 3, len(result.DocumentChanges))
	})

	t.Run("DidDeleteFiles", func(t *testing.T) {
		params := &protocol.DeleteFilesParams{}
		err := c.DidDeleteFiles(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})
}

// sampleDocumentMethods a sample of RuntimePrioritizedMethods to be used for testing.
// For each method, simulates two assigned plugins: the first returns nil and the second returns an error.
func sampleDocumentMethods(id uuid.UUID) map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods {
	err := errors.New("sample")
	m := []*ulspplugin.Methods{
		{
			DidChange: func(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
				return nil
			},
			DidChangeWatchedFiles: func(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
				return nil
			},
			DidOpen: func(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
				return nil
			},
			DidClose: func(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
				return nil
			},
			WillSave: func(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
				return nil
			},
			WillSaveWaitUntil: func(ctx context.Context, params *protocol.WillSaveTextDocumentParams, result *[]protocol.TextEdit) error {
				if result != nil {
					*result = append(*result, protocol.TextEdit{}, protocol.TextEdit{}, protocol.TextEdit{})
				}
				return nil
			},
			DidSave: func(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
				return nil
			},
			WillRenameFiles: func(ctx context.Context, params *protocol.RenameFilesParams, result *protocol.WorkspaceEdit) error {
				if result != nil {
					result.DocumentChanges = []protocol.TextDocumentEdit{{}, {}, {}}
				}
				return nil
			},
			DidRenameFiles: func(ctx context.Context, params *protocol.RenameFilesParams) error {
				return nil
			},
			WillCreateFiles: func(ctx context.Context, params *protocol.CreateFilesParams, result *protocol.WorkspaceEdit) error {
				if result != nil {
					result.DocumentChanges = []protocol.TextDocumentEdit{{}, {}, {}}
				}
				return nil
			},
			DidCreateFiles: func(ctx context.Context, params *protocol.CreateFilesParams) error {
				return nil
			},
			WillDeleteFiles: func(ctx context.Context, params *protocol.DeleteFilesParams, result *protocol.WorkspaceEdit) error {
				if result != nil {
					result.DocumentChanges = []protocol.TextDocumentEdit{{}, {}, {}}
				}
				return nil
			},
			DidDeleteFiles: func(ctx context.Context, params *protocol.DeleteFilesParams) error {
				return nil
			},
		},
		{
			DidChange: func(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
				return err
			},
			DidChangeWatchedFiles: func(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
				return err
			},
			DidOpen: func(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
				return err
			},
			DidClose: func(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
				return err
			},
			WillSave: func(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
				return err
			},
			WillSaveWaitUntil: func(ctx context.Context, params *protocol.WillSaveTextDocumentParams, result *[]protocol.TextEdit) error {
				return err
			},
			DidSave: func(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
				return err
			},
			WillRenameFiles: func(ctx context.Context, params *protocol.RenameFilesParams, result *protocol.WorkspaceEdit) error {
				return err
			},
			DidRenameFiles: func(ctx context.Context, params *protocol.RenameFilesParams) error {
				return err
			},
			WillCreateFiles: func(ctx context.Context, params *protocol.CreateFilesParams, result *protocol.WorkspaceEdit) error {
				return err
			},
			DidCreateFiles: func(ctx context.Context, params *protocol.CreateFilesParams) error {
				return err
			},
			WillDeleteFiles: func(ctx context.Context, params *protocol.DeleteFilesParams, result *protocol.WorkspaceEdit) error {
				return err
			},
			DidDeleteFiles: func(ctx context.Context, params *protocol.DeleteFilesParams) error {
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
		protocol.MethodTextDocumentDidChange,
		protocol.MethodWorkspaceDidChangeWatchedFiles,
		protocol.MethodTextDocumentDidOpen,
		protocol.MethodTextDocumentDidClose,
		protocol.MethodTextDocumentWillSave,
		protocol.MethodTextDocumentWillSaveWaitUntil,
		protocol.MethodTextDocumentDidSave,
		protocol.MethodWillRenameFiles,
		protocol.MethodDidRenameFiles,
		protocol.MethodWillCreateFiles,
		protocol.MethodDidCreateFiles,
		protocol.MethodWillDeleteFiles,
		protocol.MethodDidDeleteFiles,
	} {
		result[val] = methodLists
	}

	return map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{id: result}

}
