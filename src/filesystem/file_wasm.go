//go:build js
// +build js

package filesystem

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

var ErrorNotExist = errors.New("Path doesn't exist")
var ErrorAlreadyExist = errors.New("Path Already Exists")
var ErrorIsFile = errors.New("Path is File, Can't use as Directory")
var ErrorIsDir = errors.New("Path is Directory, Can't use as File")
var ErrorSeekToFar = errors.New("Cannot Seek that far")
var ErrorIsClosed = errors.New("File Is Closed")

// const ErrorNotExist = 1
// const ErrorAlreadyExist = 2
// const ErrorIsFile = 3
// const ErrorIsDir = 4

func NewFileSystem() AbstractFileSystem {
	fmt.Println("About to build BrowserFS")

	jsVar := js.Global().Get("IKEMEN_GO_BROWSER_FS")
	fs := &BrowserFS{jsVar: jsVar}

	fmt.Println("built BrowserFS")
	return fs
}

type BrowserFS struct {
	jsVar js.Value
}

func (fs *BrowserFS) Exists(name string) bool {
	return fs.jsVar.Get("existsSync").Invoke(name).Bool()
}

func (fs *BrowserFS) Mkdir(name string, perm IFileMode) error {
	if fs.Exists(name) {
		return ErrorAlreadyExist
	}

	fs.jsVar.Get("mkdirSync").Invoke(name, perm)

	return nil
}

func (fs *BrowserFS) Create(name string) (IFile, error) {
	fs.jsVar.Get("writeFileSync").Invoke(name, "")
	return makeNewFile(fs.jsVar, name), nil
}

func (fs *BrowserFS) Open(name string) (IFile, error) {
	stat, statErr := fs.Stat(name)
	if statErr != nil {
		return nil, statErr
	}
	if stat.IsDir() {
		return wrapFolder(fs.jsVar, name), nil
	}
	buffer := fs.jsVar.Get("readFileSync").Invoke(name, nil)
	return wrapBufferAsFile(fs.jsVar, name, buffer), nil
}

func (fs *BrowserFS) ReadFile(path string) ([]byte, error) {
	buffer := fs.jsVar.Get("readFileSync").Invoke(path, nil)
	return jsValueToByteArray(buffer), nil
}

func (fs *BrowserFS) WriteFile(name string, data []byte, perm IFileMode) error {
	fs.jsVar.Get("writeFileSync").Invoke(name, data)
	return nil
}

func (fs *BrowserFS) Stat(path string) (IFileInfo, error) {
	if !fs.Exists(path) {
		return nil, ErrorNotExist
	}
	fmt.Println(("about to statSync"))
	statVar := fs.jsVar.Get("statSync").Invoke(path)
	fmt.Println(("statSync Successful"))
	stat := new(BrowserFsFileInfo)
	stat.path = path
	stat.js = statVar

	js.Global().Get("console").Call("log", stat.js)
	fmt.Println(("testing stat fns now!"))

	stat.IsDir()

	fmt.Println(("statSync isDir Successful"))
	stat.ModTime()
	fmt.Println(("statSync ModTime Successful"))
	stat.Mode()
	fmt.Println(("statSync Mode Successful"))
	stat.Name()
	fmt.Println(("statSync Name Successful"))
	stat.Size()
	fmt.Println(("statSync Size Successful"))

	return stat, nil
}

func (fs *BrowserFS) IsNotExist(err error) bool {
	return errors.Is(err, ErrorNotExist)
}

const Separator string = "/"

func (fs *BrowserFS) ReadDir(path string) ([]string, error) {
	fmt.Println("reading dir " + path)
	statValue, statErr := fs.Stat(path)
	if statErr != nil {
		fmt.Println("failed stat")
		return nil, statErr
	}
	if !statValue.IsDir() {
		fmt.Println("file is not dir")
		return nil, ErrorIsFile
	}
	fmt.Println("file is dir")
	dirArray := fs.jsVar.Get("readdirSync").Invoke(path)
	fmt.Println("read dir success")

	dirLen := dirArray.Get("length").Int()
	fmt.Println("got length: " + strconv.Itoa(dirLen))

	children := make([]string, dirLen)
	for i := 0; i < dirLen; i++ {
		key := strconv.Itoa(i)
		children[i] = dirArray.Get(key).String()
	}
	return children, nil
}

func (fs *BrowserFS) Walk(root string, fn WalkFn) error {
	return fs.WalkRecursive(root, fn)
}

