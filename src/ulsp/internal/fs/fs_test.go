package fs

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCacheDir(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp")
	fs := New()
	dir, err := fs.UserCacheDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
}

func TestMkdirAll(t *testing.T) {
	dir := t.TempDir()
	fs := New()
	err := fs.MkdirAll(path.Join(dir, "foo/bar"))
	assert.NoError(t, err)
}

func TestWorkspaceRoot(t *testing.T) {
	workspace := prepareWorkspaceDirectory(t)
	fs := New()
	_, err := fs.WorkspaceRoot(workspace)
	assert.NoError(t, err)
}

func TestDirExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		dir := t.TempDir()
		fs := New()
		result, err := fs.DirExists(dir)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("does not exist", func(t *testing.T) {
		dir := t.TempDir()
		fs := New()
		result, err := fs.DirExists(dir + "foo")
		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestFileExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		t.Skip()
		filePath := ""
		// filePath := iotest.TempFile(t, "foo")
		fs := New()
		result, err := fs.FileExists(filePath)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("does not exist", func(t *testing.T) {
		t.Skip()
		filePath := ""
		// filePath := iotest.TempFile(t, "foo")
		os.Remove(filePath)
		fs := New()
		result, err := fs.FileExists(filePath)
		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	file := path.Join(dir, "a")
	os.WriteFile(path.Join(dir, "a"), []byte("contents"), 0666)
	fs := New()
	result, err := fs.Open(file)
	defer result.Close()
	assert.NoError(t, err)
	assert.Equal(t, "a", path.Base(result.Name()))
}

func TestReadDir(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(path.Join(dir, "a"), []byte("a"), 0666)
		os.WriteFile(path.Join(dir, "b"), []byte("b"), 0666)
		fs := New()
		result, err := fs.ReadDir(dir)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("empty", func(t *testing.T) {
		dir := t.TempDir()
		fs := New()
		result, err := fs.ReadDir(dir)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("does not exist", func(t *testing.T) {
		dir := t.TempDir()
		fs := New()
		_, err := fs.ReadDir(dir + "foo")
		assert.Error(t, err)
	})
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	file := path.Join(dir, "a")
	os.WriteFile(path.Join(dir, "a"), []byte("contents"), 0666)
	fs := New()
	result, err := fs.ReadFile(file)
	assert.NoError(t, err)
	assert.Equal(t, "contents", string(result))

}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	file := path.Join(dir, "a")
	fs := New()
	err := fs.WriteFile(file, "data")
	assert.NoError(t, err)
	result, _ := os.ReadFile(file)
	assert.Equal(t, "data", string(result))
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	file := path.Join(dir, "a")
	fs := New()
	result, err := fs.Create(file)
	assert.NoError(t, err)
	assert.Equal(t, "a", path.Base(result.Name()))
}

func TestTempFile(t *testing.T) {
	dir := t.TempDir()
	fs := New()
	result, err := fs.TempFile(dir, "foo")
	defer os.Remove(result.Name())
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(result.Name(), path.Join(dir, "foo")))
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	file := path.Join(dir, "a")
	os.WriteFile(path.Join(dir, "a"), []byte("contents"), 0666)
	fs := New()
	err := fs.Remove(file)
	assert.NoError(t, err)
}

func prepareWorkspaceDirectory(t *testing.T) string {
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	return workspace
}

func initGitRepo(t *testing.T, tmpDir string) {
	gitCommandInDir(t, tmpDir, "init")
	gitCommandInDir(t, tmpDir, "config", "user.email", "test@uber.com")
	gitCommandInDir(t, tmpDir, "config", "user.name", "Test User")
}

func gitCommandInDir(t *testing.T, repoDir string, args ...string) string {
	exec := exec.Command("git", args...)
	exec.Dir = repoDir
	out, err := exec.CombinedOutput()
	require.NoError(t, err, "failed git command %s - %v", out, err)
	return string(out)
}
