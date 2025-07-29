package notifier

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/uber/scip-lsp/src/ulsp/entity"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	ideclient "github.com/uber/scip-lsp/src/ulsp/gateway/ide-client"
	"github.com/uber/scip-lsp/src/ulsp/repository/session"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

// Channel buffer size for the number of notifications, to avoid blocking the sender.
// In current usage, each thread posts about 10 updates to the notifier, so this accommodates a backup of ~2 runs worth of notifications.
// As they are processed in real time, this should be sufficient to avoid blocking the sender, but can be increased if necessary.
const _bufferSize = 20

// Notification represents a single update reported via the channel to this notifier.  These will be consolidated and displayed to the user.
type Notification struct {
	// SenderToken identifies the sending thread that is reporting this update.
	SenderToken string
	// IdentifierToken serves as a unique identifier to group updates together. Multiple threads may report to the same IdentifierToken.
	IdentifierToken string
	// Priority will be be used to select which update to show if multiple updates are received for the same IdentifierToken.
	Priority int
	// Message represents the text of this update.
	Message string
}

// NotificationHandler receives notifications, consolidates them, and manages a single progress notification.
// Each time a Notification is received on the channel, it will be used to produce an updated ProgressNotification displayed to the user.
// The notification will remain active until all callers have called Done() on the handler.
type NotificationHandler interface {
	Channel() chan Notification
	Add(ctx context.Context)
	Done(ctx context.Context)
	IsClosed() bool
}

type notificationHandlerImpl struct {
	parentManager NotificationManager

	sessions   session.Repository
	ideGateway ideclient.Gateway
	logger     *zap.SugaredLogger

	workspaceRoot         string
	id                    string
	isClosed              bool
	token                 *protocol.ProgressToken
	notificationsBySender map[string]map[string]Notification

	channel chan Notification
	closeCh chan bool

	progressInfoMu sync.Mutex
	senderMu       sync.Mutex
	senderCount    int
	senderWg       sync.WaitGroup
	handlerWg      sync.WaitGroup
}

type notificationHandlerParams struct {
	ParentManager NotificationManager
	Sessions      session.Repository
	IdeGateway    ideclient.Gateway
	Logger        *zap.SugaredLogger
	WorkspaceRoot string
	Title         string
}

// NewNotificationHandler creates a new notification handler, and begins handling updates.
func NewNotificationHandler(ctx context.Context, p notificationHandlerParams, notificationID string) (NotificationHandler, error) {
	h := &notificationHandlerImpl{
		parentManager: p.ParentManager,
		workspaceRoot: p.WorkspaceRoot,
		sessions:      p.Sessions,
		ideGateway:    p.IdeGateway,
		logger:        p.Logger,
		senderCount:   1,

		token:                 protocol.NewProgressToken(factory.UUID().String()),
		channel:               make(chan Notification, _bufferSize),
		closeCh:               make(chan bool),
		id:                    notificationID,
		notificationsBySender: make(map[string]map[string]Notification),
	}
	h.senderWg.Add(1)

	sessions, err := h.sessions.GetAllFromWorkspaceRoot(ctx, h.workspaceRoot)
	if err != nil {
		return nil, err
	}
	for _, session := range sessions {
		h.initNotification(ctx, session, p.Title)
	}

	h.handleUpdates()
	return h, nil
}

// IsClosed returns true if the handler has no remaining senders.
// When a handler reaches 0 senders, it will be cleaned up and is no longer usable.
func (h *notificationHandlerImpl) IsClosed() bool {
	h.senderMu.Lock()
	defer h.senderMu.Unlock()
	return h.senderCount <= 0
}

// Add adds an additional caller of this handler. When all callers have called Done(), the handler will be cleaned up.
func (h *notificationHandlerImpl) Add(ctx context.Context) {
	h.senderMu.Lock()
	defer h.senderMu.Unlock()

	if h.senderCount <= 0 {
		// If the sender count is zero, the handler is already closed.
		return
	}
	h.senderCount++
}

// Close removes one user of the notification handler. Cleanup will be triggered if this is the last caller.
func (h *notificationHandlerImpl) Done(ctx context.Context) {
	h.senderMu.Lock()
	defer h.senderMu.Unlock()
	h.senderCount--

	if h.senderCount == 0 {
		h.senderWg.Done()
	} else if h.senderCount < 0 {
		h.logger.Warnf("Done() called %v extra times on notification handler %s", h.senderCount*-1, h.id)
	}
}

// Channel returns the channel that can be used to send updates.
func (h *notificationHandlerImpl) Channel() chan Notification {
	return h.channel
}

// handleUpdates receives inbound notifications and broadcasts them to all IDE sessions with the same workspace root.
// It also includes a separate goroutine which will begin cleanup once there are no remaining senders.
func (h *notificationHandlerImpl) handleUpdates() {
	ctx := context.Background()
	h.handlerWg.Add(1)
	go func() {
		defer h.handlerWg.Done()
		for {
			select {
			case update, ok := <-h.channel:
				if ok {
					h.broadcastUpdate(ctx, update)
				}
			case <-h.closeCh:
				for update := range h.channel {
					// Drain remaining updates on close.
					h.broadcastUpdate(ctx, update)
				}
				close(h.closeCh)
				return
			}
		}
	}()

	h.handlerWg.Add(1)
	go func() {
		defer h.handlerWg.Done()
		h.senderWg.Wait()
		h.cleanup(ctx)
	}()
}

