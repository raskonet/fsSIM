package memfs

import (
	"errors"
	"fmt"
	"io"
	"bufio"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileSystem struct {
	root      *Directory
	mu        sync.RWMutex
	openFiles int
}

func NewFileSystem() *FileSystem {
	return &FileSystem{
		root: &Directory{
			name:      "/",
			children:  make(map[string]FSNode),
			createdAt: time.Now(),
			updatedAt: time.Now(),
		},
	}
}

func (fs *FileSystem) splitPath(path string) []string {
	path = filepath.Clean(path)
	if path == "/" {
		return []string{}
	}
	path = strings.TrimPrefix(path, "/")
	return strings.Split(path, "/")
}

func (fs *FileSystem) getNode(path string) (FSNode, error) {
	path = filepath.Clean(path)
	if path == "/" {
		return fs.root, nil
	}
	components := fs.splitPath(path)
	currentDir := fs.root
	for i, component := range components {
		child, exists := currentDir.children[component]
		if !exists {
			return nil, NewFSError(OpStat, path, ErrNotExist)
		}
		if i == len(components)-1 {
			return child, nil
		}
		if !child.IsDir() {
			return nil, NewFSError(OpStat, path, ErrNotDir)
		}
		currentDir = child.(*Directory)
	}
	return currentDir, nil
}

func (fs *FileSystem) getParentDir(path string) (*Directory, string, error) {
	path = filepath.Clean(path)
	if path == "/" {
		return nil, "", NewFSError(OpStat, path, ErrIsRoot)
	}
	parentPath := filepath.Dir(path)
	baseName := filepath.Base(path)
	node, err := fs.getNode(parentPath)
	if err != nil {
		return nil, "", err
	}
	dir, ok := node.(*Directory)
	if !ok {
		return nil, "", NewFSError(OpStat, parentPath, ErrNotDir)
	}
	return dir, baseName, nil
}

func (fs *FileSystem) getOpenFileCount() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.openFiles
}

func (fs *FileSystem) decrementOpenFiles() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.openFiles--
}

func (fs *FileSystem) Stat(path string) (FSInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	node, err := fs.getNode(path)
	if err != nil {
		return FSInfo{}, err
	}
	var size int64
	if !node.IsDir() {
		file := node.(*File)
		size = int64(len(file.content))
	}
	return FSInfo{
		Name:      node.Name(),
		IsDir:     node.IsDir(),
		Size:      size,
		CreatedAt: node.CreatedAt(),
		UpdatedAt: node.UpdatedAt(),
	}, nil
}

func (fs *FileSystem) DirSize(path string) (int64, error) {
	node, err := fs.getNode(path)
	if err != nil {
		return 0, err
	}
	if !node.IsDir() {
		return 0, fmt.Errorf("not a directory")
	}
	return fs.dirSizeRecursive(node.(*Directory)), nil
}

func (fs *FileSystem) dirSizeRecursive(dir *Directory) int64 {
	var size int64
	for _, child := range dir.children {
		if child.IsDir() {
			size += fs.dirSizeRecursive(child.(*Directory))
		} else {
			size += int64(len(child.(*File).content))
		}
	}
	return size
}

