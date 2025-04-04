package memfs

import (
	"io"
	"time"
)

type FSNode interface {
	IsDir() bool
	Name() string
	UpdatedAt() time.Time
	CreatedAt() time.Time
}

type FSInfo struct {
	IsDir     bool
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
	Size      int64
}

type File struct {
	name      string
	updatedAt time.Time
	createdAt time.Time
	position  int64
	content   []byte
	fs        *FileSystem
}

type Directory struct {
	name      string
	children  map[string]FSNode
	createdAt time.Time
	updatedAt time.Time
}

func (f *File) Name() string {
	return f.name
}

func (f *File) IsDir() bool {
	return false
}

func (f *File) CreatedAt() time.Time {
	return f.createdAt
}

func (f *File) UpdatedAt() time.Time {
	return f.updatedAt
}

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) IsDir() bool {
	return true
}

func (d *Directory) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Directory) CreatedAt() time.Time {
	return d.createdAt
}

func (f *File) Size() int {
	return len(f.content)
}

func (f *File) Read(p []byte) (int, error) {
	if f.position >= int64(len(f.content)) {
		return 0, io.EOF
	}
	n := copy(p, f.content[f.position:])
	f.position += int64(n)
	return n, nil
}

func (f *File) Write(p []byte) (int, error) {
	requiredLength := len(p) + int(f.position)
	if requiredLength > len(f.content) {
		newBuffer := make([]byte, requiredLength)
		copy(newBuffer, f.content)
		f.content = newBuffer
	}
	n := copy(f.content[f.position:], p)
	f.position += int64(n)
	f.updatedAt = time.Now()
	return n, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	var newPosition int64

	switch whence {
	case io.SeekStart:
		newPosition = offset
	case io.SeekCurrent:
		newPosition = f.position + offset
	case io.SeekEnd:
		newPosition = int64(len(f.content)) + offset
	default:
		return 0, NewFSError(OpSeek, f.name, ErrNegativeOffset)
	}

	if newPosition < 0 {
		return 0, NewFSError(OpSeek, f.name, ErrNegativeOffset)
	}

	f.position = newPosition
	return newPosition, nil
}

func (f *File) Close() error {
	f.fs.decrementOpenFiles()
	return nil
}

func (f *File) Truncate(size int64) error {
	if size < 0 {
		return NewFSError(OpTruncate, f.name, ErrNegativeOffset)
	}

	newBuffer := make([]byte, size)
	if size > int64(len(f.content)) {
		copy(newBuffer, f.content)
	} else {
		copy(newBuffer, f.content[:size])
	}

	f.content = newBuffer
	if f.position > size {
		f.position = size
	}

	f.updatedAt = time.Now()
	return nil
}

