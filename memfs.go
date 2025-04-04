package memfs

import(
	"fmt"
	"os"
	"io"
	"strings"
	"sync"
	"time"
	"bufio"
	"path/filepath"
)

type FileSystem struct{
	root *Directory,
	mu sync.RwMutex,
	openFiles int
} 

func NewFileSystem() *FileSystem{
	return &FileSystem{
		root =&Directory{
			name="/",
			children=map[string]FSNode,
			createdAt=time.Now(),
			updatedAt=time.Now(),
		},
	}
}

func (fs *FileSystem) splitPath(path string) ([]string){
	path=filepath.Clean(path)

	if path=="/"{
		return []string{}
	}

	path.TrimPrefix(path,"/")

	return string.Split(path,"/")
}

func (fs *FileSystem) getNode(path string) (*FSNode,error{
	path=filepath.Clean(path)

	if path=="/"{
		return fs.root,nil
	}

	components:=splitPath(path)
	currentDir="/"

	for i,component in range components{
		child,exists=currentDir.children[component]

		if !exists{
			return nil,NewFSError(OpStat,path,ErrNotExist)
		}

		if i=len(components)-1{
			return child,nil
		}

		if !child.IsDir(){
			return nil,NewFSError(OpStat,path,ErrNotDir)
		}

		currentDir=child.(*Directory)
	} 
	return currentDir,nil
}

func (fs *FileSystem) getParentDir(path string) (*Directory, string, error){
	path=filepath.Clean(path)

	if(path == "/"){
		return nil, "",NewFSError(OpStat,path,ErrIsRoot)
	}
	parentPath := filepath.Dir(path)
	baseName := filepath.Base(path)

	node,err := fs.getNode(parentPath)

	if err!=nil{
		return nil,"",err
	}

	dir,ok=node.(*Directory)
	if !ok{
		return nil,"",NewFSError(OpStat,parentPath,ErrNotDir)
	}

	return dir,baseName,nil
}

func (fs *FileSystem) getOpenFileCount() int{
	mu.RLock()
	defer mu.RUnlock()
	return fs.openFiles
}

func (fs *FileSystem) decrementOpenFiles(){
	mu.Lock()
	defer mu.Unlock()
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
