package scip

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	notifier "github.com/uber/scip-lsp/src/ulsp/internal/persistent-notifier"
)

// IndexLoadStatus is the status of a SCIP index load
type IndexLoadStatus string

const (
	// IndexLoadSuccess is the status of a successful SCIP index load
	IndexLoadSuccess IndexLoadStatus = "done"
	// IndexLoadError is the status of a failed SCIP index load
	IndexLoadError IndexLoadStatus = "error"
)

// IndexNotifier provides a simple way to show SCIP index loading notifications
type IndexNotifier struct {
	mu                  sync.Mutex
	notifiers           notifier.NotificationManager
	activeNotifications map[string]*notificationState
}

// notificationState tracks the state for a workspace notification
type notificationState struct {
	handler   notifier.NotificationHandler
	processed int
	pending   map[string]bool
}

// NewIndexNotifier creates a minimal SCIP index load notifier
func NewIndexNotifier(notifiers notifier.NotificationManager) *IndexNotifier {
	return &IndexNotifier{
		activeNotifications: make(map[string]*notificationState),
		notifiers:           notifiers,
	}
}

// TrackFile adds a file to be tracked for a workspace
func (n *IndexNotifier) TrackFile(ctx context.Context, workspaceRoot, filePath string) error {
	return n.trackFileInternal(ctx, workspaceRoot, filePath, false)
}

// NotifyStart shows a notification when starting to load a SCIP index file
func (n *IndexNotifier) NotifyStart(ctx context.Context, workspaceRoot, filePath string) error {
	return n.trackFileInternal(ctx, workspaceRoot, filePath, true)
}

// NotifyComplete shows a notification when a SCIP index file has been loaded
func (n *IndexNotifier) NotifyComplete(ctx context.Context, workspaceRoot, filePath string, status IndexLoadStatus) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	state, exists := n.activeNotifications[workspaceRoot]
	if !exists {
		return fmt.Errorf("no notification handler found for workspace %s", workspaceRoot)
	}
	state.processed++
	delete(state.pending, filePath)

	if status != IndexLoadSuccess {
		state.handler.Channel() <- newNotification(fmt.Sprintf("Failed to load %s: %s",
			filepath.Base(filePath), status))
	}
	notifyProgress(state, filePath)

	// If no more pending files, clean up
	if len(state.pending) == 0 {
		state.handler.Channel() <- newNotification(fmt.Sprintf("Completed loading indices (%d files)", state.processed))
		delete(n.activeNotifications, workspaceRoot)
		state.handler.Done(ctx)
	}
	return nil
}

func notifyProgress(state *notificationState, filePath string) {
	fileName := filepath.Base(filePath)
	totalFiles := len(state.pending) + state.processed
	progress := 0
	if totalFiles > 0 {
		progress = (state.processed * 100) / totalFiles
	}
	state.handler.Channel() <- newNotification(fmt.Sprintf("%d%% (%s)",
		progress, fileName))
}

func newNotification(text string) notifier.Notification {
	return notifier.Notification{
		Message:         text,
		IdentifierToken: "scip",
		SenderToken:     "scip",
		Priority:        1,
	}
}

// trackFileInternal handles both tracking and notification with a single lock
func (n *IndexNotifier) trackFileInternal(ctx context.Context, workspaceRoot, filePath string, showProgress bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	state, exists := n.activeNotifications[workspaceRoot]
	if !exists {
		handler, err := n.notifiers.StartNotification(ctx, workspaceRoot, "Loading indices")
		if err != nil {
			return err
		}
		state = &notificationState{
			handler:   handler,
			processed: 0,
			pending:   make(map[string]bool),
		}
		n.activeNotifications[workspaceRoot] = state
	}

	state.pending[filePath] = true

	if showProgress {
		notifyProgress(state, filePath)
	}

	return nil
}
