// Package ulspdaemon implements the ulsp-daemon business logic.
package ulspdaemon

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofrs/uuid"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// Initialize will store information about a new connection and perform any setup needed.
func (c *controller) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {

	result := &protocol.InitializeResult{
		ServerInfo: &protocol.ServerInfo{
			Name: "Uber Language Server",
		},
	}

	s, err := c.sessions.GetFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting session from context: %w", err)
	}

	s.InitializeParams = params
	if s.WorkspaceRoot, err = c.workspaceUtils.GetWorkspaceRoot(ctx, params.WorkspaceFolders); err != nil {
		s.UlspEnabled = false
		c.logger.Warnf("getting workspace root: %s", err)
	}

	if s.WorkspaceRoot != "" {
		if s.Monorepo, err = c.workspaceUtils.GetRepoName(ctx, s.WorkspaceRoot); err != nil {
			return nil, fmt.Errorf("getting repo name: %w", err)
		}

		if s.Env, err = c.workspaceUtils.GetEnv(ctx, s.WorkspaceRoot); err != nil {
			return nil, fmt.Errorf("getting environment: %w", err)
		}
	}

	if err := c.sessions.Set(ctx, s); err != nil {
		return nil, fmt.Errorf("setting updated session state: %w", err)
	}

	if s.UlspEnabled {
		result.Capabilities = protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindIncremental,
				Save: &protocol.SaveOptions{
					IncludeText: true,
				},
				WillSave:          true,
				WillSaveWaitUntil: true,
			},
		}

		if err := c.registerSessionPlugins(ctx); err != nil {
			return nil, fmt.Errorf("registering session plugins: %w", err)
		}
	}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Initialize(ctx, params, result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Initialize(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	if err := c.executePluginMethods(ctx, protocol.MethodInitialize, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

// Initialized handles any actions that need to occur immediately after initialization.
func (c *controller) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Initialized(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	if err := c.executePluginMethods(ctx, protocol.MethodInitialized, call, call); err != nil {
		return fmt.Errorf(_errBadPluginCall, err)
	}

	s, err := c.sessions.GetFromContext(ctx)
	if err != nil {
		return fmt.Errorf("getting session from context: %w", err)
	}

	if s.UlspEnabled {
		c.ideGateway.ShowMessage(ctx, &protocol.ShowMessageParams{
			Message: "Connection to Uber Language Server is now initialized.",
			Type:    protocol.MessageTypeInfo,
		})
	} else {
		c.ideGateway.ShowMessage(ctx, &protocol.ShowMessageParams{
			Message: "Uber Language Server is not available for this workspace. See t.uber.com/ulsp for more info.",
			Type:    protocol.MessageTypeWarning,
		})
	}

	return nil
}

// Shutdown is sent just before Exit to indicate that the session will exit.
func (c *controller) Shutdown(ctx context.Context) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Shutdown(ctx); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	if err := c.executePluginMethods(ctx, protocol.MethodShutdown, call, call); err != nil {
		return fmt.Errorf(_errBadPluginCall, err)
	}
	return nil
}

// Exit will be used to either clean up from an individual connection, or shutdown the whole server.
func (c *controller) Exit(ctx context.Context) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Exit(ctx); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	if err := c.executePluginMethods(ctx, protocol.MethodExit, call, call); err != nil {
		c.logger.Errorf(_errBadPluginCall, err)
	}

	if c.fullShutdown == true {
		// Zero out the timer to trigger immediate shutdown.
		c.idleTimerMu.Lock()
		c.idleTimer.Reset(0)
		c.idleTimerMu.Unlock()
		return nil
	}
	s, err := c.sessions.GetFromContext(ctx)
	if err != nil {
		return fmt.Errorf("error during session exit: %w", err)
	}

	return c.EndSession(ctx, s.UUID)
}

// RequestFullShutdown will set the controller to treat subsequent Shutdown and Exit requests as requests to exit the entire process.
func (c *controller) RequestFullShutdown(ctx context.Context) error {
	c.fullShutdown = true

	return nil
}

// InitSession creates a new empty session and returns its UUID.
func (c *controller) InitSession(ctx context.Context, conn *jsonrpc2.Conn) (uuid.UUID, error) {
	defer c.refreshIdleTimer(ctx)

	id, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, err
	}

	session := mapper.UUIDToSession(id, conn)
	if err := c.ideGateway.RegisterClient(ctx, id, conn); err != nil {
		return uuid.Nil, err
	}

	if err := c.sessions.Set(ctx, session); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// EndSession includes any cleanup at the end of the session, during or after the last JSON-RPC request.
func (c *controller) EndSession(ctx context.Context, uuid uuid.UUID) error {
	defer c.refreshIdleTimer(ctx)

	if _, ok := c.pluginMethods[uuid]; ok {
		call := func(ctx context.Context, m *ulspplugin.Methods) {
			if err := m.EndSession(ctx, uuid); err != nil {
				c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
			}
		}
		if err := c.executePluginMethods(ctx, ulspplugin.MethodEndSession, call, call); err != nil {
			c.logger.Errorf(_errBadPluginCall, err)
		}
	}

	err := c.ideGateway.DeregisterClient(ctx, uuid)
	if err != nil {
		c.logger.Error(err)
	}

	delete(c.pluginMethods, uuid)
	return c.sessions.Delete(ctx, uuid)
}

// refreshIdleTimer ensures that the service shuts down after a defined inactivity period with no connections.
func (c *controller) refreshIdleTimer(ctx context.Context) error {
	c.idleTimerMu.Lock()
	defer c.idleTimerMu.Unlock()

	// First call initializes new timer and leaves it running prior to first connection.
	if c.idleTimer == nil {
		c.idleTimer = time.NewTimer(c.idleTimeoutMinutes)
		go func() {
			<-c.idleTimer.C
			c.logger.Info("Shutdown signal received.")
			if err := c.shutdowner.Shutdown(); err != nil {
				os.Exit(1)
			}
		}()
		return nil
	}

	// Subsequent calls stop the timer and reset it only if no connections are active.
	currentSessions, err := c.sessions.SessionCount(ctx)
	if err != nil {
		return fmt.Errorf("error resetting timeout: %w", err)
	}

	c.idleTimer.Stop()
	if currentSessions == 0 {
		c.idleTimer.Reset(c.idleTimeoutMinutes)
	}
	return nil
}
