package handler

import (
	_ "github.com/uber/scip-lsp/idl/ulsp/service"
	controller "github.com/uber/scip-lsp/src/ulsp/controller"
	ulspdaemon "github.com/uber/scip-lsp/src/ulsp/controller/ulsp-daemon"
	handler "github.com/uber/scip-lsp/src/ulsp/handler/ulsp-daemon"
	"github.com/uber/scip-lsp/src/ulsp/repository/session"
	"go.uber.org/fx"
)

// Module provides the ulsp-daemon server into an Fx application.
var Module = fx.Options(
	controller.Module,
	fx.Provide(session.New),
	fx.Provide(handler.New),
	fx.Invoke(outputYARPCConnectionInfo),
	fx.Invoke(func(m handler.Handler) {}),
	fx.Invoke(func(m ulspdaemon.Controller) {}),
)
