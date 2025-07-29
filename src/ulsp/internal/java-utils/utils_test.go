package javautils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestJavaUtilsGetJavaTarget(t *testing.T) {
	workspaceRoot := "/home/user/fievel"
	t.Run("valid src", func(t *testing.T) {
		validDoc := protocol.DocumentURI("file:///home/user/fievel/tooling/intellij/uber-intellij-plugin-core/src/main/java/com/uber/intellij/bazel/BazelSyncListener.java")
		target, err := GetJavaTarget(workspaceRoot, validDoc)
		assert.NoError(t, err)
		assert.Equal(t, "tooling/intellij/uber-intellij-plugin-core/...", target)
	})

	t.Run("valid test doc", func(t *testing.T) {
		validDoc := protocol.DocumentURI("file:///home/user/fievel/roadrunner/application-dw/src/test/java/com/uber/roadrunner/application/exception/GatewayErrorExceptionMapperTest.java")
		target, err := GetJavaTarget(workspaceRoot, validDoc)
		assert.NoError(t, err)
		assert.Equal(t, "roadrunner/application-dw/...", target)
	})

	t.Run("missing workspace root", func(t *testing.T) {
		invalidDoc := protocol.DocumentURI("file:///home/roadrunner/application-dw/src/test/GatewayErrorExceptionMapperTest.java")
		_, err := GetJavaTarget(workspaceRoot, invalidDoc)
		assert.Error(t, err)
	})

	t.Run("missing src", func(t *testing.T) {
		invalidDoc := protocol.DocumentURI("file:///home/user/fievel/roadrunner/GatewayErrorExceptionMapperTest.java")
		_, err := GetJavaTarget(workspaceRoot, invalidDoc)
		assert.Error(t, err)
	})
}

func TestNormalizeIDEClient(t *testing.T) {
	assert.Equal(t, "vscode", NormalizeIDEClient("Visual Studio Code"))
	assert.Equal(t, "cursor", NormalizeIDEClient("Cursor"))
	assert.Equal(t, "intellij", NormalizeIDEClient("intellij"))
}
