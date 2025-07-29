package userguidance

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/idl/mock/configmock"
	"github.com/uber/scip-lsp/src/ulsp/gateway/ide-client/ideclientmock"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs/fsmock"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/config"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		hasValidConfig bool
		config         string
	}{
		{
			name:           "comprehensive guidance config",
			hasValidConfig: true,
			config: `
guidance:
  messages:
    - key: test-output
      kind: output
      message: test output message
      type: info
    - key: test-notification
      kind: notification
      message: test notification message
      type: error
    - key: test-actionable-notification
      kind: notification
      message: "test notification message with actions"
      type: warning
      actions:
        - title: go
          uri: "https://go.dev"
        - title: java
          uri: "https://java.com"
        - title: hide
          save: true
`,
		},
		{
			name:           "guidance config with no messages",
			hasValidConfig: true,
			config: `
guidance:
  messages: []
`,
		},
		{
			name:           "erroneous guidance config",
			hasValidConfig: false,
			config: `
guidance: []
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			configProvider, err := config.NewYAML(config.Source(strings.NewReader(test.config)))
			if err != nil {
				t.Fatalf("new yaml provider from config: %v", err)
			}

			configProviderMock := configmock.NewMockProvider(ctrl)
			configProviderMock.
				EXPECT().
				Get(_configKey).
				Return(configProvider.Get(_configKey)).
				AnyTimes()

			controller, err := New(Params{Config: configProviderMock})
			if test.hasValidConfig {
				assert.NoError(t, err)
				assert.NotNil(t, controller)
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "configure guidance")
				assert.Nil(t, controller)
			}
		})
	}
}

func TestStartupInfo(t *testing.T) {
	ctx := context.Background()
	c := controller{}
	result, err := c.StartupInfo(ctx)

	assert.NoError(t, err)
	assert.NoError(t, result.Validate())
	assert.Equal(t, _nameKey, result.NameKey)
}

func TestInitialized(t *testing.T) {
	tests := []struct {
		name       string
		messages   []Message
		mockSetup  func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS)
		assertions func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error)
	}{
		{
			name: "output message",
			messages: []Message{
				{
					Key:     "test-output",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test output message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					LogMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.LogMessageParams) bool {
						assert.Equal(t, "test output message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(nil).
					Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil).AnyTimes()
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, os.ErrNotExist)
				fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
				fsMock.EXPECT().Create(gomock.Any()).Return(nil, nil)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "notification message",
			messages: []Message{
				{
					Key:     "test-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test notification message",
					Type:    "warning",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageParams) bool {
						assert.Equal(t, "test notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeWarning, params.Type)
						return true
					}).
					Return(nil).
					Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil).AnyTimes()
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, os.ErrNotExist)
				fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
				fsMock.EXPECT().Create(gomock.Any()).Return(nil, nil)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "actionable notification message with uri selection",
			messages: []Message{
				{
					Key:     "test-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test actionable notification message",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(&protocol.MessageActionItem{Title: "java"}, nil).
					Times(1)
				ideGatewayMock.
					EXPECT().
					ShowDocument(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowDocumentParams) bool {
						assert.Equal(t, "https://java.com", string(params.URI))
						return true
					}).
					Return(&protocol.ShowDocumentResult{Success: true}, nil).
					Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "actionable notification message with save selection",
			messages: []Message{
				{
					Key:     "test-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test actionable notification message",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
						{Title: "hide", Save: true},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(&protocol.MessageActionItem{Title: "hide"}, nil).
					Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil).AnyTimes()
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, os.ErrNotExist)
				fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
				fsMock.EXPECT().Create(gomock.Any()).Return(nil, nil).Times(1)
				fsMock.EXPECT().WriteFile(gomock.Any(), gomock.Any()).Return(nil)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "actionable notification message with no selection",
			messages: []Message{
				{
					Key:     "test-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test actionable notification message",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(nil, nil).
					Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, os.ErrNotExist)
				ideGatewayMock.EXPECT().ShowDocument(gomock.Any(), gomock.Any()).Times(0)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "duplicate output messages",
			messages: []Message{
				{
					Key:     "test-duplicate-output",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test output message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, nil)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
				ideGatewayMock.EXPECT().LogMessage(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name: "duplicate notification messages",
			messages: []Message{
				{
					Key:     "test-duplicate-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test notification message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, nil)
				ideGatewayMock.EXPECT().ShowMessage(gomock.Any(), gomock.Any()).Times(0)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error stat shown output message due to cache dir",
			messages: []Message{
				{
					Key:     "test-error-has-shown-message",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test error stat shown message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				fsMock.EXPECT().UserCacheDir().Return("", errors.New("mock error"))
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "stat shown message")
			},
		},
		{
			name: "error stat shown notification message due to cache dir",
			messages: []Message{
				{
					Key:     "test-error-has-shown-message",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error stat shown message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				fsMock.EXPECT().UserCacheDir().Return("", errors.New("mock error"))
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "stat shown message")
			},
		},
		{
			name: "error stat shown message due to read file",
			messages: []Message{
				{
					Key:     "test-error-has-shown-message",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test error stat shown message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, errors.New("mock error"))
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "stat shown message")
			},
		},
		{
			name: "error output message",
			messages: []Message{
				{
					Key:     "test-error-output",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test error output message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					LogMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.LogMessageParams) bool {
						assert.Equal(t, "test error output message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(errors.New("mock error"))

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "output message 'test-error-output'")
			},
		},
		{
			name: "error notify message",
			messages: []Message{
				{
					Key:     "test-error-notify",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error notify message",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageParams) bool {
						assert.Equal(t, "test error notify message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(errors.New("mock error"))

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "notify message 'test-error-notify'")
			},
		},
		{
			name: "error mark message as shown due to cache dir",
			messages: []Message{
				{
					Key:     "test-error-mark-message-as-shown",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test error mark message as shown",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					LogMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.LogMessageParams) bool {
						assert.Equal(t, "test error mark message as shown", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(nil)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().UserCacheDir().Return("", errors.New("mock error"))
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "mark message as shown")
			},
		},
		{
			name: "error mark message as shown due to mkdir all",
			messages: []Message{
				{
					Key:     "test-error-mark-message-as-shown",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error mark message as shown",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageParams) bool {
						assert.Equal(t, "test error mark message as shown", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(nil)

				fsMock.EXPECT().UserCacheDir().Return("", nil).AnyTimes()
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
				fsMock.EXPECT().MkdirAll(gomock.Any()).Return(errors.New("mock error"))
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "mark message as shown")
			},
		},
		{
			name: "error mark message as shown due to create",
			messages: []Message{
				{
					Key:     "test-error-mark-message-as-shown",
					Kind:    userGuidanceMessageKindOutput,
					Message: "test error mark message as shown",
					Type:    "info",
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					LogMessage(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.LogMessageParams) bool {
						assert.Equal(t, "test error mark message as shown", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(nil)

				fsMock.EXPECT().UserCacheDir().Return("", nil).AnyTimes()
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
				fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
				fsMock.EXPECT().Create(gomock.Any()).Return(nil, errors.New("mock error"))
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "mark message as shown")
			},
		},
		{
			name: "error mark message as shown due to write file",
			messages: []Message{
				{
					Key:     "test-error-mark-message-as-shown",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error mark message as shown",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com", Save: true},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test error mark message as shown", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(&protocol.MessageActionItem{Title: "java"}, nil).
					Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil).AnyTimes()
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
				fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
				fsMock.EXPECT().Create(gomock.Any()).Return(nil, nil)
				fsMock.EXPECT().WriteFile(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "mark message as shown")
			},
		},
		{
			name: "error actionable notification message with uri selection due to show message request",
			messages: []Message{
				{
					Key:     "test-error-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error actionable notification message",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.EXPECT().ShowMessageRequest(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock error")).Times(1)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "notify message 'test-error-actionable-notification'")
			},
		},
		{
			name: "error actionable notification message with uri selection due to show document",
			messages: []Message{
				{
					Key:     "test-error-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error actionable notification message",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test error actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(&protocol.MessageActionItem{Title: "java"}, nil).
					Times(1)
				ideGatewayMock.EXPECT().ShowDocument(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock error"))

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "notify message 'test-error-actionable-notification'")
			},
		},
		{
			name: "error actionable notification message with doNotShowAgain selection due to cache dir",
			messages: []Message{
				{
					Key:     "test-error-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error actionable notification message",
					Type:    "error",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
						{Title: "hide", Save: true},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test error actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeError, params.Type)
						return true
					}).
					Return(&protocol.MessageActionItem{Title: "hide"}, nil)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().UserCacheDir().Return("", errors.New("mock error"))
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "notify message 'test-error-actionable-notification'")
				assert.ErrorContains(t, err, "mark message as shown")
			},
		},
		{
			name: "error actionable notification message with selection due to no matching action",
			messages: []Message{
				{
					Key:     "test-error-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test error actionable notification message",
					Type:    "log",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
						{Title: "hide", Save: true},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test error actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeLog, params.Type)
						return true
					}).
					Return(&protocol.MessageActionItem{Title: "unmatched"}, nil)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "display applicable messages")
				assert.ErrorContains(t, err, "notify message 'test-error-actionable-notification'")
				assert.ErrorContains(t, err, "no action matches selection 'unmatched'")
			},
		},
		{
			name: "actionable notification message with canceled show message request",
			messages: []Message{
				{
					Key:     "test-actionable-notification",
					Kind:    userGuidanceMessageKindNotification,
					Message: "test actionable notification message",
					Type:    "info",
					Actions: []Action{
						{Title: "go", URI: "https://go.dev"},
						{Title: "java", URI: "https://java.com"},
					},
				},
			},
			mockSetup: func(ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS) {
				ideGatewayMock.
					EXPECT().
					ShowMessageRequest(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, params *protocol.ShowMessageRequestParams) bool {
						assert.Equal(t, "test actionable notification message", params.Message)
						assert.Equal(t, protocol.MessageTypeInfo, params.Type)
						return true
					}).
					Return(nil, jsonrpc2.NewError(jsonrpc2.InternalError, "Request window/showMessageRequest failed with message: Canceled")).
					Times(1)
				ideGatewayMock.EXPECT().ShowDocument(gomock.Any(), gomock.Any()).Times(0)

				fsMock.EXPECT().UserCacheDir().Return("", nil)
				fsMock.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)
			},
			assertions: func(t *testing.T, ideGatewayMock *ideclientmock.MockGateway, fsMock *fsmock.MockUlspFS, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ideGatewayMock := ideclientmock.NewMockGateway(ctrl)
			fsMock := fsmock.NewMockUlspFS(ctrl)

			c := controller{
				ideGateway: ideGatewayMock,
				fs:         fsMock,
				guidance: guidance{
					Messages: test.messages,
				},
			}

			test.mockSetup(ideGatewayMock, fsMock)
			err := c.initialized(context.Background(), &protocol.InitializedParams{})
			test.assertions(t, ideGatewayMock, fsMock, err)
		})
	}
}
