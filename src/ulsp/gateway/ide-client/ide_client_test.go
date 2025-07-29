package notifier

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/scip-lsp/idl/mock/jsonrpc2mock"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestRegisterClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	g := gateway{
		clients:     make(map[uuid.UUID]protocol.Client),
		connections: make(map[uuid.UUID]jsonrpc2.Conn),
		logger:      zap.NewNop(),
	}

	for i := 0; i < 10; i++ {
		id := factory.UUID()
		mockConn := jsonrpc2mock.NewMockConn(ctrl)
		var conn jsonrpc2.Conn = mockConn
		err := g.RegisterClient(ctx, id, &conn)
		assert.NoError(t, err)
	}

	assert.Len(t, g.clients, 10)
	assert.Len(t, g.connections, 10)
}

func TestDeregisterClient(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	g := gateway{
		clients:     make(map[uuid.UUID]protocol.Client),
		connections: make(map[uuid.UUID]jsonrpc2.Conn),
		logger:      zap.NewNop(),
	}

	// Set up 10 sample clients.
	for i := 0; i < 10; i++ {
		mockConn := jsonrpc2mock.NewMockConn(ctrl)
		var conn jsonrpc2.Conn = mockConn
		err := g.RegisterClient(ctx, factory.UUID(), &conn)
		require.NoError(t, err)
	}

	// Remove clients one by one and confirm removal.
	for key := range g.clients {
		assert.NotNil(t, g.clients[key])
		err := g.DeregisterClient(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, g.clients[key])
	}
	assert.Len(t, g.clients, 0)
	assert.Len(t, g.connections, 0)
}

func TestProgress(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	progressParams := &protocol.ProgressParams{
		Token: *protocol.NewNumberProgressToken(5),
		Value: "sampleValue",
	}

	t.Run("notification success", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Eq(progressParams)).Return(nil)
		err := g.Progress(ctx, progressParams)
		assert.NoError(t, err)
	})
	t.Run("notification failure", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Eq(progressParams)).Return(errors.New("error"))
		err := g.Progress(ctx, progressParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.Progress(ctx, progressParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.Progress(ctx, progressParams)
		assert.Error(t, err)
	})
}

func TestWorkDoneProgressCreate(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	workDoneProgressCreateParams := &protocol.WorkDoneProgressCreateParams{
		Token: *protocol.NewNumberProgressToken(5),
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Eq(workDoneProgressCreateParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		err := g.WorkDoneProgressCreate(ctx, workDoneProgressCreateParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Eq(workDoneProgressCreateParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		err := g.WorkDoneProgressCreate(ctx, workDoneProgressCreateParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.WorkDoneProgressCreate(ctx, workDoneProgressCreateParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.WorkDoneProgressCreate(ctx, workDoneProgressCreateParams)
		assert.Error(t, err)
	})
}

func TestPublishDiagnostics(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	publishDiagnosticsParams := &protocol.PublishDiagnosticsParams{
		URI:         "file:///sample.go",
		Diagnostics: []protocol.Diagnostic{},
	}

	t.Run("notification success", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodTextDocumentPublishDiagnostics), gomock.Eq(publishDiagnosticsParams)).Return(nil)
		err := g.PublishDiagnostics(ctx, publishDiagnosticsParams)
		assert.NoError(t, err)
	})
	t.Run("notification failure", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodTextDocumentPublishDiagnostics), gomock.Eq(publishDiagnosticsParams)).Return(errors.New("error"))
		err := g.PublishDiagnostics(ctx, publishDiagnosticsParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.PublishDiagnostics(ctx, publishDiagnosticsParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.PublishDiagnostics(ctx, publishDiagnosticsParams)
		assert.Error(t, err)
	})
}

func TestShowMessage(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	messageParams := &protocol.ShowMessageParams{
		Message: "Connection to Uber Language Server is now initialized.",
		Type:    protocol.MessageTypeInfo,
	}

	t.Run("notification success", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowShowMessage), gomock.Eq(messageParams)).Return(nil)
		err := g.ShowMessage(ctx, messageParams)
		assert.NoError(t, err)
	})
	t.Run("notification failure", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowShowMessage), gomock.Eq(messageParams)).Return(errors.New("error"))
		err := g.ShowMessage(ctx, messageParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.ShowMessage(ctx, messageParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.ShowMessage(ctx, messageParams)
		assert.Error(t, err)
	})
}

