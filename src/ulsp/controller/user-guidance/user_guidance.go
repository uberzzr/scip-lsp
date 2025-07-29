package userguidance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	ideclient "github.com/uber/scip-lsp/src/ulsp/gateway/ide-client"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
	"github.com/uber/scip-lsp/src/ulsp/repository/session"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/config"
	"go.uber.org/fx"
)

type userGuidanceMessageKind string

const (
	// userGuidanceMessageKindOutput writes to the IDE Output window
	// Good for longer messages
	userGuidanceMessageKindOutput userGuidanceMessageKind = "output"
	// userGuidanceMessageKindNotification sends an IDE notification
	userGuidanceMessageKindNotification userGuidanceMessageKind = "notification"
)

type guidance struct {
	Messages []Message `yaml:"messages"`
}

// Message represents a user guidance message.
type Message struct {
	Key     string                  `yaml:"key"`
	Kind    userGuidanceMessageKind `yaml:"kind"`
	Message string                  `yaml:"message"`
	Type    string                  `yaml:"type"`
	Actions []Action                `yaml:"actions"`
}

// Action represents an action that can be taken in response to a message.
type Action struct {
	Title    string       `yaml:"title"`
	URI      protocol.URI `yaml:"uri"`
	External bool         `yaml:"external"`
	Save     bool         `yaml:"save"`
}

const (
	_nameKey               = "user-guidance"
	_configKey             = "guidance"
	_shownMessagesCacheDir = "ulsp/shown-messages"
)

var (
	errNeverShownMessage = fmt.Errorf("never shown message")
)

// Controller defines the interface for a guidance controller.
type Controller interface {
	StartupInfo(ctx context.Context) (ulspplugin.PluginInfo, error)
	OutputMessage(ctx context.Context, msg Message) error
	NotifyMessage(ctx context.Context, msg Message) (*Action, error)
}

// Params are inbound parameters to initialize a new plugin.
type Params struct {
	fx.In

	Sessions   session.Repository
	IdeGateway ideclient.Gateway
	Config     config.Provider
	FS         fs.UlspFS
}

type controller struct {
	sessions   session.Repository
	ideGateway ideclient.Gateway
	fs         fs.UlspFS
	guidance   guidance
}

// New creates a new controller to provide messaging with information about uLSP usage.
func New(p Params) (Controller, error) {
	controller := controller{
		sessions:   p.Sessions,
		ideGateway: p.IdeGateway,
		fs:         p.FS,
	}

	if err := p.Config.Get(_configKey).Populate(&controller.guidance); err != nil {
		return nil, fmt.Errorf("configure guidance: %w", err)
	}
	return &controller, nil
}

// StartupInfo returns PluginInfo for this controller.
func (c *controller) StartupInfo(ctx context.Context) (ulspplugin.PluginInfo, error) {
	// Set a priority for each method that this module provides.
	priorities := map[string]ulspplugin.Priority{
		protocol.MethodInitialized: ulspplugin.PriorityAsync,
	}

	// Assign method keys to implementations.
	methods := &ulspplugin.Methods{
		PluginNameKey: _nameKey,

		Initialized: c.initialized,
	}
	return ulspplugin.PluginInfo{
		Priorities: priorities,
		Methods:    methods,
		NameKey:    _nameKey,
	}, nil
}

func (c *controller) initialized(ctx context.Context, params *protocol.InitializedParams) error {
	if err := c.displayApplicableMessages(ctx); err != nil {
		return fmt.Errorf("display applicable messages: %w", err)
	}
	return nil
}

func (c *controller) displayApplicableMessages(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(c.guidance.Messages))

	for _, msg := range c.guidance.Messages {
		wg.Add(1)
		go func(msg Message) {
			defer wg.Done()

			switch msg.Kind {
			case userGuidanceMessageKindOutput:
				if err := c.OutputMessage(ctx, msg); err != nil {
					errChan <- fmt.Errorf("output message '%s': %w", msg.Key, err)
					return
				}
			case userGuidanceMessageKindNotification:
				if _, err := c.NotifyMessage(ctx, msg); err != nil {
					errChan <- fmt.Errorf("notify message '%s': %w", msg.Key, err)
					return
				}
			}
		}(msg)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) OutputMessage(ctx context.Context, msg Message) error {
	if _, err := c.statShownMessage(msg); err == nil {
		return nil
	} else if !errors.Is(err, errNeverShownMessage) {
		return fmt.Errorf("stat shown message: %w", err)
	}

	if err := c.ideGateway.LogMessage(ctx, &protocol.LogMessageParams{
		Type:    protocol.ToMessageType(msg.Type),
		Message: msg.Message,
	}); err != nil {
		return fmt.Errorf("log message: %w", err)
	}

	if err := c.markMessageAsShown(msg, nil); err != nil {
		return fmt.Errorf("mark message as shown: %w", err)
	}
	return nil
}

