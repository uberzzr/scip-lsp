package notifier

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

const (
	_errSendToClient = "sending call/notification to IDE: %w"

	_timeoutUserSelectionMoreInfo = time.Second * 5
	_timeoutUserSelectionEnd      = time.Minute * 2

	_titleUserInputProgress          = "User Input Needed"
	_messageUserInputProgressInitial = "Please make a selection from the prompt."
	_messageUserInputProgressUpdate  = "Waiting for a selection. Click here to expand notifications if you don't see a prompt."
)

// Gateway is used to send outbound notifications and calls to the IDE.
// All calls to the gateway should include a context with a session UUID, which will be used to route outbound calls and notifications to the correct IDE session.
type Gateway interface {
	// Methods used to manage the client for each session.

	// RegisterClient registers a new client with the gateway. Should be called each time a new IDE connection is initialized.
	RegisterClient(ctx context.Context, id uuid.UUID, conn *jsonrpc2.Conn) error
	// DeregisterClient removes a client from the gateway. Should be called each time an IDE connection is closed.
	DeregisterClient(ctx context.Context, id uuid.UUID) error

	// Methods from protocol.Client interface.
	Progress(ctx context.Context, params *protocol.ProgressParams) (err error)
	WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error)
	// LogMessage sends a LogMessage notification to the IDE. When using Uber Dev Portal client, use level Info or higher to force display of output panel.
	LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error)
	PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error)
	ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error)
	ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error)
	Telemetry(ctx context.Context, params interface{}) (err error)
	RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) (err error)
	UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error)
	ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result *protocol.ApplyWorkspaceEditResponse, err error)
	Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []interface{}, err error)
	WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error)
	ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (result *protocol.ShowDocumentResult, err error)

	// GetLogMessageWriter returns an io.Writer that can be used to log messages to the IDE client.
	// Do not store or use across requests, get a new one each time as needed.
	GetLogMessageWriter(ctx context.Context, prefix string) (io.Writer, error)
}

type gateway struct {
	clients     map[uuid.UUID]protocol.Client
	connections map[uuid.UUID]jsonrpc2.Conn
	clientsMu   sync.Mutex
	logger      *zap.Logger
}

// New returns a Gateway for sending IDE notifications and calls.
func New(logger *zap.Logger) Gateway {
	return &gateway{
		clients:     make(map[uuid.UUID]protocol.Client),
		connections: make(map[uuid.UUID]jsonrpc2.Conn),
		logger:      logger,
	}
}

func (g *gateway) RegisterClient(ctx context.Context, id uuid.UUID, conn *jsonrpc2.Conn) error {
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()

	client := protocol.ClientDispatcher(*conn, g.logger)
	g.clients[id] = client
	g.connections[id] = *conn

	return nil
}

func (g *gateway) DeregisterClient(ctx context.Context, id uuid.UUID) error {
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()

	delete(g.clients, id)
	delete(g.connections, id)

	return nil
}

func (g *gateway) Progress(ctx context.Context, params *protocol.ProgressParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.Progress(ctx, params)
}

func (g *gateway) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.WorkDoneProgressCreate(ctx, params)
}

func (g *gateway) LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.LogMessage(ctx, params)
}

func (g *gateway) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.PublishDiagnostics(ctx, params)
}

func (g *gateway) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.ShowMessage(ctx, params)
}

func (g *gateway) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}

	if params.Type > protocol.MessageTypeError {
		// Messages levels Warn and below get hidden if the user has their notifications silenced.
		// Guide them to check their notifications.
		showMessageDone, err := g.showWaitingForUserSelection(ctx)
		if err != nil {
			return nil, fmt.Errorf(_errSendToClient, err)
		}
		defer showMessageDone()
	}

	return c.ShowMessageRequest(ctx, params)
}

func (g *gateway) Telemetry(ctx context.Context, params interface{}) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.Telemetry(ctx, params)
}

func (g *gateway) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.RegisterCapability(ctx, params)
}

func (g *gateway) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return fmt.Errorf(_errSendToClient, err)
	}
	return c.UnregisterCapability(ctx, params)
}

func (g *gateway) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result *protocol.ApplyWorkspaceEditResponse, err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}
	return c.ApplyEdit(ctx, params)
}