func (fs *BrowserFS) WalkLoop(root string, fn WalkFn) error {
	var decendents [][]string
	var activeChildren = []string{root}
	for true {
		if len(activeChildren) == 0 {
			if len(decendents) == 0 {
				return nil
			}
			activeChildren, decendents = decendents[len(activeChildren)-1], decendents[:len(activeChildren)-1]
			activeChildren = activeChildren[1:]
		}
		child := activeChildren[0]

		fullPath := ""
		for i := 0; i < len(decendents); i++ {
			fullPath = fullPath + decendents[i][0] + Separator
		}
		fullPath = fullPath + child

		stat, statErr := fs.Stat(fullPath)
		if statErr != nil {
			return statErr
		}
		error := fn(fullPath, stat, nil)
		if error != nil {
			return error
		}

		if stat.IsDir() {
			decendents = append(decendents, activeChildren)
			newChildren, readDirError := fs.ReadDir(fullPath)
			if readDirError != nil {
				return readDirError
			}
			activeChildren = newChildren
		} else {
			activeChildren = activeChildren[1:]
		}
	}

	return nil
}

func (fs *BrowserFS) WalkRecursive(root string, fn WalkFn) error {
	dirArray := fs.jsVar.Get("readdirSync").Invoke(root)
	len := dirArray.Get("length").Int()
	for i := 0; i < len; i++ {
		key := strconv.Itoa(i)
		fullPath := root + Separator + dirArray.Get(key).String()
		stat, statError := fs.Stat(fullPath)
		if statError != nil {
			return statError
		}
		error := fn(fullPath, stat, nil)
		if error != nil {
			return error
		}
		if !stat.IsDir() {
			continue
		}
		error = fs.WalkRecursive(fullPath, fn)
		if error != nil {
			return error
		}
	}
	return nil
}

/*
==================================================

Browser File

==================================================
*/

type BrowserFsFile struct {
	jsVar    js.Value
	isDir    bool
	path     string
	content  []byte
	offset   uint64
	isClosed bool
}

func makeNewFile(jsVar js.Value, path string) *BrowserFsFile {
	file := new(BrowserFsFile)
	file.jsVar = jsVar
	file.isDir = false
	file.content = make([]byte, 0, 0)
	file.path = path
	file.offset = 0
	file.isClosed = false
	return file
}

func wrapFolder(jsVar js.Value, path string) *BrowserFsFile {
	dir := new(BrowserFsFile)
	// https://www.reddit.com/r/golang/comments/bxcxxe/make_byte_array_as_empty/
	dir.jsVar = jsVar
	dir.isDir = true
	dir.content = make([]byte, 0, 0)
	dir.path = path
	dir.offset = 0
	dir.isClosed = false
	return dir
}

func wrapBufferAsFile(jsVar js.Value, path string, buffer js.Value) *BrowserFsFile {
	file := new(BrowserFsFile)
	file.jsVar = jsVar
	file.isDir = false
	file.content = jsValueToByteArray(buffer)
	file.path = path
	file.offset = 0
	file.isClosed = false
	return file
}

func jsValueToByteArray(buffer js.Value) []byte {
	destination := make([]byte, 0, 0)
	js.CopyBytesToGo(destination, buffer)
	return destination
	// https://github.com/gopherjs/gopherjs/issues/165#issuecomment-71513058
	// return js.Global.Get("Uint8Array").New(buffer).Interface().([]byte)
}

func (file *BrowserFsFile) Seek(offset int64, whence int) (ret int64, err error) {
	if file.isClosed {
		return 0, ErrorIsClosed
	}
	if file.isDir {
		return 0, ErrorIsDir
	}

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(file.offset) + offset
	case io.SeekEnd:
		newOffset = int64(len(file.content)) + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if newOffset < 0 {
		return 0, errors.New("negative result offset")
	}

	file.offset = uint64(newOffset)
	return newOffset, nil
}

func (f *BrowserFsFile) Write(p []byte) (n int, err error) {
	if f.isClosed {
		return 0, ErrorIsClosed
	}
	if f.isDir {
		return 0, ErrorIsDir
	}

	// Check if the offset is beyond the current content size
	if f.offset > uint64(len(f.content)) {
		// If so, extend the content slice with zeroes until the offset
		padding := make([]byte, f.offset-uint64(len(f.content)))
		f.content = append(f.content, padding...)
	}

	// Append the data to the content slice
	f.content = append(f.content, p...)
	n = len(p)

	// Update the offset
	f.offset += uint64(n)

	return n, nil
}

