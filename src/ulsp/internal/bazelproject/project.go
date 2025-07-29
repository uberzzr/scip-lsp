package bazelproject

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// ProjectView stores information in .bazelproject file.
type ProjectView struct {
	Path                          string
	Targets, Directories, Imports []string
	DeriveTargets                 bool
}

const _importPrefix = "import "

// ParseProjectView parses a .bazelproject file and its imports
// with depth-first order. It assumes that the file is
// a YAML except the imports. More information:
// https://ij.bazel.build/docs/project-views.html
// This function parses the file with best effort, so it may
// return partially valid content with errors indicating the issues
// it runs into.
func ParseProjectView(projectFile io.Reader) (project ProjectView, err error) {
	var rawYAML bytes.Buffer
	scanner := bufio.NewScanner(projectFile)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, _importPrefix) {
			rawYAML.WriteString(line)
			rawYAML.WriteString("\n")
			continue
		}
		importPath := strings.TrimPrefix(line, _importPrefix)
		if commentStart := strings.Index(importPath, "#"); commentStart != -1 {
			importPath = importPath[:commentStart]
		}
		importPath = strings.TrimSpace(importPath)
		if importPath == "" {
			err = multierr.Append(err, fmt.Errorf("invalid import %q", line))
			continue
		}
		project.Imports = append(project.Imports, importPath)
	}
	var content struct {
		Targets       string
		Directories   string
		DeriveTargets bool `yaml:"derive_targets_from_directories"`
	}
	if e := yaml.NewDecoder(&rawYAML).Decode(&content); e != nil {
		return project, multierr.Append(err, e)
	}
	project.Targets = strings.Fields(content.Targets)
	project.Directories = strings.Fields(content.Directories)
	project.DeriveTargets = content.DeriveTargets
	return project, err
}
