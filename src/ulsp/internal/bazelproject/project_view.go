package bazelproject

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
)

// PresetPatternsProvider provides a list of preset target patterns to the registry
type PresetPatternsProvider interface {
	GetPresetPatterns(ctx context.Context, getTargetsFunc func(ctx context.Context, patterns []string) ([]string, error)) ([]string, error)
}

// Interface compliance checks.
var _ PresetPatternsProvider = (*projectViewPatternsProvider)(nil)

// Multiple requests within this window will reuse the previous result.
const _reuseTimeout = time.Second * 60

// Params is the configuration for the provider.
type Params struct {
	FS            fs.UlspFS
	OutputWriter  io.Writer
	WorkspaceRoot string

	// Paths, relative to the workspace root, which will be checked.
	// Once the first match is found, it will be used.
	Paths []string
}

type projectViewPatternsProvider struct {
	mu sync.Mutex

	fs            fs.UlspFS
	output        io.Writer
	workspaceRoot string

	// Paths, relative to the workspace root, which will be checked.
	// Once the first match is found, it will be used.
	paths []string

	// For consistent behavior, the same source file will be used once found.
	selectedSource string

	result      []string
	lastChecked time.Time
}

// New creates a new project view patterns provider.
func New(p Params) PresetPatternsProvider {
	return &projectViewPatternsProvider{
		fs:            p.FS,
		output:        p.OutputWriter,
		workspaceRoot: p.WorkspaceRoot,
		paths:         p.Paths,
	}
}

func (p *projectViewPatternsProvider) GetPresetPatterns(ctx context.Context, getTargetsFunc func(ctx context.Context, patterns []string) ([]string, error)) ([]string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	defer func() { p.lastChecked = time.Now() }()

	// First check or any other within the timeout window.
	if time.Since(p.lastChecked) < _reuseTimeout {
		return p.result, nil
	}

	p.setSource(ctx)
	if p.selectedSource == "" {
		return nil, nil
	}

	patterns, err := p.getPatterns(p.selectedSource)
	if err != nil {
		p.output.Write([]byte(fmt.Sprintf("Error while checking project configuration: %s", err)))
		return nil, err
	}

	p.output.Write([]byte(fmt.Sprintf("Found the following user-configured project scope: [%s]", strings.Join(patterns, ", "))))
	p.result, err = getTargetsFunc(ctx, patterns)
	return p.result, err
}

func (p *projectViewPatternsProvider) getPatterns(absPath string) ([]string, error) {
	file, err := p.fs.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// TODO(IDE-1226): Currently this will parse the targets from a single file.
	// To traverse imported project view files and derive targets from directories, we can move existing logic from git-bzl into a shared library.
	project, err := ParseProjectView(file)
	if err != nil {
		return nil, err
	}

	return project.Targets, nil
}

func (p *projectViewPatternsProvider) setSource(ctx context.Context) {
	if p.selectedSource != "" {
		// If the source is already set, just confirm that it still exists.
		// Search again only if has been deleted.
		ok, err := p.fs.FileExists(p.selectedSource)
		if ok && err == nil {
			return
		}
		p.selectedSource = ""
	}

	p.output.Write([]byte(fmt.Sprintf("Checking the following paths for Bazel project configuration: [%s]", strings.Join(p.paths, ", "))))
	for _, current := range p.paths {
		absPath := path.Join(p.workspaceRoot, current)
		ok, err := p.fs.FileExists(absPath)
		if ok {
			p.output.Write([]byte(fmt.Sprintf("Selected Bazel project configuration: %s", current)))
			p.selectedSource = absPath
			return
		} else if err != nil {
			p.output.Write([]byte(fmt.Sprintf("Skipping %s due to error: %s", current, err)))
		}
	}
	p.output.Write([]byte("No matches, proceeding without a Bazel project configuration."))
}