func TestShowMessageRequest(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	messageParams := &protocol.ShowMessageRequestParams{
		Message: "Connection to Uber Language Server is now initialized.",
		Type:    protocol.MessageTypeInfo,
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), nil)
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Any()).Return(nil).Times(2)
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowShowMessageRequest), gomock.Eq(messageParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		_, err := g.ShowMessageRequest(ctx, messageParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), nil)
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Any()).Return(nil).Times(2)
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowShowMessageRequest), gomock.Eq(messageParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		_, err := g.ShowMessageRequest(ctx, messageParams)
		assert.Error(t, err)
	})
	t.Run("progress create failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), errors.New("error"))
		_, err := g.ShowMessageRequest(ctx, messageParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		_, err := g.ShowMessageRequest(ctx, messageParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		_, err := g.ShowMessageRequest(ctx, messageParams)
		assert.Error(t, err)
	})
}

func TestTelemetry(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	telemetryParams := "sample telemetry message"

	t.Run("notification success", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodTelemetryEvent), gomock.Eq(telemetryParams)).Return(nil)
		err := g.Telemetry(ctx, telemetryParams)
		assert.NoError(t, err)
	})
	t.Run("notification failure", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodTelemetryEvent), gomock.Eq(telemetryParams)).Return(errors.New("error"))
		err := g.Telemetry(ctx, telemetryParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.Telemetry(ctx, telemetryParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.Telemetry(ctx, telemetryParams)
		assert.Error(t, err)
	})
}

func TestRegisterCapability(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	registerCapabilityParams := &protocol.RegistrationParams{
		Registrations: []protocol.Registration{
			{
				ID:     "sampleID",
				Method: protocol.MethodTextDocumentPublishDiagnostics,
			},
		},
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodClientRegisterCapability), gomock.Eq(registerCapabilityParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		err := g.RegisterCapability(ctx, registerCapabilityParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodClientRegisterCapability), gomock.Eq(registerCapabilityParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		err := g.RegisterCapability(ctx, registerCapabilityParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.RegisterCapability(ctx, registerCapabilityParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.RegisterCapability(ctx, registerCapabilityParams)
		assert.Error(t, err)
	})
}

func TestUnregisterCapability(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	unregisterCapabilityParams := &protocol.UnregistrationParams{
		Unregisterations: []protocol.Unregistration{
			{
				ID:     "sampleID",
				Method: protocol.MethodTextDocumentPublishDiagnostics,
			},
		},
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodClientUnregisterCapability), gomock.Eq(unregisterCapabilityParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		err := g.UnregisterCapability(ctx, unregisterCapabilityParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodClientUnregisterCapability), gomock.Eq(unregisterCapabilityParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		err := g.UnregisterCapability(ctx, unregisterCapabilityParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.UnregisterCapability(ctx, unregisterCapabilityParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.UnregisterCapability(ctx, unregisterCapabilityParams)
		assert.Error(t, err)
	})
}

func TestApplyEdit(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	applyEditParams := &protocol.ApplyWorkspaceEditParams{
		Label: "sample",
		Edit:  protocol.WorkspaceEdit{},
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkspaceApplyEdit), gomock.Eq(applyEditParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		_, err := g.ApplyEdit(ctx, applyEditParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkspaceApplyEdit), gomock.Eq(applyEditParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		_, err := g.ApplyEdit(ctx, applyEditParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		_, err := g.ApplyEdit(ctx, applyEditParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		_, err := g.ApplyEdit(ctx, applyEditParams)
		assert.Error(t, err)
	})
}

func TestConfiguration(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	configurationParams := &protocol.ConfigurationParams{
		Items: []protocol.ConfigurationItem{},
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkspaceConfiguration), gomock.Eq(configurationParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		_, err := g.Configuration(ctx, configurationParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkspaceConfiguration), gomock.Eq(configurationParams), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		_, err := g.Configuration(ctx, configurationParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		_, err := g.Configuration(ctx, configurationParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		_, err := g.Configuration(ctx, configurationParams)
		assert.Error(t, err)
	})
}

func TestWorkspaceFolders(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkspaceWorkspaceFolders), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		_, err := g.WorkspaceFolders(ctx)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkspaceWorkspaceFolders), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		_, err := g.WorkspaceFolders(ctx)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		_, err := g.WorkspaceFolders(ctx)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		_, err := g.WorkspaceFolders(ctx)
		assert.Error(t, err)
	})
}

func TestShowDocument(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	showDocumentParams := &protocol.ShowDocumentParams{
		URI:      protocol.URI("http://example.com"),
		External: true,
	}

	t.Run("call success", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodShowDocument), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(5), nil)
		_, err := g.ShowDocument(ctx, showDocumentParams)
		assert.NoError(t, err)
	})
	t.Run("call failure", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodShowDocument), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(5), errors.New("error"))
		_, err := g.ShowDocument(ctx, showDocumentParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		_, err := g.ShowDocument(ctx, showDocumentParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		_, err := g.ShowDocument(ctx, showDocumentParams)
		assert.Error(t, err)
	})
}

