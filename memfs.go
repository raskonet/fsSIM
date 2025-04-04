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
