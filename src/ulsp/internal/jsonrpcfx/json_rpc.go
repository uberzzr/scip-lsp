package jsonrpcfx

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gofrs/uuid"
	"github.com/uber/scip-lsp/src/ulsp/internal/serverinfofile"
	"go.lsp.dev/jsonrpc2"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	_configKeyAddress = "jsonrpc.address"
	_outputKey        = "lsp-address"
)

// Module is an fx module to handle JSON-RPC requests.
var Module = fx.Provide(New)

// JSONRPCModule represents a module to manage JSON-RPC requests.
type JSONRPCModule interface {
	OnStart(ctx context.Context) error
	ServeStream(ctx context.Context, conn jsonrpc2.Conn) error
	RegisterConnectionManager(connectionManager ConnectionManager) error
}

// Router serves as the interface through which handling of requests will be implemented.
type Router interface {
	HandleReq(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error
	UUID() uuid.UUID
}

// ConnectionManager will manage each active connection and its corresponding Router throughout the lifecycle of a connection.
type ConnectionManager interface {
	NewConnection(ctx context.Context, conn *jsonrpc2.Conn) (router Router, err error)
	RemoveConnection(ctx context.Context, id uuid.UUID)
}

type module struct {
	Address string `json:"address"`

	connectionMgr  ConnectionManager
	ln             *net.TCPListener
	logger         *zap.SugaredLogger
	serverInfoFile serverinfofile.ServerInfoFile
}

// Params define values to be used by JsonRpcHandler.
type Params struct {
	fx.In

	Config         config.Provider
	Lifecycle      fx.Lifecycle
	Logger         *zap.SugaredLogger
	ServerInfoFile serverinfofile.ServerInfoFile
}

// New creates a new server to handle JSON-RPC requests on the given port and host.
func New(p Params) (JSONRPCModule, error) {
	if p.Lifecycle == nil || p.Config == nil {
		return nil, errors.New("required parameters are missing")
	}

	m := module{
		logger:         p.Logger,
		serverInfoFile: p.ServerInfoFile,
	}

	if err := m.processConfig(p.Config); err != nil {
		return nil, err
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: m.OnStart,
	})

	return &m, nil
}

// OnStart will initialize a JSON-RPC handler and then begin handling incoming connections.
func (m *module) OnStart(ctx context.Context) error {
	if err := m.setup(); err != nil {
		return err
	}

	go m.start()
	return nil
}

// ServeStream is called when a new connection is initiated. Requests received via the connection will be routed to the handler, and answered via the connection's replier.
func (m *module) ServeStream(ctx context.Context, conn jsonrpc2.Conn) error {
	if m.connectionMgr == nil {
		m.logger.Errorf("cannot serve connection, no connection manager set")
		return errors.New("cannot serve connection, no connection manager set")
	}

	// Start handling the connection.
	handler, err := m.connectionMgr.NewConnection(ctx, &conn)
	if err != nil {
		return err
	}
	m.logger.Infow("client connected", zap.Stringer("uuid", handler.UUID()))
	conn.Go(ctx, handler.HandleReq)

	// Block indefinitely until connection closed.
	<-conn.Done()

	// Cleanup after connection.
	m.connectionMgr.RemoveConnection(ctx, handler.UUID())
	m.logger.Infow("client disconnected", zap.Stringer("uuid", handler.UUID()))

	return conn.Err()
}

// RegisterConnectionManager sets the connection manager, which keeps track of current active connections and provides a Router implementation.
func (m *module) RegisterConnectionManager(connectionMgr ConnectionManager) error {
	if m.connectionMgr != nil {
		return errors.New("cannot register a duplicate connection manager")
	}
	m.connectionMgr = connectionMgr
	return nil
}

// setup should be called after creation of a new handler to set initial values.
func (m *module) setup() error {
	if m.Address == "" {
		return errors.New("setup called before address is set")
	}

	addr, err := net.ResolveTCPAddr("tcp", m.Address)
	if err != nil {
		return err
	}

	m.ln, err = net.ListenTCP("tcp", addr)
	return err
}

// start will begin serving connections, and panic on error.
func (m *module) start() {
	if err := m.serverInfoFile.UpdateField(_outputKey, m.Address); err != nil {
		panic(err)
	}

	// TODO(IDE-757): Adjusting this to Warn for now, so we can reduce the log level and still receive this message.
	// It can be removed once adjustments have been made to the client to not depend on this message.
	m.logger.Warnw("started JSON-RPC inbound", zap.String("address", m.Address))
	if err := jsonrpc2.Serve(context.Background(), m.ln, m, 0); err != nil {
		panic(err)
	}
}

// processConfig will parse the configuration for any values required by this module.
func (m *module) processConfig(cfg config.Provider) error {
	val := cfg.Get(_configKeyAddress)
	if err := val.Populate(&m.Address); err != nil {
		// incorrectly formatted config
		return fmt.Errorf("getting config field %q: %w", _configKeyAddress, err)
	}

	if m.Address == "" {
		// yaml is missing either the key or value
		return fmt.Errorf("missing field %q in config", _configKeyAddress)
	}

	return nil
}