func TestShowWaitingForUserSelection(t *testing.T) {
	id := factory.UUID()
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, id)
	ctrl := gomock.NewController(t)
	mockConn := jsonrpc2mock.NewMockConn(ctrl)

	g := gateway{
		logger:      zap.NewNop(),
		clients:     make(map[uuid.UUID]protocol.Client),
		connections: make(map[uuid.UUID]jsonrpc2.Conn),
	}

	var conn jsonrpc2.Conn = mockConn
	g.RegisterClient(ctx, id, &conn)

	t.Run("success without delay", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), nil)
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Any()).Return(nil).Times(2)
		done, err := g.showWaitingForUserSelection(ctx)
		done()
		assert.NoError(t, err)
	})

	t.Run("success with delay", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), nil)
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Any()).Return(nil).Times(3)
		done, err := g.showWaitingForUserSelection(ctx)
		time.Sleep(_timeoutUserSelectionMoreInfo + 1*time.Second)
		done()
		assert.NoError(t, err)
	})

	t.Run("create progress error", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), errors.New("sample"))

		_, err := g.showWaitingForUserSelection(ctx)
		assert.Error(t, err)
	})

	t.Run("start progress error", func(t *testing.T) {
		mockConn.EXPECT().Call(gomock.Eq(ctx), gomock.Eq(protocol.MethodWorkDoneProgressCreate), gomock.Any(), gomock.Any()).Return(jsonrpc2.NewNumberID(4), nil)
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodProgress), gomock.Any()).Return(errors.New("sample"))

		_, err := g.showWaitingForUserSelection(ctx)
		assert.Error(t, err)
	})
}

func TestLogMessage(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	logMessageParams := &protocol.LogMessageParams{
		Message: "sample message",
		Type:    protocol.MessageTypeInfo,
	}

	t.Run("notification success", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowLogMessage), gomock.Eq(logMessageParams)).Return(nil)
		err := g.LogMessage(ctx, logMessageParams)
		assert.NoError(t, err)
	})
	t.Run("notification failure", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowLogMessage), gomock.Eq(logMessageParams)).Return(errors.New("error"))
		err := g.LogMessage(ctx, logMessageParams)
		assert.Error(t, err)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		err := g.LogMessage(ctx, logMessageParams)
		assert.Error(t, err)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		err := g.LogMessage(ctx, logMessageParams)
		assert.Error(t, err)
	})
}

func TestGetLogMessageWriter(t *testing.T) {
	g, _, ctx := getTestGateway(t)

	t.Run("success", func(t *testing.T) {
		writer, err := g.GetLogMessageWriter(ctx, "sample")
		assert.NoError(t, err)
		assert.NotNil(t, writer)
	})
	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		writer, err := g.GetLogMessageWriter(ctx, "sample")
		assert.Error(t, err)
		assert.Nil(t, writer)
	})
	t.Run("client not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, factory.UUID())
		writer, err := g.GetLogMessageWriter(ctx, "sample")
		assert.Error(t, err)
		assert.Nil(t, writer)
	})
}

func TestWrite(t *testing.T) {
	g, mockConn, ctx := getTestGateway(t)

	sampleMsg := "sample message"
	prefix := "my-prefix"
	expectedLogMessageParams := &protocol.LogMessageParams{
		Message: fmt.Sprintf("[%s] %s", prefix, sampleMsg),
		Type:    protocol.MessageTypeLog,
	}

	t.Run("notification success", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowLogMessage), gomock.Eq(expectedLogMessageParams)).Return(nil)
		writer, err := g.GetLogMessageWriter(ctx, prefix)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		n, err := writer.Write([]byte(sampleMsg))
		assert.NoError(t, err)
		assert.Equal(t, len([]byte(sampleMsg)), n)
	})
	t.Run("notification failure", func(t *testing.T) {
		mockConn.EXPECT().Notify(gomock.Eq(ctx), gomock.Eq(protocol.MethodWindowLogMessage), gomock.Eq(expectedLogMessageParams)).Return(errors.New("sample"))
		writer, err := g.GetLogMessageWriter(ctx, prefix)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		n, err := writer.Write([]byte(sampleMsg))
		assert.Error(t, err)
		assert.Equal(t, 0, n)
	})
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func getTestGateway(t *testing.T) (Gateway, *jsonrpc2mock.MockConn, context.Context) {
	id := factory.UUID()
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, id)
	ctrl := gomock.NewController(t)

	mockConn := jsonrpc2mock.NewMockConn(ctrl)
	var conn jsonrpc2.Conn = mockConn
	g := New(zap.NewNop())
	g.RegisterClient(ctx, id, &conn)
	return g, mockConn, ctx
}
