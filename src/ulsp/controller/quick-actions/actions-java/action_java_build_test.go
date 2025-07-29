package quickactions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	action "github.com/uber/scip-lsp/src/ulsp/controller/quick-actions/action"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/gateway/ide-client/ideclientmock"
	"github.com/uber/scip-lsp/src/ulsp/internal/executor/executormock"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
)

func TestJavaBuildExecute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	a := ActionJavaBuild{}

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	s := &entity.Session{
		UUID: factory.UUID(),
		InitializeParams: &protocol.InitializeParams{
			ClientInfo: &protocol.ClientInfo{
				Name: "Visual Studio Code",
			},
		},
	}
	s.WorkspaceRoot = "/home/user/fievel"
	s.Monorepo = entity.MonorepoNameJava

	t.Run("success", func(t *testing.T) {
		executorMock := executormock.NewMockExecutor(ctrl)
		ideGatewayMock := ideclientmock.NewMockGateway(ctrl)
		c := &action.ExecuteParams{
			IdeGateway: ideGatewayMock,
			Sessions:   sessionRepository,
			Executor:   executorMock,
		}

		var writer bytes.Buffer
		ideGatewayMock.EXPECT().LogMessage(gomock.Any(), gomock.Any()).Return(nil)
		ideGatewayMock.EXPECT().GetLogMessageWriter(gomock.Any(), gomock.Any()).Return(&writer, nil)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Return(nil)

		assert.NoError(t, a.Execute(ctx, c, []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/fievel/tooling/intellij/uber-intellij-plugin-core/src/main/java/com/uber/intellij/bazel/BazelSyncListener.java"}}`)))
	})

	t.Run("bad args", func(t *testing.T) {
		executorMock := executormock.NewMockExecutor(ctrl)
		ideGatewayMock := ideclientmock.NewMockGateway(ctrl)
		c := &action.ExecuteParams{
			IdeGateway: ideGatewayMock,
			Sessions:   sessionRepository,
			Executor:   executorMock,
		}
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		assert.Error(t, a.Execute(ctx, c, []byte(`{"brokenJSON`)))
	})

	t.Run("log message failure", func(t *testing.T) {
		executorMock := executormock.NewMockExecutor(ctrl)
		ideGatewayMock := ideclientmock.NewMockGateway(ctrl)
		c := &action.ExecuteParams{
			IdeGateway: ideGatewayMock,
			Sessions:   sessionRepository,
			Executor:   executorMock,
		}

		var writer bytes.Buffer
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		ideGatewayMock.EXPECT().GetLogMessageWriter(gomock.Any(), gomock.Any()).Return(&writer, nil)
		ideGatewayMock.EXPECT().LogMessage(gomock.Any(), gomock.Any()).Return(errors.New("error"))

		assert.Error(t, a.Execute(ctx, c, []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/fievel/roadrunner/application-dw/src/test/java/com/uber/roadrunner/application/exception/GatewayErrorExceptionMapperTest.java"}}`)))
	})

	t.Run("writer failure", func(t *testing.T) {
		executorMock := executormock.NewMockExecutor(ctrl)
		ideGatewayMock := ideclientmock.NewMockGateway(ctrl)
		c := &action.ExecuteParams{
			IdeGateway: ideGatewayMock,
			Sessions:   sessionRepository,
			Executor:   executorMock,
		}

		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		ideGatewayMock.EXPECT().GetLogMessageWriter(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
		assert.Error(t, a.Execute(ctx, c, []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/fievel/tooling/intellij/uber-intellij-plugin-core/src/main/java/com/uber/intellij/bazel/BazelSyncListener.java"}}`)))
	})

	t.Run("execution failure", func(t *testing.T) {
		executorMock := executormock.NewMockExecutor(ctrl)
		ideGatewayMock := ideclientmock.NewMockGateway(ctrl)
		c := &action.ExecuteParams{
			IdeGateway: ideGatewayMock,
			Sessions:   sessionRepository,
			Executor:   executorMock,
		}

		var writer bytes.Buffer
		ideGatewayMock.EXPECT().GetLogMessageWriter(gomock.Any(), gomock.Any()).Return(&writer, nil)
		ideGatewayMock.EXPECT().LogMessage(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Return(errors.New("error"))

		assert.Error(t, a.Execute(ctx, c, []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/fievel/tooling/intellij/uber-intellij-plugin-core/src/main/java/com/uber/intellij/bazel/BazelSyncListener.java"}}`)))
	})

}

func TestJavaBuildProcessDocument(t *testing.T) {
	a := ActionJavaBuild{}
	doc := protocol.TextDocumentItem{
		URI:        "file:///MyExampleTest.java",
		LanguageID: "java",
		Text: `package com.uber.rider.growth.jobs;

import com.uber.fievel.testing.base.FievelTestBase;
import org.junit.Test;

public class MyExampleTest extends FievelTestBase {

	@Test
	public void myTestMethod() throws Exception {}
}

	`}
	results, err := a.ProcessDocument(context.Background(), doc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	args := results[0].(protocol.CodeLens).Command.Arguments[0].(argsJavaBuild)
	assert.Equal(t, doc.URI, args.Document.URI)
	assert.Equal(t, "MyExampleTest", args.ClassName)
}

func TestJavaBuildCommandName(t *testing.T) {
	a := ActionJavaBuild{}
	cmd := a.CommandName()
	assert.True(t, len(cmd) > 0)
}

func TestJavaBuildShouldEnable(t *testing.T) {
	a := ActionJavaBuild{}
	s := &entity.Session{
		UUID: factory.UUID(),
		InitializeParams: &protocol.InitializeParams{
			ClientInfo: &protocol.ClientInfo{
				Name: "Visual Studio Code",
			},
		},
	}
	assert.False(t, a.ShouldEnable(s))

	s.Monorepo = entity.MonorepoNameJava
	assert.True(t, a.ShouldEnable(s))
}

func TestJavaBuildIsRelevantDocument(t *testing.T) {
	a := ActionJavaBuild{}

	relevantDoc := protocol.TextDocumentItem{URI: "file:///test.java", LanguageID: "java"}
	assert.True(t, a.IsRelevantDocument(nil, relevantDoc))

	irrelevantDoc := protocol.TextDocumentItem{URI: "file:///test.go", LanguageID: "go"}
	assert.False(t, a.IsRelevantDocument(nil, irrelevantDoc))
}

func TestJavaBuildProvideWorkDoneProgressParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	a := ActionJavaBuild{}

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	s := &entity.Session{
		UUID: factory.UUID(),
		InitializeParams: &protocol.InitializeParams{
			ClientInfo: &protocol.ClientInfo{
				Name: "Visual Studio Code",
			},
		},
	}
	s.WorkspaceRoot = "/home/user/fievel"
	s.Monorepo = entity.MonorepoNameJava

	rawArgs := []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/fievel/tooling/src/main/java/com/uber/intellij/bazel/BazelSync.java"}}`)

	execParams := &action.ExecuteParams{
		Sessions: sessionRepository,
	}
	t.Run("Success", func(t *testing.T) {
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		params, err := a.ProvideWorkDoneProgressParams(ctx, execParams, rawArgs)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Equal(t, fmt.Sprintf(_titleJavaBuildProgressBar, "BazelSync.java"), params.Title)
		assert.Equal(t, fmt.Sprintf(_messageJavaBuildProgressBar, "tooling/..."), params.Message)
	})

	t.Run("Context Error", func(t *testing.T) {
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(nil, errors.New("ctx error"))
		_, err := a.ProvideWorkDoneProgressParams(ctx, execParams, rawArgs)
		assert.Error(t, err, "context error should be thrown")
	})

	t.Run("missing src", func(t *testing.T) {
		invalidPathArgs := []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/fievel"}}`)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		_, err := a.ProvideWorkDoneProgressParams(ctx, execParams, invalidPathArgs)
		assert.Error(t, err)

	})

	t.Run("missing workspace root", func(t *testing.T) {
		invalidPathArgs := []byte(`{"interfaceName": "myInterface", "document": {"uri": "file:///home/user/src/bazel/BazelSync.java"}}`)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)
		_, err := a.ProvideWorkDoneProgressParams(ctx, execParams, invalidPathArgs)
		assert.Error(t, err)

	})
}

func TestJavaBuildPrepareCommandAndEnv(t *testing.T) {
	a := ActionJavaBuild{}
	ctx := context.Background()
	var writer bytes.Buffer
	workspaceRoot := "/home/user/fievel"
	target := "//a/b/c:src_main"

	testCases := []struct {
		name            string
		ideClient       string
		expectedToolTag string
	}{
		{
			name:            "vscode",
			ideClient:       "Visual Studio Code",
			expectedToolTag: "--tool_tag=vscode:actions:build",
		},
		{
			name:            "cursor",
			ideClient:       "Cursor",
			expectedToolTag: "--tool_tag=cursor:actions:build",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, env := a.prepareCommandAndEnv(ctx, &writer, workspaceRoot, target, tc.ideClient)

			assert.Equal(t, "/home/user/fievel/tools/bazel", cmd.Path)
			assert.Equal(t, []string{_bazelBuild, target, tc.expectedToolTag}, cmd.Args[1:])
			assert.Equal(t, workspaceRoot, cmd.Dir)

			// Verify environment variables
			assert.Contains(t, env, "WORKSPACE_ROOT="+workspaceRoot)
			assert.Contains(t, env, "PROJECT_ROOT="+workspaceRoot)
		})
	}
}