func (c *controller) NotifyMessage(ctx context.Context, msg Message) (*Action, error) {
	selection, err := c.statShownMessage(msg)
	if err != nil && !errors.Is(err, errNeverShownMessage) {
		return nil, fmt.Errorf("stat shown message: %w", err)
	}

	if len(msg.Actions) == 0 {
		if err == nil {
			return nil, nil
		}

		if err := c.ideGateway.ShowMessage(ctx, &protocol.ShowMessageParams{
			Type:    protocol.ToMessageType(msg.Type),
			Message: msg.Message,
		}); err != nil {
			return nil, fmt.Errorf("show message: %w", err)
		}

		if err := c.markMessageAsShown(msg, nil); err != nil {
			return nil, fmt.Errorf("mark message as shown: %w", err)
		}
		return nil, nil
	}

	if selection == nil {
		showMessageRequestParams := protocol.ShowMessageRequestParams{
			Type:    protocol.ToMessageType(msg.Type),
			Message: msg.Message,
			Actions: make([]protocol.MessageActionItem, 0, len(msg.Actions)),
		}
		for _, action := range msg.Actions {
			showMessageRequestParams.Actions = append(showMessageRequestParams.Actions, protocol.MessageActionItem{
				Title: action.Title,
			})
		}

		selection, err = c.ideGateway.ShowMessageRequest(ctx, &showMessageRequestParams)
		if err != nil {
			// Ignore cancellation error caused by client disconnecting.
			var rpcError *jsonrpc2.Error
			if errors.As(err, &rpcError) &&
				rpcError.Code == jsonrpc2.InternalError &&
				rpcError.Message == "Request window/showMessageRequest failed with message: Canceled" {
				return nil, nil
			}
			return nil, fmt.Errorf("show message request: %w", err)
		}
		if selection == nil {
			return nil, nil
		}
	}

	for _, action := range msg.Actions {
		if selection.Title != action.Title {
			continue
		}

		if action.Save {
			if err := c.markMessageAsShown(msg, selection); err != nil {
				return nil, fmt.Errorf("mark message as shown: %w", err)
			}
		}

		if action.URI == "" {
			return &action, nil
		}

		if _, err = c.ideGateway.ShowDocument(ctx, &protocol.ShowDocumentParams{
			URI:       protocol.URI(action.URI),
			External:  action.External,
			TakeFocus: true,
		}); err != nil {
			return &action, fmt.Errorf("show document: %w", err)
		}
		return &action, nil
	}

	return nil, fmt.Errorf("no action matches selection '%s'", selection.Title)
}

func (c *controller) statShownMessage(msg Message) (*protocol.MessageActionItem, error) {
	sentinelFilePath, err := c.buildSentinelFilePath(msg)
	if err != nil {
		return nil, fmt.Errorf("build sentinel file path: %w", err)
	}

	title, err := c.fs.ReadFile(sentinelFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errNeverShownMessage
	}
	if err != nil {
		return nil, fmt.Errorf("read sentinel file: %w", err)
	}

	return &protocol.MessageActionItem{Title: string(title)}, nil
}

func (c *controller) markMessageAsShown(msg Message, selection *protocol.MessageActionItem) error {
	sentinelFilePath, err := c.buildSentinelFilePath(msg)
	if err != nil {
		return fmt.Errorf("build sentinel file path: %w", err)
	}

	if err := c.fs.MkdirAll(filepath.Dir(sentinelFilePath)); err != nil {
		return fmt.Errorf("mkdir all sentinel: %w", err)
	}

	if _, err := c.fs.Create(sentinelFilePath); err != nil {
		return fmt.Errorf("create sentinel file: %w", err)
	}

	if selection == nil {
		return nil
	}

	if err := c.fs.WriteFile(sentinelFilePath, selection.Title); err != nil {
		return fmt.Errorf("write sentinel file: %w", err)
	}

	return nil
}

func (c *controller) buildSentinelFilePath(msg Message) (string, error) {
	cache, err := c.fs.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}

	return path.Join(cache, _shownMessagesCacheDir, msg.Key), nil
}
