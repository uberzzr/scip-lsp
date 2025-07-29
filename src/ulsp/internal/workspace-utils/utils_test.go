package workspaceutils

import (
	"context"
	"errors"
	"net/url"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	"github.com/uber/scip-lsp/src/ulsp/gateway/ide-client/ideclientmock"
	"github.com/uber/scip-lsp/src/ulsp/internal/executor/executormock"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs/fsmock"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const monorepoNameSample entity.MonorepoName = "sample"

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	assert.NotPanics(t, func() {
		New(Params{
			IdeGateway: ideclientmock.NewMockGateway(ctrl),
			Logger:     zap.NewNop().Sugar(),
			FS:         fsmock.NewMockUlspFS(ctrl),
		})
	})
}

func TestGetWorkspaceRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	fsMock := fsmock.NewMockUlspFS(ctrl)
	ideClientMock := ideclientmock.NewMockGateway(ctrl)

	c := workspaceUtilsImpl{
		logger:     zap.NewNop().Sugar(),
		ideGateway: ideClientMock,
		fs:         fsMock,
	}

	t.Run("valid workspace folders", func(t *testing.T) {
		sampleRoot := "/sample/root/" + string(monorepoNameSample)
		workspaceFolders := []protocol.WorkspaceFolder{
			{
				URI: "file://" + sampleRoot + "/foo/bar",
			},
			{
				URI: "file://" + sampleRoot + "/foo",
			},
		}

		fsMock.EXPECT().WorkspaceRoot(gomock.Any()).Return([]byte(sampleRoot), nil).Times(len(workspaceFolders))
		result, err := c.GetWorkspaceRoot(ctx, workspaceFolders)

		assert.NoError(t, err)
		assert.Equal(t, sampleRoot, result)
	})

	t.Run("conflicting roots", func(t *testing.T) {
		sampleRoots := []string{"/sample/root/" + string(monorepoNameSample), "/another/root"}
		workspaceFolders := []protocol.WorkspaceFolder{
			{
				URI: "file://" + sampleRoots[0] + "/foo/bar",
			},
			{
				URI: "file://" + sampleRoots[1] + "/foo/bar",
			},
		}

		for i := range workspaceFolders {
			folderPath, _ := url.Parse(workspaceFolders[i].URI)
			fsMock.EXPECT().WorkspaceRoot(folderPath.Path).Return([]byte(sampleRoots[i]), nil)
		}

		ideClientMock.EXPECT().ShowMessage(gomock.Any(), gomock.Any()).Return(nil)
		result, err := c.GetWorkspaceRoot(ctx, workspaceFolders)

		assert.NoError(t, err)
		assert.Equal(t, sampleRoots[0], result)
	})

	t.Run("invalid uri", func(t *testing.T) {
		sampleRoot := "/sample/root"
		workspaceFolders := []protocol.WorkspaceFolder{
			{
				URI: "file://" + sampleRoot + "/foo%2Gbar",
			},
			{
				URI: "file://" + sampleRoot + "/foo",
			},
		}

		folderPath, _ := url.Parse(workspaceFolders[1].URI)
		fsMock.EXPECT().WorkspaceRoot(folderPath.Path).Return([]byte(sampleRoot), nil)
		_, err := c.GetWorkspaceRoot(ctx, workspaceFolders)

		assert.NoError(t, err)
	})

	t.Run("no workspace found", func(t *testing.T) {
		sampleRoot := "/sample/root"
		workspaceFolders := []protocol.WorkspaceFolder{
			{
				URI: "file://" + sampleRoot + "/foo%2Gbar",
			},
			{
				URI: "file://" + sampleRoot + "/foo",
			},
		}

		folderPath, _ := url.Parse(workspaceFolders[1].URI)
		fsMock.EXPECT().WorkspaceRoot(folderPath.Path).Return(nil, errors.New("sample"))
		_, err := c.GetWorkspaceRoot(ctx, workspaceFolders)

		assert.Error(t, err)
	})

	t.Run("no folders", func(t *testing.T) {
		workspaceFolders := []protocol.WorkspaceFolder{}
		_, err := c.GetWorkspaceRoot(ctx, workspaceFolders)

		assert.Error(t, err)
	})
}

func TestGetRepoName(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	executorMock := executormock.NewMockExecutor(ctrl)
	sampleRoot := "/sample/root/" + string(monorepoNameSample)

	c := workspaceUtilsImpl{
		logger:   zap.NewNop().Sugar(),
		executor: executorMock,
	}

	t.Run("valid repo name", func(t *testing.T) {
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Do(func(cmd *exec.Cmd, env []string) error {
			cmd.Stdout.Write([]byte("sample@code.example.internal:" + monorepoNameSample))
			return nil
		})

		result, err := c.GetRepoName(ctx, sampleRoot)

		assert.NoError(t, err)
		assert.Equal(t, monorepoNameSample, result)
	})

	t.Run("error running command", func(t *testing.T) {
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Return(errors.New("sample"))
		_, err := c.GetRepoName(ctx, sampleRoot)

		assert.Error(t, err)
	})

	t.Run("error parsing path", func(t *testing.T) {
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Do(func(cmd *exec.Cmd, env []string) error {
			cmd.Stdout.Write([]byte("sample@code.example.internal-" + monorepoNameSample))
			return nil
		})

		_, err := c.GetRepoName(ctx, sampleRoot)
		assert.Error(t, err)
	})

}

func TestGetEnv(t *testing.T) {
	t.Run("valid env", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		dir := "/sample/dir"

		sampleRawOutput := "KEY1=VAL1\nKEY2=VAL2"
		expectedResult := []string{"KEY1=VAL1", "KEY2=VAL2"}

		executorMock := executormock.NewMockExecutor(ctrl)
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Do(func(cmd *exec.Cmd, env []string) error {
			assert.Equal(t, dir, cmd.Dir)
			cmd.Stdout.Write([]byte(sampleRawOutput))
			return nil
		})

		c := workspaceUtilsImpl{
			executor: executorMock,
		}

		result, err := c.GetEnv(ctx, dir)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedResult, result)
	})

	t.Run("direnv denied", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		dir := "/sample/dir"

		executorMock := executormock.NewMockExecutor(ctrl)
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).DoAndReturn(func(cmd *exec.Cmd, env []string) error {
			assert.Equal(t, dir, cmd.Dir)
			cmd.Stderr.Write([]byte(_dirEnvDeniedSubstring))
			return errors.New("sample")
		})

		c := workspaceUtilsImpl{
			executor: executorMock,
		}

		_, err := c.GetEnv(ctx, dir)
		assert.Error(t, err)
	})

	t.Run("other exec error", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		dir := "/sample/dir"

		executorMock := executormock.NewMockExecutor(ctrl)
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Return(errors.New("sample"))

		c := workspaceUtilsImpl{
			executor: executorMock,
		}

		_, err := c.GetEnv(ctx, dir)
		assert.Error(t, err)
	})

}
