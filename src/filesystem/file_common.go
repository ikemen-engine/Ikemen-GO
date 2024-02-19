package filesystem

import (
	"time"
)

type IFile interface {
	Readdirnames(n int) (names []string, err error)
	Seek(offset int64, whence int) (ret int64, err error)
	Write(p []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	Close() error
}

type IFileInfo interface {
	IsDir() bool
	ModTime() time.Time
	Mode() IFileMode
	Name() string
	Size() int64
}

type IFileMode uint32

type WalkFn func(path string, info IFileInfo, err error) error

type AbstractFileSystem interface {
	Mkdir(name string, perm IFileMode) error
	Create(name string) (IFile, error)
	Open(name string) (IFile, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm IFileMode) error
	Stat(name string) (IFileInfo, error)
	IsNotExist(err error) bool

	ReadDir(path string) ([]string, error)
	Walk(root string, fn WalkFn) error
	Glob(pattern string) (matches []string, err error)
}