func (file *BrowserFsFile) Read(b []byte) (n int, err error) {
	if file.isClosed {
		return 0, ErrorIsClosed
	}
	if file.isDir {
		return 0, ErrorIsDir
	}

	if file.offset >= uint64(len(file.content)) {
		return 0, io.EOF // End of file
	}

	n = copy(b, file.content[file.offset:])
	file.offset += uint64(n)
	return n, nil
}

func (file *BrowserFsFile) Close() error {
	if file.isClosed {
		return ErrorIsClosed
	}
	file.isClosed = true
	return nil
}

func (file *BrowserFsFile) Readdirnames(n int) (names []string, err error) {
	if !file.isDir {
		return nil, ErrorIsFile
	}
	dirArray := file.jsVar.Get("readdirSync").Invoke(file.path)
	len := dirArray.Get("length").Int()

	if n <= 0 || n > len {
		n = len
	}
	children := make([]string, n)
	for i := 0; i < n; i++ {
		key := strconv.Itoa(i)
		children[i] = dirArray.Get(key).String()
	}
	return children, nil
}

/*

==================================================

Browser File Info

==================================================

*/

type BrowserFsFileInfo struct {
	path string
	js   js.Value
}

func (info *BrowserFsFileInfo) IsDir() bool {
	return info.js.Call("isDirectory").Bool()
}
func (info *BrowserFsFileInfo) ModTime() time.Time {
	modTimestamp := info.js.Get("mtime").Call("getTime").Int()
	return time.Unix(int64(modTimestamp), 0)
}

func (info *BrowserFsFileInfo) Mode() IFileMode {
	intValue := info.js.Get("mode").Int()
	return IFileMode(intValue)
}

func (info *BrowserFsFileInfo) Name() string {
	return info.path
}

func (info *BrowserFsFileInfo) Size() int64 {
	return int64(info.js.Get("size").Int())
}

/*

==================================================

This is all related to Glob

==================================================

*/

var ErrBadPattern = errors.New("syntax error in pattern")

func (fs *BrowserFS) Glob(pattern string) (matches []string, err error) {
	return fs.globWithLimit(pattern, 0)
}

func (fs *BrowserFS) globWithLimit(pattern string, depth int) (matches []string, err error) {
	// This limit is used prevent stack exhaustion issues. See CVE-2022-30632.
	const pathSeparatorsLimit = 10000
	if depth == pathSeparatorsLimit {
		return nil, ErrBadPattern
	}

	// Check pattern is well-formed.
	if _, err := filepath.Match(pattern, ""); err != nil {
		return nil, err
	}
	if !hasMeta(pattern) {
		if _, err = fs.Stat(pattern); err != nil {
			return nil, nil
		}
		return []string{pattern}, nil
	}

	dir, file := filepath.Split(pattern)
	volumeLen := 0
	dir = cleanGlobPath(dir)

	if !hasMeta(dir[volumeLen:]) {
		return fs.glob(dir, file, nil)
	}

	// Prevent infinite recursion. See issue 15879.
	if dir == pattern {
		return nil, ErrBadPattern
	}

	var m []string
	m, err = fs.globWithLimit(dir, depth+1)
	if err != nil {
		return
	}
	for _, d := range m {
		matches, err = fs.glob(d, file, matches)
		if err != nil {
			return
		}
	}
	return
}

// glob searches for files matching pattern in the directory dir
// and appends them to matches. If the directory cannot be
// opened, it returns the existing matches. New matches are
// added in lexicographical order.
func (fs *BrowserFS) glob(dir, pattern string, matches []string) (m []string, e error) {
	m = matches
	fi, err := fs.Stat(dir)
	if err != nil {
		return // ignore I/O error
	}
	if !fi.IsDir() {
		return // ignore I/O error
	}
	d, err := fs.Open(dir)
	if err != nil {
		return // ignore I/O error
	}
	defer d.Close()

	names, _ := d.Readdirnames(-1)
	sort.Strings(names)

	for _, n := range names {
		matched, err := filepath.Match(pattern, n)
		if err != nil {
			return m, err
		}
		if matched {
			m = append(m, filepath.Join(dir, n))
		}
	}
	return
}

func cleanGlobPath(path string) string {
	switch path {
	case "":
		return "."
	case string(Separator):
		// do nothing to the path
		return path
	default:
		return path[0 : len(path)-1] // chop off trailing separator
	}
}

func hasMeta(path string) bool {
	magicChars := `*?[`
	return strings.ContainsAny(path, magicChars)
}
