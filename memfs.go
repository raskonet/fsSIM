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