func (g *gateway) Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []interface{}, err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}
	return c.Configuration(ctx, params)
}

func (g *gateway) WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}
	return c.WorkspaceFolders(ctx)
}

func (g *gateway) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (result *protocol.ShowDocumentResult, err error) {
	_, c, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}

	// Due to https://github.com/go-language-server/protocol/issues/51, ShowDocument is not currently included in protocol.Client.
	// Call the method directly.
	err = protocol.Call(ctx, c, protocol.MethodShowDocument, params, result)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}
	return result, nil
}

func (g *gateway) getClient(ctx context.Context) (protocol.Client, jsonrpc2.Conn, error) {
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()

	id, err := mapper.ContextToSessionUUID(ctx)
	if err != nil {
		return nil, nil, err
	}

	client, ok := g.clients[id]
	if !ok {
		return nil, nil, fmt.Errorf("client with id %q not found", id)
	}

	conn, ok := g.connections[id]
	if !ok {
		return nil, nil, fmt.Errorf("client with id %q not found", id)
	}
	return client, conn, nil
}

// showWaitingForUserSelection sends a notification to the IDE client that a user selection is required.
// This is added in case the IDE client has notifications hidden, to make the user aware that their action is needed.
func (g *gateway) showWaitingForUserSelection(ctx context.Context) (doneFunc func(), err error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}

	tokenID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf(_errSendToClient, err)
	}

	// Initial message indicating pending put.
	token := protocol.NewProgressToken(tokenID.String())
	if err := c.WorkDoneProgressCreate(ctx, &protocol.WorkDoneProgressCreateParams{Token: *protocol.NewProgressToken(tokenID.String())}); err != nil {
		return nil, fmt.Errorf("creating user input progress: %w", err)
	}
	if err := c.Progress(ctx, &protocol.ProgressParams{
		Token: *token,
		Value: &protocol.WorkDoneProgressBegin{
			Kind:        protocol.WorkDoneProgressKindBegin,
			Title:       _titleUserInputProgress,
			Message:     _messageUserInputProgressInitial,
			Cancellable: true,
		},
	}); err != nil {
		return nil, fmt.Errorf("starting user input progress: %w", err)
	}

	// After brief delay without input, update the prompt.
	updateProgressTimer := time.AfterFunc(_timeoutUserSelectionMoreInfo, func() {
		c.Progress(ctx, &protocol.ProgressParams{
			Token: *token,
			Value: &protocol.WorkDoneProgressReport{
				Kind:    protocol.WorkDoneProgressKindReport,
				Message: _messageUserInputProgressUpdate,
			},
		})
	})

	// Set a maximum timeout to avoid showing this indefinitely, since the user may ignore the prompt and move onto other tasks.
	endProgressFunc := func() {
		c.Progress(ctx, &protocol.ProgressParams{
			Token: *token,
			Value: &protocol.WorkDoneProgressEnd{Kind: protocol.WorkDoneProgressKindEnd},
		})
	}
	endProgressTimer := time.AfterFunc(_timeoutUserSelectionEnd, endProgressFunc)

	// Return a function that can be called to immediately end the progress.
	doneFunc = func() {
		updateProgressTimer.Stop()
		if endProgressTimer.Stop() {
			endProgressFunc()
		}
	}
	return doneFunc, nil
}

// ideWriter implements io.Writer to allow logging to the IDE client in situations that require an io.Writer.
type logMessageWriter struct {
	client protocol.Client
	ctx    context.Context
	prefix string
}

func (g *gateway) GetLogMessageWriter(ctx context.Context, prefix string) (io.Writer, error) {
	c, _, err := g.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting IDE log message writer: %w", err)
	}
	w := &logMessageWriter{
		client: c,
		ctx:    ctx,
		prefix: prefix,
	}
	return w, nil
}

func (w *logMessageWriter) Write(p []byte) (n int, err error) {
	str := strings.TrimSuffix(string(p), "\n")
	if err := w.client.LogMessage(w.ctx, &protocol.LogMessageParams{
		Message: fmt.Sprintf("[%s] %s", w.prefix, str),
		Type:    protocol.MessageTypeLog,
	}); err != nil {
		return 0, fmt.Errorf("writing to IDE log message writer: %w", err)
	}
	return len(p), nil
}
