package scip

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	notifier "github.com/uber/scip-lsp/src/ulsp/internal/persistent-notifier"
	"github.com/uber/scip-lsp/src/ulsp/internal/persistent-notifier/notifiermock"
	"go.uber.org/mock/gomock"
)

func TestNewScipNotifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	notifier := NewIndexNotifier(notifiermock.NewMockNotificationManager(ctrl))
	assert.NotNil(t, notifier)
	assert.Empty(t, notifier.activeNotifications)
}

func TestTrackFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	notifManagerMock := notifiermock.NewMockNotificationManager(ctrl)
	notifHandlerMock := notifiermock.NewMockNotificationHandler(ctrl)
	notifManagerMock.EXPECT().StartNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(notifHandlerMock, nil)

	n := NewIndexNotifier(notifManagerMock)

	n.TrackFile(context.Background(), "/workspace", "/workspace/path/to/file.go")
	assert.Equal(t, 1, len(n.activeNotifications["/workspace"].pending))
	assert.True(t, n.activeNotifications["/workspace"].pending["/workspace/path/to/file.go"])
}

func TestNotifyStart(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	workspaceRoot := "/workspace"
	filePath := "/workspace/path/to/file.go"

	t.Run("new notification", func(t *testing.T) {
		notifManagerMock := notifiermock.NewMockNotificationManager(ctrl)
		notifHandlerMock := notifiermock.NewMockNotificationHandler(ctrl)
		notifChannel := make(chan notifier.Notification, 10)

		notifHandlerMock.EXPECT().Channel().Return(notifChannel)
		notifManagerMock.EXPECT().StartNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(notifHandlerMock, nil)

		n := NewIndexNotifier(notifManagerMock)

		err := n.NotifyStart(ctx, workspaceRoot, filePath)
		assert.NoError(t, err)

		// Verify a notification state was created for the workspace
		state, exists := n.activeNotifications[workspaceRoot]
		assert.True(t, exists)
		assert.Equal(t, 0, state.processed)
		assert.True(t, state.pending[filePath])

		// Verify a notification was sent
		notification := <-notifChannel
		assert.Contains(t, notification.Message, "file.go")
		assert.Equal(t, "scip", notification.SenderToken)
		assert.Equal(t, "scip", notification.IdentifierToken)
	})

	t.Run("existing notification", func(t *testing.T) {

		handlerMock := notifiermock.NewMockNotificationHandler(ctrl)
		managerMock := notifiermock.NewMockNotificationManager(ctrl)
		notifChannel := make(chan notifier.Notification, 10)
		handlerMock.EXPECT().Channel().Return(notifChannel)

		n := NewIndexNotifier(managerMock)
		// Pre-populate with an existing notification state
		n.activeNotifications[workspaceRoot] = &notificationState{
			handler:   handlerMock,
			processed: 2,
			pending:   map[string]bool{"/workspace/another/file.go": true},
		}

		err := n.NotifyStart(ctx, workspaceRoot, filePath)
		assert.NoError(t, err)

		// Verify notification was updated
		state := n.activeNotifications[workspaceRoot]
		assert.True(t, state.pending[filePath])
		assert.Equal(t, 2, state.processed)

		notification := <-notifChannel
		assert.Contains(t, notification.Message, "50%")
		assert.Contains(t, notification.Message, "file.go")
	})
}

func TestNotifyComplete(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	workspaceRoot := "/workspace"
	filePath := "/workspace/path/to/file.go"

	t.Run("no notification for workspace", func(t *testing.T) {
		managerMock := notifiermock.NewMockNotificationManager(ctrl)
		n := NewIndexNotifier(managerMock)

		// Should not panic when workspace doesn't exist
		n.NotifyComplete(ctx, workspaceRoot, filePath, "done")
	})

	t.Run("success status", func(t *testing.T) {

		managerMock := notifiermock.NewMockNotificationManager(ctrl)
		handlerMock := notifiermock.NewMockNotificationHandler(ctrl)
		notifChannel := make(chan notifier.Notification, 10)

		handlerMock.EXPECT().Channel().Return(notifChannel)

		n := NewIndexNotifier(managerMock)
		// Set up a notification with pending files
		n.activeNotifications[workspaceRoot] = &notificationState{
			handler:   handlerMock,
			processed: 1,
			pending: map[string]bool{
				filePath:                     true,
				"/workspace/another/file.go": true,
			},
		}

		n.NotifyComplete(ctx, workspaceRoot, filePath, "done")

		// Verify state was updated
		state := n.activeNotifications[workspaceRoot]
		assert.Equal(t, 2, state.processed)
		assert.False(t, state.pending[filePath])
		assert.Len(t, state.pending, 1)

		// Check notification was sent with progress
		notification := <-notifChannel
		assert.Contains(t, notification.Message, "66%")
	})

	t.Run("error status", func(t *testing.T) {
		managerMock := notifiermock.NewMockNotificationManager(ctrl)
		handlerMock := notifiermock.NewMockNotificationHandler(ctrl)
		notifChannel := make(chan notifier.Notification, 10)

		handlerMock.EXPECT().Channel().Return(notifChannel).AnyTimes()

		n := NewIndexNotifier(managerMock)
		// Set up a notification with pending files
		n.activeNotifications[workspaceRoot] = &notificationState{
			handler:   handlerMock,
			processed: 1,
			pending: map[string]bool{
				filePath:                     true,
				"/workspace/another/file.go": true,
			},
		}

		n.NotifyComplete(ctx, workspaceRoot, filePath, "error loading")

		// error notification
		notification := <-notifChannel
		assert.Contains(t, notification.Message, "Failed to load")
		assert.Contains(t, notification.Message, "file.go")
		assert.Contains(t, notification.Message, "error loading")

		// progress notification
		notification = <-notifChannel
		assert.Contains(t, notification.Message, "66%")
	})

	t.Run("completion of all files", func(t *testing.T) {
		managerMock := notifiermock.NewMockNotificationManager(ctrl)
		handlerMock := notifiermock.NewMockNotificationHandler(ctrl)
		notifChannel := make(chan notifier.Notification, 10)

		handlerMock.EXPECT().Channel().Return(notifChannel).AnyTimes()
		handlerMock.EXPECT().Done(gomock.Any()).Return()

		n := NewIndexNotifier(managerMock)
		// Set up a notification with only one pending file
		n.activeNotifications[workspaceRoot] = &notificationState{
			handler:   handlerMock,
			processed: 2,
			pending: map[string]bool{
				filePath: true,
			},
		}

		n.NotifyComplete(ctx, workspaceRoot, filePath, "done")

		// progress notification
		notification := <-notifChannel
		assert.Contains(t, notification.Message, "100%")

		// completion notification
		notification = <-notifChannel
		assert.Contains(t, notification.Message, "Completed loading indices")
		assert.Contains(t, notification.Message, "3 files")

		// Verify state was cleaned up
		_, exists := n.activeNotifications[workspaceRoot]
		assert.False(t, exists)
	})
}
