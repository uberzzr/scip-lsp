package notifier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/gateway/ide-client/ideclientmock"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNewNotificationManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)
	logger := zap.NewNop().Sugar()

	assert.NotPanics(t, func() {
		NewNotificationManager(NotificationManagerParams{
			Sessions:   sessionMock,
			IdeGateway: ideClientMock,
			Logger:     logger,
		})
	})
}

func TestStartNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)
	logger := zap.NewNop().Sugar()
	ctx := context.Background()

	m := notificationManagerImpl{
		sessions:   sessionMock,
		ideGateway: ideClientMock,
		logger:     logger,
		managers:   make(map[string]NotificationHandler),
	}

	sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	calls := []struct {
		workspaceRoot   string
		title           string
		expectedEntries int
	}{
		// Adding a new combination of workspaceRoot and title should increase the number of expected entries.
		{"sampleWorkspace", "sample title", 1},
		{"sampleWorkspace", "sample title", 1},
		{"anotherWorkspace", "sample title", 2},
		{"sampleWorkspace", "another title", 3},
		{"sampleWorkspace", "another title", 3},
		{"anotherWorkspace", "sample title", 3},
	}

	for _, call := range calls {
		h, err := m.StartNotification(ctx, call.workspaceRoot, call.title)
		defer h.Done(ctx)
		assert.Len(t, m.managers, call.expectedEntries)
		assert.NoError(t, err)
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)
	logger := zap.NewNop().Sugar()

	t.Run("no remaining senders", func(t *testing.T) {
		m := notificationManagerImpl{
			sessions:   sessionMock,
			ideGateway: ideClientMock,
			logger:     logger,
			managers:   make(map[string]NotificationHandler),
		}

		m.managers["sampleWorkspace-sample title1"] = &notificationHandlerImpl{}
		m.managers["sampleWorkspace-sample title2"] = &notificationHandlerImpl{}
		m.Delete("sampleWorkspace-sample title1")
		assert.Len(t, m.managers, 1)
		m.Delete("sampleWorkspace-sample title2")
		assert.Len(t, m.managers, 0)
	})

	t.Run("still in use", func(t *testing.T) {
		m := notificationManagerImpl{
			sessions:   sessionMock,
			ideGateway: ideClientMock,
			logger:     logger,
			managers:   make(map[string]NotificationHandler),
		}

		m.managers["sampleWorkspace-sample title1"] = &notificationHandlerImpl{
			senderCount: 1,
		}
		m.managers["sampleWorkspace-sample title2"] = &notificationHandlerImpl{
			senderCount: 1,
		}
		m.Delete("sampleWorkspace-sample title1")
		assert.Len(t, m.managers, 2)
		m.Delete("sampleWorkspace-sample title2")
		assert.Len(t, m.managers, 2)
	})
}
