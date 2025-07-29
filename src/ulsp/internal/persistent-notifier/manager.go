package notifier

import (
	"context"
	"fmt"
	"sync"

	ideclient "github.com/uber/scip-lsp/src/ulsp/gateway/ide-client"
	"github.com/uber/scip-lsp/src/ulsp/repository/session"
	"go.uber.org/zap"
)

// NotificationManagerParams are used to initialize a new NotificationManager.
type NotificationManagerParams struct {
	Sessions   session.Repository
	IdeGateway ideclient.Gateway
	Logger     *zap.SugaredLogger
}

// NotificationManager is to manage multiple simultaneous NotificationHandlers.
// StartNotification will start a new NotificationHandler, or get the existing one if present for this workspaceRoot and title combination.
type NotificationManager interface {
	StartNotification(ctx context.Context, workspaceRoot string, title string) (NotificationHandler, error)
	Delete(id string)
}

// NewNotificationManager creates a new NotificationManager.
func NewNotificationManager(p NotificationManagerParams) NotificationManager {
	return &notificationManagerImpl{
		sessions:   p.Sessions,
		ideGateway: p.IdeGateway,
		logger:     p.Logger,

		managers: make(map[string]NotificationHandler),
	}
}

type notificationManagerImpl struct {
	sessions   session.Repository
	ideGateway ideclient.Gateway
	logger     *zap.SugaredLogger

	managers map[string]NotificationHandler
	mu       sync.Mutex
}

// StartNotification starts a new NotificationHandler, or gets the existing one if present for this workspaceRoot and title combination.
// The NotificationHandler will remain active until all callers have called Done() on the handler.
func (m *notificationManagerImpl) StartNotification(ctx context.Context, workspaceRoot string, title string) (NotificationHandler, error) {
	// WorkspaceRoot and title are used to generate a unique identifier for the notification handler.
	notificationID := fmt.Sprintf("%s-%s", workspaceRoot, title)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure that we do not return an old handler that is already closed.
	m.deleteIfClosed(notificationID)
	existing, ok := m.managers[notificationID]
	if ok {
		// Existing handler for this notificationId. Add an additional caller and return it.
		existing.Add(ctx)
		return existing, nil
	}

	// No existing handler for this notificationId. Create a new one, store it, and return it.
	h, err := NewNotificationHandler(ctx, notificationHandlerParams{
		ParentManager: m,
		Sessions:      m.sessions,
		IdeGateway:    m.ideGateway,
		Logger:        m.logger,
		WorkspaceRoot: workspaceRoot,
		Title:         title,
	}, notificationID)
	if err != nil {
		return nil, err
	}
	m.managers[notificationID] = h
	return h, nil
}

// Delete allows a NotificationHandler to remove itself from the NotificationHandler. This should be called by the NotificationHandler before closing its channel.
func (m *notificationManagerImpl) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove only if closed. This allows for flexible ordering of StartNotification and Delete calls.
	m.deleteIfClosed(id)
}

// Removes a notification handler only if it is already closed.
func (m *notificationManagerImpl) deleteIfClosed(id string) {
	if mgr, ok := m.managers[id]; ok {
		if mgr.IsClosed() {
			delete(m.managers, id)
		}
	}
}
