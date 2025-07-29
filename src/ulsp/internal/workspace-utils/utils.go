package workspaceutils

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/uber/scip-lsp/src/ulsp/entity"
	ideclient "github.com/uber/scip-lsp/src/ulsp/gateway/ide-client"
	"github.com/uber/scip-lsp/src/ulsp/internal/executor"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
	"go.lsp.dev/protocol"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const _dirEnvDeniedSubstring = "Run `direnv allow` to approve its content"

// Module provides a new WorkspaceUtils.
var Module = fx.Provide(New)

// WorkspaceUtils is a utility interface for getting workspace related information.
type WorkspaceUtils interface {
	GetWorkspaceRoot(ctx context.Context, workspaceFolders []protocol.WorkspaceFolder) (string, error)
	GetRepoName(ctx context.Context, dir string) (entity.MonorepoName, error)
	GetEnv(ctx context.Context, dir string) ([]string, error)
}

// Params are the parameters required to create a new WorkspaceUtils.
type Params struct {
	fx.In

	IdeGateway ideclient.Gateway
	Logger     *zap.SugaredLogger
	FS         fs.UlspFS
	Executor   executor.Executor
}

type workspaceUtilsImpl struct {
	ideGateway ideclient.Gateway
	logger     *zap.SugaredLogger
	fs         fs.UlspFS
	executor   executor.Executor
}

// New creates a new WorkspaceUtils.
func New(p Params) WorkspaceUtils {
	return &workspaceUtilsImpl{
		ideGateway: p.IdeGateway,
		logger:     p.Logger,
		fs:         p.FS,
		executor:   p.Executor,
	}
}

func (c *workspaceUtilsImpl) GetWorkspaceRoot(ctx context.Context, workspaceFolders []protocol.WorkspaceFolder) (string, error) {
	if len(workspaceFolders) == 0 {
		return "", fmt.Errorf("no workspace folders provided")
	}

	// Find the workspace root and look for any conflicting folders.
	result := ""
	for _, folder := range workspaceFolders {
		// code-workspace files may contain improperly formatted or nonexistent folders.
		// only return an error if no workspace root can be found among any of the given folders.
		fileSystemPath, err := url.Parse(folder.URI)
		if err != nil {
			continue
		}

		out, err := c.fs.WorkspaceRoot(fileSystemPath.Path)
		if err != nil {
			continue
		}

		if result == "" {
			// First result will be used as the workspace root.
			result = string(out)
		} else if result != string(out) {
			// Any difference in subsequent results will be considered a conflict.
			msg := fmt.Sprintf("Workspace root is %q, but a folder in %q is also included. Please remove %q from this workspace and launch it in its own IDE session.", result, string(out), folder.URI)
			c.ideGateway.ShowMessage(ctx, &protocol.ShowMessageParams{
				Type:    protocol.MessageTypeWarning,
				Message: msg,
			})
			c.logger.Warn(msg)
			break
		}
	}

	if result == "" {
		folderStrings := []string{}
		for _, folder := range workspaceFolders {
			folderStrings = append(folderStrings, folder.URI)
		}

		return "", fmt.Errorf("unable to determine a workspace root among the following searched folders: %v", strings.Join(folderStrings, ", "))
	}

	return strings.TrimSpace(string(result)), nil
}

func (c *workspaceUtilsImpl) GetRepoName(ctx context.Context, dir string) (entity.MonorepoName, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir

	err := c.executor.RunCommand(cmd, os.Environ())
	if err != nil {
		return "", fmt.Errorf("getting git remote url: %w", err)
	}

	repoRemoteURL := strings.TrimSpace(stdout.String())
	c.logger.Infof("git remote url: %s", repoRemoteURL)
	repoURLSegments := strings.Split(repoRemoteURL, ":")
	if len(repoURLSegments) != 2 {
		return "", fmt.Errorf("invalid git remote url: %s", repoRemoteURL)
	}

	repo := entity.MonorepoName(repoURLSegments[1])
	return repo, nil
}

func (c *workspaceUtilsImpl) GetEnv(ctx context.Context, dir string) ([]string, error) {
	getEnvScript := `
		set -e
		output=$(direnv export bash)
		eval "$output"
		env
	`
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "bash", "-c", getEnvScript)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := c.executor.RunCommand(cmd, os.Environ())
	if err != nil {
		if strings.Contains(stderr.String(), _dirEnvDeniedSubstring) {
			return nil, fmt.Errorf("please run `direnv allow` inside %s to approve the latest .envrc content, then reload this window", dir)
		}

		return nil, fmt.Errorf("running direnv: %w", err)
	}

	result := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	return result, nil
}