func (fs *FileSystem) Mkdir(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if path == "/" {
		return NewFSError(OpMkdir, path, ErrRootExists)
	}
	parentDir, baseName, err := fs.getParentDir(path)
	if err != nil {
		var fsErr *FSError
		if errors.As(err, &fsErr) && errors.Is(fsErr.Err, ErrNotExist) {
			err = fs.Mkdir(filepath.Dir(path))
			if err != nil {
				return err
			}
			parentDir, baseName, err = fs.getParentDir(path)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if _, exists := parentDir.children[baseName]; exists {
		return NewFSError(OpMkdir, path, ErrExist)
	}
	parentDir.children[baseName] = &Directory{
		name:      baseName,
		children:  make(map[string]FSNode),
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
	parentDir.updatedAt = time.Now()
	return nil
}

func (fs *FileSystem) Create(path string) (*File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if path == "/" {
		return nil, NewFSError(OpCreate, path, ErrIsDir)
	}
	parentDir, baseName, err := fs.getParentDir(path)
	if err != nil {
		return nil, err
	}
	if node, exists := parentDir.children[baseName]; exists {
		if node.IsDir() {
			return nil, NewFSError(OpCreate, path, ErrIsDir)
		}
		file := node.(*File)
		file.content = []byte{}
		file.position = 0
		file.updatedAt = time.Now()
		return file, nil
	}
	now := time.Now()
	file := &File{
		name:      baseName,
		content:   []byte{},
		position:  0,
		createdAt: now,
		updatedAt: now,
		fs:        fs,
	}
	parentDir.children[baseName] = file
	parentDir.updatedAt = now
	fs.openFiles++
	return file, nil
}

func (fs *FileSystem) Open(path string) (*File, error) {
	fs.mu.RLock()
	node, err := fs.getNode(path)
	if err != nil {
		fs.mu.RUnlock()
		return nil, err
	}
	if node.IsDir() {
		fs.mu.RUnlock()
		return nil, NewFSError(OpOpen, path, ErrIsDir)
	}
	file := node.(*File)
	fs.mu.RUnlock()
	fs.mu.Lock()
	fs.openFiles++
	fs.mu.Unlock()
	return file, nil
}

func (fs *FileSystem) ReadDir(path string) ([]FSNode, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	node, err := fs.getNode(path)
	if err != nil {
		return nil, err
	}
	if !node.IsDir() {
		return nil, NewFSError(OpReadDir, path, ErrNotDir)
	}
	dir := node.(*Directory)
	result := make([]FSNode, 0, len(dir.children))
	for _, child := range dir.children {
		result = append(result, child)
	}
	return result, nil
}

func (fs *FileSystem) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	node, err := fs.getNode(path)
	fs.mu.RUnlock()
	if err != nil {
		return nil, err
	}
	if node.IsDir() {
		return nil, NewFSError(OpReadFile, path, ErrIsDir)
	}
	file := node.(*File)
	fs.mu.RLock()
	content := make([]byte, len(file.content))
	copy(content, file.content)
	fs.mu.RUnlock()
	return content, nil
}

func (fs *FileSystem) WriteFile(path string, data []byte) error {
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func (fs *FileSystem) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if path == "/" {
		return NewFSError(OpRemove, path, ErrIsRoot)
	}
	parentDir, baseName, err := fs.getParentDir(path)
	if err != nil {
		return err
	}
	node, exists := parentDir.children[baseName]
	if !exists {
		return NewFSError(OpRemove, path, ErrNotExist)
	}
	if node.IsDir() {
		dir := node.(*Directory)
		if len(dir.children) > 0 {
			return NewFSError(OpRemove, path, ErrDirNotEmpty)
		}
	}
	delete(parentDir.children, baseName)
	parentDir.updatedAt = time.Now()
	return nil
}

func (fs *FileSystem) RemoveAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if path == "/" {
		fs.root.children = make(map[string]FSNode)
		fs.root.updatedAt = time.Now()
		return nil
	}
	parentDir, baseName, err := fs.getParentDir(path)
	if err != nil {
		return err
	}
	_, exists := parentDir.children[baseName]
	if !exists {
		return NewFSError(OpRemoveAll, path, ErrNotExist)
	}
	delete(parentDir.children, baseName)
	parentDir.updatedAt = time.Now()
	return nil
}

func (fs *FileSystem) Rename(oldPath, newPath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if oldPath == "/" {
		return NewFSError(OpRename, oldPath, ErrIsRoot)
	}
	if newPath == "/" {
		return NewFSError(OpRename, newPath, ErrIsRoot)
	}
	oldParentDir, oldBaseName, err := fs.getParentDir(oldPath)
	if err != nil {
		return err
	}
	oldNode, exists := oldParentDir.children[oldBaseName]
	if !exists {
		return NewFSError(OpRename, oldPath, ErrNotExist)
	}
	newParentDir, newBaseName, err := fs.getParentDir(newPath)
	if err != nil {
		return err
	}
	if _, exists := newParentDir.children[newBaseName]; exists {
		return NewFSError(OpRename, newPath, ErrExist)
	}
	delete(oldParentDir.children, oldBaseName)
	newParentDir.children[newBaseName] = oldNode
	if file, ok := oldNode.(*File); ok {
		file.name = newBaseName
	} else if dir, ok := oldNode.(*Directory); ok {
		dir.name = newBaseName
	}
	now := time.Now()
	oldParentDir.updatedAt = now
	newParentDir.updatedAt = now
	return nil
}

