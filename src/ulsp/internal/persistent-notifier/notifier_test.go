package notifier

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/gateway/ide-client/ideclientmock"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNewNotificationHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)
	logger := zap.NewNop().Sugar()

	p := notificationHandlerParams{
		ParentManager: &notificationManagerImpl{managers: make(map[string]NotificationHandler)},
		Sessions:      sessionMock,
		IdeGateway:    ideClientMock,
		Logger:        logger,
		WorkspaceRoot: "sampleWorkspace",
		Title:         "sample title",
	}

	sampleSessions := []*entity.Session{
		&entity.Session{
			UUID:          factory.UUID(),
			WorkspaceRoot: "sampleWorkspace",
		},
		&entity.Session{
			UUID:          factory.UUID(),
			WorkspaceRoot: "sampleWorkspace",
		},
	}
	sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(sampleSessions, nil).Times(2)

	for range sampleSessions {
		ideClientMock.EXPECT().WorkDoneProgressCreate(gomock.Any(), gomock.Any()).Return(nil)
		ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil)
		ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil)
	}

	h, err := NewNotificationHandler(ctx, p, "notificationId")
	impl := h.(*notificationHandlerImpl)
	h.Done(ctx)
	impl.handlerWg.Wait()
	assert.NoError(t, err)
}

func TestAdd(t *testing.T) {
	t.Run("positive value", func(t *testing.T) {
		h := &notificationHandlerImpl{
			senderCount: 1,
		}

		h.Add(context.Background())
		assert.Equal(t, 2, h.senderCount)
	})

	t.Run("already zero", func(t *testing.T) {
		h := &notificationHandlerImpl{
			senderCount: 0,
		}

		h.Add(context.Background())
		assert.Equal(t, 0, h.senderCount)
	})
}

func TestDone(t *testing.T) {
	t.Run("last sender", func(t *testing.T) {
		h := &notificationHandlerImpl{
			closeCh:     make(chan bool, 1),
			senderCount: 1,
		}
		h.senderWg.Add(1)
		h.Done(context.Background())
		assert.Equal(t, 0, h.senderCount)
		h.senderWg.Wait()
	})

	t.Run("other senders", func(t *testing.T) {
		h := &notificationHandlerImpl{
			closeCh:     make(chan bool, 1),
			senderCount: 2,
		}
		h.Done(context.Background())
		assert.Equal(t, 1, h.senderCount)
	})

	t.Run("extra calls", func(t *testing.T) {
		h := &notificationHandlerImpl{
			closeCh:     make(chan bool, 1),
			senderCount: 0,
			logger:      zap.NewNop().Sugar(),
		}
		h.Done(context.Background())
		assert.Equal(t, -1, h.senderCount)
	})
}

func TestChannel(t *testing.T) {
	h := &notificationHandlerImpl{
		channel: make(chan Notification, 20),
	}

	assert.Equal(t, h.channel, h.Channel())
}

func TestHandleUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)

	h := &notificationHandlerImpl{
		parentManager:         &notificationManagerImpl{managers: map[string]NotificationHandler{"sample": &notificationHandlerImpl{}}},
		ideGateway:            ideClientMock,
		sessions:              sessionMock,
		closeCh:               make(chan bool),
		channel:               make(chan Notification, 20),
		token:                 protocol.NewProgressToken(factory.UUID().String()),
		notificationsBySender: map[string]map[string]Notification{},
		senderCount:           1,
	}
	h.senderWg.Add(1)

	// Expect one update and one close for each notification.
	sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return([]*entity.Session{&entity.Session{
		UUID: factory.UUID(),
	}}, nil).Times(2)
	ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	h.channel <- Notification{SenderToken: "sample", IdentifierToken: "sample-identifier", Priority: 0, Message: "abc"}
	h.handleUpdates()
	h.Done(context.Background())
	h.handlerWg.Wait()
}

func TestCleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)

	t.Run("success", func(t *testing.T) {
		h := &notificationHandlerImpl{
			id:            "sample",
			parentManager: &notificationManagerImpl{managers: map[string]NotificationHandler{"sample": &notificationHandlerImpl{}}},
			ideGateway:    ideClientMock,
			closeCh:       make(chan bool, 1),
			channel:       make(chan Notification, 20),
			sessions:      sessionMock,
			token:         protocol.NewProgressToken(factory.UUID().String()),
		}

		sampleSessions := []*entity.Session{}
		for i := 0; i < 3; i++ {
			sampleSessions = append(sampleSessions, &entity.Session{
				UUID: factory.UUID(),
			})
		}
		sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(sampleSessions, nil)

		ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil).Times(len(sampleSessions))
		h.cleanup(ctx)
		assert.True(t, h.isClosed)
		assert.True(t, <-h.closeCh)
		_, ok := <-h.channel
		assert.False(t, ok)
	})

	t.Run("failed to get sessions", func(t *testing.T) {
		h := &notificationHandlerImpl{
			id:            "sample",
			parentManager: &notificationManagerImpl{managers: make(map[string]NotificationHandler)},
			ideGateway:    ideClientMock,
			closeCh:       make(chan bool, 1),
			channel:       make(chan Notification, 20),
			sessions:      sessionMock,
			token:         protocol.NewProgressToken(factory.UUID().String()),
		}

		sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(nil, errors.New("sample"))
		h.cleanup(ctx)
		assert.True(t, h.isClosed)
		assert.True(t, <-h.closeCh)
		_, ok := <-h.channel
		assert.False(t, ok)
	})

	t.Run("already closed", func(t *testing.T) {
		h := &notificationHandlerImpl{
			isClosed: true,
		}

		assert.NotPanics(t, func() { h.cleanup(ctx) })
	})
}

func TestBroadcastUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	ideClientMock := ideclientmock.NewMockGateway(ctrl)
	sessionMock := repositorymock.NewMockRepository(ctrl)

	t.Run("success", func(t *testing.T) {
		sampleSessions := []*entity.Session{}
		for i := 0; i < 3; i++ {
			sampleSessions = append(sampleSessions, &entity.Session{
				UUID: factory.UUID(),
			})
		}
		sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(sampleSessions, nil)

		h := &notificationHandlerImpl{
			ideGateway: ideClientMock,
			sessions:   sessionMock,

			closeCh:               make(chan bool),
			token:                 protocol.NewProgressToken(factory.UUID().String()),
			notificationsBySender: map[string]map[string]Notification{},
		}

		ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil).Times(len(sampleSessions))
		h.broadcastUpdate(ctx, Notification{
			SenderToken:     "a",
			IdentifierToken: "sample-identifier",
			Priority:        0,
			Message:         "abc",
		})
	})

	t.Run("failed to get sessions", func(t *testing.T) {
		sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(nil, errors.New("sample"))

		h := &notificationHandlerImpl{
			ideGateway:            ideClientMock,
			sessions:              sessionMock,
			logger:                zap.NewNop().Sugar(),
			closeCh:               make(chan bool),
			token:                 protocol.NewProgressToken(factory.UUID().String()),
			notificationsBySender: map[string]map[string]Notification{},
		}

		h.broadcastUpdate(ctx, Notification{
			SenderToken:     "a",
			IdentifierToken: "sample-identifier",
			Priority:        0,
			Message:         "abc",
		})
	})

	t.Run("failed to get sessions", func(t *testing.T) {
		sessionMock.EXPECT().GetAllFromWorkspaceRoot(gomock.Any(), gomock.Any()).Return(nil, errors.New("sample"))

		h := &notificationHandlerImpl{
			ideGateway:            ideClientMock,
			sessions:              sessionMock,
			logger:                zap.NewNop().Sugar(),
			closeCh:               make(chan bool),
			token:                 protocol.NewProgressToken(factory.UUID().String()),
			notificationsBySender: map[string]map[string]Notification{},
		}

		h.broadcastUpdate(ctx, Notification{
			SenderToken:     "a",
			IdentifierToken: "sample-identifier",
			Priority:        0,
			Message:         "abc",
		})
	})

}

