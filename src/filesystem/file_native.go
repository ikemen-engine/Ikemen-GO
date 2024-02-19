//go:build !js
// +build !js

package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func NewFileSystem() AbstractFileSystem {
	return &NativeFS{}
}

type NativeFS struct{}

func (osFileSystem *NativeFS) Mkdir(name string, perm IFileMode) error {
	return os.Mkdir(name, fs.FileMode(perm))
}
func (osFileSystem *NativeFS) Create(name string) (IFile, error) {
	return os.Create(name)
}
func (osFileSystem *NativeFS) Open(name string) (IFile, error) {
	return os.Open(name)
}
func (osFileSystem *NativeFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
func (osFileSystem *NativeFS) WriteFile(name string, data []byte, perm IFileMode) error {
	return os.WriteFile(name, data, fs.FileMode(perm))
}
func (osFileSystem *NativeFS) Stat(name string) (IFileInfo, error) {
	stat, error := os.Stat(name)
	if error != nil {
		return nil, error
	}
	return NewFileInfoWrapper(stat), nil
}
func (osFileSystem *NativeFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (osFileSystem *NativeFS) ReadDir(path string) ([]string, error) {
	entries, error := os.ReadDir(path)
	if error != nil {
		return nil, error
	}
	entryLen := len(entries)
	children := make([]string, entryLen)
	for i := 0; i < entryLen; i++ {
		children[i] = entries[i].Name()
	}
	return children, nil
}

func (osFileSystem *NativeFS) Walk(root string, wallkFn WalkFn) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		return wallkFn(path, NewFileInfoWrapper(info), err)
	})
}
func (osFileSystem *NativeFS) Glob(pattern string) (matches []string, err error) {
	return filepath.Glob(pattern)
}

/*
=======================================================
FileInfoWrapper
- to handle the IFileMode incompatabilities between abstraction and os.FileInfo
=======================================================
*/
type FileInfoWrapper struct {
	FileInfo os.FileInfo
}

func NewFileInfoWrapper(fi os.FileInfo) *FileInfoWrapper {
	return &FileInfoWrapper{FileInfo: fi}
}

func (fiw *FileInfoWrapper) IsDir() bool {
	return fiw.FileInfo.IsDir()
}

func (fiw *FileInfoWrapper) ModTime() time.Time {
	return fiw.FileInfo.ModTime()
}

func (fiw *FileInfoWrapper) Mode() IFileMode {
	return IFileMode(fiw.FileInfo.Mode())
}

func (fiw *FileInfoWrapper) Name() string {
	return fiw.FileInfo.Name()
}

func (fiw *FileInfoWrapper) Size() int64 {
	return fiw.FileInfo.Size()
}
