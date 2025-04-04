package memfs_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/raskonet/fsSIM/memfs"
)

func TestFSError(t *testing.T) {
	err := memfs.NewFSError(memfs.OpOpen, "/file.txt", memfs.ErrNotExist)
	assert.Equal(t, "open /file.txt: file does not exist", err.Error())
}

func TestIsNotExist(t *testing.T) {
	assert.True(t, memfs.IsNotExist(memfs.ErrNotExist))
	assert.False(t, memfs.IsNotExist(errors.New("random error")))
}

func TestIsExist(t *testing.T) {
	assert.True(t, memfs.IsExist(memfs.ErrExist))
	assert.False(t, memfs.IsExist(errors.New("random error")))
}

func TestIsDir(t *testing.T) {
	assert.True(t, memfs.IsDir(memfs.ErrIsDir))
	assert.False(t, memfs.IsDir(errors.New("random error")))
}

func TestIsNotDir(t *testing.T) {
	assert.True(t, memfs.IsNotDir(memfs.ErrNotDir))
	assert.False(t, memfs.IsNotDir(errors.New("random error")))
}

