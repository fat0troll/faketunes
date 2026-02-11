package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type MusicFile struct {
	fs.Inode

	f           *FS
	sourcePath  string
	virtualName string
	isMetaFile  bool
}

var (
	_ = (fs.NodeGetattrer)((*MusicFile)(nil))
	_ = (fs.NodeOpener)((*MusicFile)(nil))
	_ = (fs.NodeSetattrer)((*MusicFile)(nil))
)

func (f *MusicFile) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	if f.isMetaFile {
		metaPath := filepath.Join(f.f.metadataDir, f.virtualName)

		if info, err := os.Stat(metaPath); err == nil {
			out.Mode = fuse.S_IFREG | 0644
			out.Nlink = 1
			out.Ino = f.StableAttr().Ino
			out.Size = uint64(info.Size())
			out.Mtime = uint64(info.ModTime().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
			out.Blocks = (out.Size + 511) / 512
		} else {
			out.Mode = fuse.S_IFREG | 0644
			out.Nlink = 1
			out.Ino = f.StableAttr().Ino
			out.Size = 0
			out.Mtime = uint64(time.Now().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
			out.Blocks = 1
		}

		return 0
	}

	out.Mode = fuse.S_IFREG | 0444
	out.Nlink = 1
	out.Ino = f.StableAttr().Ino
	out.Blocks = 1

	if size, err := f.f.cacher.GetStat(f.sourcePath); err == nil {
		out.Size = uint64(size)
		out.Blocks = (out.Size + 511) / 512
	} else {
		out.Size = 0
	}

	out.Mtime = uint64(time.Now().Unix())
	out.Atime = out.Mtime
	out.Ctime = out.Mtime

	return 0
}

func (f *MusicFile) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	if f.isMetaFile {
		metaPath := filepath.Join(f.f.metadataDir, f.virtualName)
		if info, err := os.Stat(metaPath); err == nil {
			out.Mode = fuse.S_IFREG | 0644
			out.Nlink = 1
			out.Ino = f.StableAttr().Ino
			out.Size = uint64(info.Size())
			out.Mtime = uint64(info.ModTime().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
			out.Blocks = (out.Size + 511) / 512
		} else {
			out.Mode = fuse.S_IFREG | 0644
			out.Nlink = 1
			out.Ino = f.StableAttr().Ino
			out.Size = 0
			out.Mtime = uint64(time.Now().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
			out.Blocks = 1
		}

		return 0
	}

	return syscall.EPERM
}

func (f *MusicFile) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if f.isMetaFile {
		metaPath := filepath.Join(f.f.metadataDir, f.virtualName)

		file, err := os.OpenFile(metaPath, int(flags), 0644)
		if err != nil && os.IsNotExist(err) {
			file, err = os.Create(metaPath)
		}
		if err != nil {
			return nil, 0, syscall.EIO
		}

		return &File{file: file}, fuse.FOPEN_DIRECT_IO, 0
	}

	if flags&fuse.O_ANYWRITE != 0 {
		return nil, 0, syscall.EPERM
	}

	entry, err := f.f.cacher.GetFileDTO(f.sourcePath)
	if err != nil {
		f.f.app.Logger().WithError(err).WithField("source file", f.sourcePath).
			WithError(err).Error("Failed to convert file to cache")

		return nil, 0, syscall.EIO
	}

	f.f.app.Logger().WithField("path", entry.Path).Debug("Opening cached file")

	file, err := os.Open(entry.Path)
	if err != nil {
		return nil, 0, syscall.EIO
	}

	return &File{file: file}, fuse.FOPEN_KEEP_CACHE, 0
}

func (f *FS) NewMusicFile(sourcePath, virtualName string, isMetaFile bool) *MusicFile {
	return &MusicFile{
		f:           f,
		sourcePath:  sourcePath,
		virtualName: virtualName,
		isMetaFile:  isMetaFile,
	}
}
