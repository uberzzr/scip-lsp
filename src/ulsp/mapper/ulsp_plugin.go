package mapper

import (
	"fmt"

	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
)

// PluginInfoToRuntimePrioritizedMethods maps all PluginInfo from running plugins, into a prioritized list of modules to run per method.
func PluginInfoToRuntimePrioritizedMethods(allPluginInfo []ulspplugin.PluginInfo) (ulspplugin.RuntimePrioritizedMethods, error) {
	result := make(ulspplugin.RuntimePrioritizedMethods)
	methodPriorityBuckets := make(map[string]map[ulspplugin.Priority][]*ulspplugin.Methods)

	for _, pluginInfo := range allPluginInfo {
		if err := pluginInfo.Validate(); err != nil {
			return nil, fmt.Errorf("error validating plugin configuration: %w", err)
		}

		// Add this plugin to its assigned priority bucket for each method.
		for method, priority := range pluginInfo.Priorities {
			if _, ok := methodPriorityBuckets[method]; !ok {
				methodPriorityBuckets[method] = make(map[ulspplugin.Priority][]*ulspplugin.Methods)
			}
			methodPriorityBuckets[method][priority] = append(methodPriorityBuckets[method][priority], pluginInfo.Methods)
		}
	}

	// Consolidate the final buckets into two slices (sync and async) ordered for execution.
	for method, buckets := range methodPriorityBuckets {
		for priority := ulspplugin.PriorityHigh; priority <= ulspplugin.PriorityAsync; priority++ {
			current, ok := result[method]
			if !ok {
				current = ulspplugin.MethodLists{}
			}
			if priority < ulspplugin.PriorityAsync {
				current.Sync = append(result[method].Sync, buckets[priority]...)
			} else {
				current.Async = append(result[method].Async, buckets[priority]...)
			}
			result[method] = current
		}
	}

	return result, nil
}
