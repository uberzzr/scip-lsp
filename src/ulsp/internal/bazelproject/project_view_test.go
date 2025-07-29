package bazelproject

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs/fsmock"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() {
		New(Params{})
	})
}

func TestGetSource(t *testing.T) {
	samplePaths := []string{
		"sample1.bazelproject",
		"sample2.bazelproject",
		"sample3.bazelproject",
	}

	t.Run("all files exist", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)

		for _, current := range samplePaths {
			absPath := path.Join(workspaceRoot, current)
			os.WriteFile(absPath, []byte("test"), 0o644)
		}

		p := projectViewPatternsProvider{
			workspaceRoot: workspaceRoot,
			paths:         samplePaths,
			fs:            fs.New(),
			output:        outputWriter,
		}

		p.setSource(context.Background())
		assert.Equal(t, path.Join(workspaceRoot, samplePaths[0]), p.selectedSource)
	})

	t.Run("some files exist", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)

		for _, current := range []string{samplePaths[1], samplePaths[2]} {
			absPath := path.Join(workspaceRoot, current)
			os.WriteFile(absPath, []byte("test"), 0o644)
		}

		p := projectViewPatternsProvider{
			workspaceRoot: workspaceRoot,
			paths:         samplePaths,
			fs:            fs.New(),
			output:        outputWriter,
		}

		p.setSource(context.Background())
		assert.Equal(t, path.Join(workspaceRoot, samplePaths[1]), p.selectedSource)
	})

	t.Run("no files exist", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)

		p := projectViewPatternsProvider{
			workspaceRoot: workspaceRoot,
			paths:         samplePaths,
			fs:            fs.New(),
			output:        outputWriter,
		}

		p.setSource(context.Background())
		assert.Len(t, p.selectedSource, 0)
	})

	t.Run("error", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)
		ctrl := gomock.NewController(t)

		fs := fsmock.NewMockUlspFS(ctrl)
		p := projectViewPatternsProvider{
			workspaceRoot: workspaceRoot,
			paths:         append(samplePaths, workspaceRoot),
			fs:            fs,
			output:        outputWriter,
		}

		fs.EXPECT().FileExists(gomock.Any()).Return(false, errors.New("error"))
		fs.EXPECT().FileExists(gomock.Any()).Return(true, nil)

		p.setSource(context.Background())
		assert.Equal(t, path.Join(workspaceRoot, samplePaths[1]), p.selectedSource)
	})

	t.Run("already set", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)

		for _, current := range samplePaths {
			absPath := path.Join(workspaceRoot, current)
			os.WriteFile(absPath, []byte("test"), 0o644)
		}

		p := projectViewPatternsProvider{
			workspaceRoot:  workspaceRoot,
			paths:          samplePaths,
			fs:             fs.New(),
			output:         outputWriter,
			selectedSource: path.Join(workspaceRoot, samplePaths[1]),
		}

		p.setSource(context.Background())
		assert.Equal(t, path.Join(workspaceRoot, samplePaths[1]), p.selectedSource)
	})
}

func TestGetTargets(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     []string
		wantErr  bool
	}{
		{
			name: "valid file",
			contents: `
targets:
  //my/sample/path:all
  //another/sample/path/...

other_tag:
  sample
`,
			want: []string{"//my/sample/path:all", "//another/sample/path/..."},
		},
		{
			name: "invalid file",
			contents: `
targets:
  //my/sample/path:all
  //another/sample/path/...

other
`,
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "no file",
			wantErr: true,
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaceRoot := t.TempDir()
			absPath := filepath.Join(workspaceRoot, "sample.projectview")

			if len(tt.contents) > 0 {
				os.WriteFile(absPath, []byte(tt.contents), 0o644)
			}

			p := projectViewPatternsProvider{
				workspaceRoot: workspaceRoot,
				fs:            fs.New(),
			}

			targets, err := p.getPatterns(absPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.ElementsMatch(t, targets, tt.want)
		})
	}
}

func TestGetPresetPatterns(t *testing.T) {
	samplePaths := []string{
		"sample1.bazelproject",
		"sample2.bazelproject",
		"sample3.bazelproject",
	}

	t.Run("success", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)

		ctrl := gomock.NewController(t)
		fs := fsmock.NewMockUlspFS(ctrl)

		p := projectViewPatternsProvider{
			workspaceRoot: workspaceRoot,
			paths:         samplePaths,
			fs:            fs,
			output:        outputWriter,
		}

		fs.EXPECT().FileExists(path.Join(workspaceRoot, samplePaths[0])).Return(false, nil)
		fs.EXPECT().FileExists(path.Join(workspaceRoot, samplePaths[1])).DoAndReturn(func(file string) (bool, error) {
			fs.EXPECT().Open(file).DoAndReturn(func(name string) (*os.File, error) {
				err := os.WriteFile(name, []byte("targets:\n  //my/sample/path:all\n  //another/sample/path/..."), 0o644)
				assert.NoError(t, err)
				return os.Open(name)
			})
			return true, nil
		})

		sampleResult := []string{"//my/sample/path:target_a", "//my/sample/path:target_b", "//another/sample/path:target_c"}
		result, err := p.GetPresetPatterns(context.Background(), func(ctx context.Context, patterns []string) ([]string, error) {
			assert.ElementsMatch(t, patterns, []string{"//my/sample/path:all", "//another/sample/path/..."})
			return sampleResult, nil
		})
		assert.NoError(t, err)
		assert.ElementsMatch(t, result, sampleResult)
	})

	t.Run("error", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		outputWriter := bytes.NewBuffer(nil)

		ctrl := gomock.NewController(t)
		fs := fsmock.NewMockUlspFS(ctrl)

		p := projectViewPatternsProvider{
			workspaceRoot: workspaceRoot,
			paths:         samplePaths,
			fs:            fs,
			output:        outputWriter,
		}

		fs.EXPECT().FileExists(path.Join(workspaceRoot, samplePaths[0])).Return(false, nil)
		fs.EXPECT().FileExists(path.Join(workspaceRoot, samplePaths[1])).DoAndReturn(func(file string) (bool, error) {
			fs.EXPECT().Open(file).Return(nil, errors.New("error"))
			return true, nil
		})

		_, err := p.GetPresetPatterns(context.Background(), func(ctx context.Context, patterns []string) ([]string, error) {
			return nil, nil
		})
		assert.Error(t, err)
	})
}