// cleanup waits for all users to be done with the notification handler, and then cleans up the handler.
func (h *notificationHandlerImpl) cleanup(ctx context.Context) {
	if h.isClosed {
		return
	}

	h.parentManager.Delete(h.id)
	h.closeCh <- true
	h.isClosed = true
	close(h.channel)

	sessions, err := h.sessions.GetAllFromWorkspaceRoot(ctx, h.workspaceRoot)
	if err != nil {
		return
	}

	for _, session := range sessions {
		h.endNotification(ctx, session)
	}
}

// broadcastUpdate sends the latest update to all sessions.
func (h *notificationHandlerImpl) broadcastUpdate(ctx context.Context, update Notification) {
	latestUpdate := h.captureUpdate(update)
	sessions, err := h.sessions.GetAllFromWorkspaceRoot(ctx, h.workspaceRoot)
	if err != nil {
		h.logger.Errorf("posting progress update: %s", err)
		return
	}
	for _, session := range sessions {
		err := h.updateNotification(ctx, session, latestUpdate, 0)
		if err != nil {
			h.logger.Errorf("updating progress for session %s: %s", session.UUID, err)
		}
	}
}

// captureUpdate captures the latest update for a given thread token, and returns the consolidated message.
func (h *notificationHandlerImpl) captureUpdate(update Notification) string {
	h.progressInfoMu.Lock()
	defer h.progressInfoMu.Unlock()

	if len(update.Message) == 0 {
		// Empty update indicates removal, otherwise capture the latest update.
		if _, ok := h.notificationsBySender[update.SenderToken]; ok {
			delete(h.notificationsBySender[update.SenderToken], update.IdentifierToken)
			if len(h.notificationsBySender[update.SenderToken]) == 0 {
				delete(h.notificationsBySender, update.SenderToken)
			}
		}
	} else {
		if _, ok := h.notificationsBySender[update.SenderToken]; !ok {
			h.notificationsBySender[update.SenderToken] = make(map[string]Notification)
		}

		h.notificationsBySender[update.SenderToken][update.IdentifierToken] = update
	}

	// Select the message with the lowest priority for each IdentifierToken.
	// The notifications channel is buffered to avoid blocking the sender while this logic runs.
	condensedByPriority := make(map[string]Notification)
	for _, sender := range h.notificationsBySender {
		for _, msg := range sender {
			existing, ok := condensedByPriority[msg.IdentifierToken]
			// Store the lowest priority entry with this key.
			if !ok || existing.Priority > msg.Priority {
				condensedByPriority[msg.IdentifierToken] = msg
			}
		}
	}

	// Condense the de-duplicated results into a final result.
	i := 0
	combinedMessage := make([]string, len(condensedByPriority))
	for _, msg := range condensedByPriority {
		combinedMessage[i] = msg.Message
		i++
	}

	if len(combinedMessage) == 1 {
		return combinedMessage[0]
	}

	// Output example: [message1, message2, message3]
	slices.Sort(combinedMessage)

	return fmt.Sprintf("[%s]", strings.Join(combinedMessage, ", "))
}

// The methods below provide simple wrappers for IDE Gateway calls.

func (h *notificationHandlerImpl) initNotification(ctx context.Context, session *entity.Session, title string) error {
	sCtx := context.WithValue(ctx, entity.SessionContextKey, session.UUID)
	err := h.ideGateway.WorkDoneProgressCreate(sCtx, &protocol.WorkDoneProgressCreateParams{
		Token: *h.token,
	})
	if err != nil {
		return err
	}

	return h.ideGateway.Progress(sCtx, &protocol.ProgressParams{
		Token: *h.token,
		Value: &protocol.WorkDoneProgressBegin{
			Kind:       protocol.WorkDoneProgressKindBegin,
			Title:      title,
			Message:    "",
			Percentage: 0,
		},
	})
}

func (h *notificationHandlerImpl) updateNotification(ctx context.Context, session *entity.Session, progress string, percentage uint32) error {
	sCtx := context.WithValue(ctx, entity.SessionContextKey, session.UUID)
	return h.ideGateway.Progress(sCtx, &protocol.ProgressParams{
		Token: *h.token,
		Value: &protocol.WorkDoneProgressReport{
			Kind:       protocol.WorkDoneProgressKindReport,
			Message:    progress,
			Percentage: percentage,
		},
	})
}

func (h *notificationHandlerImpl) endNotification(ctx context.Context, session *entity.Session) error {
	sCtx := context.WithValue(ctx, entity.SessionContextKey, session.UUID)
	err := h.ideGateway.Progress(sCtx, &protocol.ProgressParams{
		Token: *h.token,
		Value: &protocol.WorkDoneProgressEnd{
			Kind: protocol.WorkDoneProgressKindEnd,
		},
	})
	return err
}
