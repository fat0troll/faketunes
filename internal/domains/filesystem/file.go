package filesystem

import (
	"context"
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type File struct {
	file      *os.File
	fileMutex sync.Mutex
}

var (
	_ = (fs.FileReader)((*File)(nil))
	_ = (fs.FileWriter)((*File)(nil))
	_ = (fs.FileFlusher)((*File)(nil))
	_ = (fs.FileReleaser)((*File)(nil))
)

func (fi *File) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	fi.fileMutex.Lock()
	defer fi.fileMutex.Unlock()

	_, err := fi.file.Seek(off, io.SeekStart)
	if err != nil {
		return nil, syscall.EIO
	}

	n, err := fi.file.Read(dest)
	if err != nil && err != io.EOF {
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(dest[:n]), 0
}

func (fi *File) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	fi.fileMutex.Lock()
	defer fi.fileMutex.Unlock()

	n, err := fi.file.WriteAt(data, off)
	if err != nil {
		return 0, syscall.EIO
	}

	return uint32(n), 0
}

func (fi *File) Flush(ctx context.Context) syscall.Errno {
	fi.fileMutex.Lock()
	defer fi.fileMutex.Unlock()

	if err := fi.file.Sync(); err != nil {
		return syscall.EIO
	}

	return 0
}

func (fi *File) Release(ctx context.Context) syscall.Errno {
	fi.fileMutex.Lock()
	defer fi.fileMutex.Unlock()

	if err := fi.file.Close(); err != nil {
		return syscall.EIO
	}

	return 0
}