func TestCaptureUpdate(t *testing.T) {
	tests := []struct {
		name                 string
		notificationSequence []Notification
		expectedResult       string
	}{
		{
			name: "single update",
			notificationSequence: []Notification{
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier",
					Priority:        0,
					Message:         "abc",
				},
			},
			expectedResult: "abc",
		},
		{
			name: "distinct updates",
			notificationSequence: []Notification{
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
					Message:         "abc",
				},
				{
					SenderToken:     "b",
					IdentifierToken: "sample-identifier2",
					Priority:        0,
					Message:         "def",
				},
				{
					SenderToken:     "c",
					IdentifierToken: "sample-identifier3",
					Priority:        0,
					Message:         "ghi",
				},
			},
			expectedResult: "[abc, def, ghi]",
		},
		{
			name: "multiple senders, same identifier",
			notificationSequence: []Notification{
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier1",
					Priority:        5,
					Message:         "abc",
				},
				{
					SenderToken:     "b",
					IdentifierToken: "sample-identifier2",
					Priority:        0,
					Message:         "def",
				},
				{
					SenderToken:     "c",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
					Message:         "ghi",
				},
			},

			expectedResult: "[def, ghi]",
		},
		{
			name: "single sender, multiple identifiers",
			notificationSequence: []Notification{
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
					Message:         "abc",
				},
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier2",
					Priority:        0,
					Message:         "def",
				},
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier3",
					Priority:        0,
					Message:         "ghi",
				},
			},
			expectedResult: "[abc, def, ghi]",
		},
		{
			name: "add and delete, same sender",
			notificationSequence: []Notification{
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
					Message:         "abc",
				},
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
				},
			},
			expectedResult: "[]",
		},
		{
			name: "add and delete, different senders",
			notificationSequence: []Notification{
				{
					SenderToken:     "a",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
					Message:         "abc",
				},
				{
					SenderToken:     "b",
					IdentifierToken: "sample-identifier1",
					Priority:        0,
				},
			},
			expectedResult: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &notificationHandlerImpl{
				notificationsBySender: map[string]map[string]Notification{},
			}

			var result string
			for _, n := range tt.notificationSequence {
				result = h.captureUpdate(n)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}

}

func TestInitNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	ideClientMock := ideclientmock.NewMockGateway(ctrl)

	s := &entity.Session{
		UUID: factory.UUID(),
	}

	h := &notificationHandlerImpl{
		ideGateway: ideClientMock,
		closeCh:    make(chan bool),
		token:      protocol.NewProgressToken(factory.UUID().String()),
	}

	t.Run("success", func(t *testing.T) {
		ideClientMock.EXPECT().WorkDoneProgressCreate(gomock.Any(), gomock.Any()).Return(nil)
		ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil)
		assert.NoError(t, h.initNotification(ctx, s, "sample"))
	})

	t.Run("failure", func(t *testing.T) {
		ideClientMock.EXPECT().WorkDoneProgressCreate(gomock.Any(), gomock.Any()).Return(errors.New("sample"))
		assert.Error(t, h.initNotification(ctx, s, "sample"))
	})

}

func TestUpdateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	ideClientMock := ideclientmock.NewMockGateway(ctrl)

	s := &entity.Session{
		UUID: factory.UUID(),
	}

	h := &notificationHandlerImpl{
		ideGateway: ideClientMock,
		closeCh:    make(chan bool),
		token:      protocol.NewProgressToken(factory.UUID().String()),
	}

	ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil)
	assert.NoError(t, h.updateNotification(ctx, s, "sample", 0))
}

func TestEndNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	ideClientMock := ideclientmock.NewMockGateway(ctrl)

	s := &entity.Session{
		UUID: factory.UUID(),
	}

	h := &notificationHandlerImpl{
		ideGateway: ideClientMock,
		closeCh:    make(chan bool),
		token:      protocol.NewProgressToken(factory.UUID().String()),
	}

	ideClientMock.EXPECT().Progress(gomock.Any(), gomock.Any()).Return(nil)
	assert.NoError(t, h.endNotification(ctx, s))
}
