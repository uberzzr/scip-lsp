package fs

import (
	"io/fs"
	"os"
	"os/exec"

	"go.uber.org/fx"
)

// Module is the Fx module for this package.
var Module = fx.Provide(New)

// UlspFS will wrap the filesystem operations used by ulsp.
type UlspFS interface {
	UserCacheDir() (string, error)
	MkdirAll(path string) error
	WorkspaceRoot(path string) ([]byte, error)
	DirExists(path string) (bool, error)
	FileExists(path string) (bool, error)
	Open(name string) (*os.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data string) error
	Create(name string) (*os.File, error)
	TempFile(dir, pattern string) (*os.File, error)
	Remove(name string) error
}

type fsImpl struct{}

// New creates a new UlspFS.
func New() UlspFS {
	return fsImpl{}
}

// UserCacheDir returns the user's cache directory.
func (fsImpl) UserCacheDir() (string, error) { return os.UserCacheDir() }

// MkdirAll creates a directory and all its parents.
func (fsImpl) MkdirAll(path string) error { return os.MkdirAll(path, os.ModePerm) }

// WorkspaceRoot returns the workspace root for the given path.
func (fsImpl) WorkspaceRoot(path string) ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	return cmd.Output()
}

// ReadDir reads all the items in a directory (non-recursive)
func (fsImpl) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

// Open opens a file for reading
func (fsImpl) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (fsImpl) DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func (fsImpl) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fsImpl) WriteFile(name string, data string) error {
	return os.WriteFile(name, []byte(data), 0644)
}

func (fsImpl) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (fsImpl) TempFile(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (fsImpl) Remove(name string) error {
	return os.Remove(name)
}

func (fsImpl) FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}
