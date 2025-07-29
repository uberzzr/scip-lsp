package javautils

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"go.lsp.dev/protocol"
)

const _targetSuffix = "..."
const _visualStudioCodeIDEClient = "Visual Studio Code"
const _cursorIDEClient = "Cursor"

// GetJavaTarget returns the target path for the given document URI.
func GetJavaTarget(workspaceRoot string, docURI protocol.DocumentURI) (string, error) {
	filePath := docURI.Filename()
	pathSegments := strings.SplitN(filePath, "/src/", 2)
	if len(pathSegments) != 2 {
		return "", fmt.Errorf("missing src directory")
	}

	targetSegments := strings.SplitN(pathSegments[0], workspaceRoot+string(filepath.Separator), 2)
	if len(targetSegments) != 2 {
		return "", fmt.Errorf("missing workspace root")
	}

	target := path.Join(targetSegments[1], _targetSuffix)
	return target, nil
}

// NormalizeIDEClient normalizes the IDE client name to a consistent format.
func NormalizeIDEClient(ideClient string) string {
	if ideClient == _visualStudioCodeIDEClient {
		return "vscode"
	} else if ideClient == _cursorIDEClient {
		return "cursor"
	}
	return ideClient
}
