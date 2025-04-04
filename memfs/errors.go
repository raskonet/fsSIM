
package memfs

import (
	"errors"
	"fmt"
)


type Operation string


const (
	OpCreate    Operation = "create"
	OpOpen      Operation = "open"
	OpClose     Operation = "close"
	OpRead      Operation = "read"
	OpWrite     Operation = "write"
	OpSeek      Operation = "seek"
	OpMkdir     Operation = "mkdir"
	OpReadDir   Operation = "readdir"
	OpReadFile  Operation = "readfile"
	OpWriteFile Operation = "writefile"
	OpRemove    Operation = "remove"
	OpRemoveAll Operation = "removeall"
	OpRename    Operation = "rename"
	OpStat      Operation = "stat"
	OpTruncate  Operation = "truncate"
)

// Standard file system errors
var (
	ErrNotExist       = errors.New("file does not exist")
	ErrExist          = errors.New("file already exists")
	ErrNotDir         = errors.New("not a directory")
	ErrIsDir          = errors.New("is a directory")
	ErrPermission     = errors.New("permission denied")
	ErrInvalidOffset  = errors.New("invalid offset")
	ErrNegativeOffset = errors.New("negative offset")
	ErrInvalidWhence  = errors.New("invalid whence value")
	ErrClosed         = errors.New("file already closed")
	ErrDirNotEmpty    = errors.New("directory not empty")
	ErrIsRoot         = errors.New("operation not permitted on root")
	ErrRootExists     = errors.New("root already exists")
)


type FSError struct {
	Op   Operation
	Path string
	Err  error
}


func NewFSError(op Operation, path string, err error) *FSError {
	return &FSError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}


func (e *FSError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}


func (e *FSError) Unwrap() error {
	return e.Err
}


func IsNotExist(err error) bool {
	return errors.Is(err, ErrNotExist)
}


func IsExist(err error) bool {
	return errors.Is(err, ErrExist)
}


func IsNotDir(err error) bool {
	return errors.Is(err, ErrNotDir)
}


func IsDir(err error) bool {
	return errors.Is(err, ErrIsDir)
}


func IsDirNotEmpty(err error) bool {
	return errors.Is(err, ErrDirNotEmpty)
}


func IsPermission(err error) bool {
	return errors.Is(err, ErrPermission)
}
