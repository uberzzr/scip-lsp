package bazelproject

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProjectView(t *testing.T) {
	tests := []struct {
		desc, input string
		expected    ProjectView
	}{
		{
			desc: "parsing directories",
			input: `directories:
  # Add the directories you want added as source here
  # By default, we've added your entire workspace ('.')
  src/code.uber.internal/devexp/ci-jobs

  src/code.uber.internal/contrib/git-bzl
  src/code.uber.internal/devexp/bazel/ide

# Automatically includes all relevant targets under the 'directories' above
derive_targets_from_directories: true

targets:
  # If source code isn't resolving, add additional targets that compile it here

additional_languages:
  # Uncomment any additional languages you want supported
  # dart
  # javascript
  # python
  # typescript
`,
			expected: ProjectView{Directories: []string{
				"src/code.uber.internal/devexp/ci-jobs",
				"src/code.uber.internal/contrib/git-bzl",
				"src/code.uber.internal/devexp/bazel/ide",
			}},
		},
		{
			desc: "parsing imports",
			input: `directories:
  # Add the directories you want added as source here
  # By default, we've added your entire workspace ('.')

import testdata/normal.bazelproject

# Automatically includes all relevant targets under the 'directories' above
derive_targets_from_directories: true

targets:
  # If source code isn't resolving, add additional targets that compile it here
`,
			expected: ProjectView{Imports: []string{"testdata/normal.bazelproject"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			reader := bytes.NewBufferString(tt.input)
			project, err := ParseProjectView(reader)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected.Directories, project.Directories)
			assert.ElementsMatch(t, tt.expected.Imports, project.Imports)
			assert.ElementsMatch(t, tt.expected.Targets, project.Targets)
		})
	}
}
