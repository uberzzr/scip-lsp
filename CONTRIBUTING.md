# Contributing to scip-lsp

Thank you for your interest in contributing to scip-lsp! This project provides Java code intelligence through SCIP (Source Code Intelligence Protocol) indices and language server capabilities.

## Getting Started

### Prerequisites

- Bazel (use `bin/bazel` which ensures Bazelisk is installed)
- Direnv (helps keep environment up to date)
- JDK 11 or newer
- Bash shell

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/uber/scip-lsp.git
   cd scip-lsp
   ```

2. **Build the project**
   ```bash
   bazel build //...
   ```

3. **Run tests**
   ```bash
   bazel test //...
   ```

## How to Contribute

### Reporting Issues

- Use GitHub Issues to report bugs or request features
- Search existing issues before creating a new one
- Provide detailed information including:
  - Steps to reproduce
  - Expected vs actual behavior
  - Environment details (OS, Java version, Bazel version)
  - Relevant logs or error messages

### Pull Requests

1. **Fork the repository** and create a feature branch
2. **Make your changes** following the coding standards
3. **Add tests** for new functionality
4. **Ensure all tests pass** with `bazel test //...`
5. **Submit a pull request** with a clear description

### Coding Standards

- Follow existing code conventions and style
- Write clear, self-documenting code
- Add appropriate comments for complex logic
- Ensure proper error handling
- Update relevant documentation

## Project Structure

### Core Components

- **`src/main/java/com/uber/scip/aggregator/`** - SCIP aggregator implementation
  - `Aggregator.java` - Main entry point
  - `FileAnalyzer.java` - Java file analysis
  - `scip/` - SCIP-specific indexing logic

- **`src/ulsp/`** - Language server implementation
  - See [ulsp CONTRIBUTING.md](src/ulsp/CONTRIBUTING.md) for detailed guidance
  - Plugin-based architecture for extensibility

- **`src/scip-lib`** - Libraries for SCIP consumption used by the language server
  - `scanner/` - Fast implementation to quickly scan SCIP files for symbols without full proto parsing
  - `partialloader/` - Stores a prefix tree with available symbols and handles loading full occurence tables when necessary
  - `registry/` - Implementation of a registry consumable for uLSP mapping LSP methods to calls into the Partial Registry

- **`src/extension/`** - VS Code/Cursor extension

- **`bsp_server/`** - Python utilities for SCIP generation

### Build System

This project uses Bazel with Bzlmod. Key files:
- `BUILD.bazel` / `BUILD` - Build configuration files
- `MODULE.bazel` - Module dependencies
- `.bazelproject` - IDE project configuration

## Development Guidelines

### Adding New Features

1. **Plan your changes** - Discuss significant changes in an issue first
2. **Update BUILD files** when adding new source files
3. **Add tests** for new functionality
4. **Update documentation** as needed

### Testing

- Write unit tests for new code
- Use existing test patterns and frameworks
- Ensure tests are deterministic and fast
- Run the full test suite before submitting PRs

### Dependencies

- Minimize new external dependencies
- Justify any new dependencies in your PR description
- Update `MODULE.bazel` for new Maven dependencies

## Language Server Development

For contributions to the language server (`src/ulsp/`), see the [dedicated contribution guide](src/ulsp/CONTRIBUTING.md) which covers:

- Plugin architecture
- LSP method implementation
- Session management
- Configuration and feature flags

## Release Process

- Releases are managed by project maintainers
- Follow semantic versioning
- Update version files and changelogs as appropriate

## Getting Help

- Check existing documentation and issues
- Ask questions in GitHub Discussions or issues
- Be patient and respectful in all interactions

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.
