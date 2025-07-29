# ulsp-daemon Contribution Guide

## Build/Debugging

See [README](src/ulsp/README.md).

## Intro

Contributions will fall into the following primarily areas:
1) **Core Service Functionality**: Changes to the core functionality of receiving IDE requests, keeping track of distinct IDE sessions, and running the correct plugins for each request to produce a consolidated response.
    * Most of these updates will fall under the handler, gateway, and ulsp-daemon controller
    * Over time, as the protocol evolves, this will need to be updated to support new methods and use cases.
2) **Business Logic**: Changes and additions to the "business logic", such as language information, workflow integrations, and interaction with build systems and monorepo tooling.
    * Most of these updates will be contributed by adding/modifying "Plugins".  These are controllers that implement an additional Plugin interface, and define logic that is mapped to each LSP method.
    * Multiple plugins may contribute to the same LSP method, and results of each plugin will be mapped into a single response to the IDE.

## Key Concepts

- **Plugin**: A controller that includes handler methods for one or more LSP methods. Multiple plugins may support the same LSP method and contribute to a single response.
- **Session**: A single IDE connection.
  - A user may run multiple IDE sessions at once, so uLSP keeps track of distinct sessions via a SessionKey in the context.
  - This is also important for outbound notifications, as the gateway must send notifications to the correct client.
- **LSP Method**: Officially supported functionality defined by the [LSP](https://microsoft.github.io/language-server-protocol/) standard.
  - Requests will be received from the IDE client in the format defined here, and uLSP is responsible for sending responses that comply with this format.
  - uLSP may also calls to the IDE, such as to send progress or diagnostics, using the LSP client methods.

## Plugin Overview

A plugin is a controller that also implements the [Plugin](src/ulsp/entity/ulsp-plugin/ulsp_plugin.go) interface.  Plugins may operate independently or depend on each other.

### Scaffolding a new Plugin

1) Create a new controller
2) Implement StartupInfo.  This returns a priority for each method that the plugin implements, and a map of method names to implementations.
   - For an example, see PluginInfo in [quick-actions](src/ulsp/controller/quick-actions/quick_actions.go)
3) Add the new plugin as a dependency in the top-level "ulsp-daemon" controller, and add the plugin to [availablePlugins](src/ulsp/controller/ulsp-daemon/ulsp_daemon.go).
4) Implement any relevant methods.  Setup logic should be placed in Initialize, and use EndSession to perform any cleanup following session exit.
   - Client behavior may be controlled by the InitializationResult. If your plugin needs anything specific set in the InitializationResult, use mappers to modify the InitializeResult with relevant updates.
5) If your method needs to contribute to the response, use mappers as opposed to editing the response directly.  This will ensure the response is generated in a consistent manner.

### Plugin Workflow Overview

1) Service launch - Plugins are made available as FX dependencies.
2) Session initialization - Each plugin has its StartupInfo method called, providing the methods that it supports and their priorities.
3) Method Prioritization - Based on the StartupInfo results, a prioritized method list is generated, including only plugins that are determined to be enabled for that session.
3) Subsequent LSP method calls - During each call, methods for a plugin are executed in order based on their defined prioritization.
4) Session exit - connection is closed, session deleted, and runtime method prioritization for that session is cleared.

### Plugin Prioritization

Plugins may define a priority for each method. At session initialization, a prioritization list is generated based only on the enabled plugins for that session.
When a call is received for a given method, plugins will be called based on this ordering, one by one, until a completed result is ready to be returned.

For longer running tasks that don't need to contribute to the result but are triggered by an LSP method, PriorityAsync can be used.  These won't block the response and may continue running after it has been returned.

### Adding Configuration Keys

Controllers may define their own configuration keys. Ensure that they key is checked during that controller's New function to avoid errors during launch.

#### Feature Flags

Devpod features are checked on launch and used to override parts of the configuration if true.
See [feature-flags.go](src/ulsp/app/feature_flags.go) for examples of using a feature flag to control configuration behavior.

### LSP Library Updates

- GitHub source: https://github.com/go-language-server/protocol
- This can be upgraded in the Go Monorepo to get updated types
