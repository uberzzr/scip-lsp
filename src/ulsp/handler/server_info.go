package handler

import (
	"fmt"

	"github.com/uber/scip-lsp/src/ulsp/internal/serverinfofile"
	"go.uber.org/config"
)

const (
	_errInvalidEntry = "type error or missing field for key %q"

	_fmtInfoFileKey = "%s-%s"

	_configKeyInbounds = "inbounds"
	_configKeyAddress  = "address"
)

// Output address info from the yarpc configuration block.
// Other connection methods (e.g. JSON-RPC) that don't use this configuration may independently add their fields to the Server Info file.
func outputYARPCConnectionInfo(cfg config.Provider, infofile serverinfofile.ServerInfoFile) error {
	var cfgData map[string]interface{}
	if err := cfg.Get("yarpc").Populate(&cfgData); err != nil {
		return fmt.Errorf("loading yarpc config: %v", err)
	}

	inbounds, ok := cfgData[_configKeyInbounds].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf(_errInvalidEntry, _configKeyInbounds)
	}

	for currentKey, currentInboundRaw := range inbounds {
		currentInboundValue, ok := currentInboundRaw.(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf(_errInvalidEntry, currentKey)
		}
		address, ok := currentInboundValue[_configKeyAddress].(string)
		if !ok {
			return fmt.Errorf(_errInvalidEntry, _configKeyAddress)
		}
		if err := infofile.UpdateField(fmt.Sprintf(_fmtInfoFileKey, currentKey, _configKeyAddress), address); err != nil {
			return fmt.Errorf("outputting %q address to info file: %w", currentKey, err)
		}
	}

	return nil
}
