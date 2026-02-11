package filesystem

import (
	"context"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type MusicAppMetadataFile struct {
	fs.Inode

	f    *FS
	path string
}

var (
	_ = (fs.NodeGetattrer)((*MusicAppMetadataFile)(nil))
	_ = (fs.NodeOpener)((*MusicAppMetadataFile)(nil))
	_ = (fs.NodeCreater)((*MusicAppMetadataFile)(nil))
	_ = (fs.NodeWriter)((*MusicAppMetadataFile)(nil))
	_ = (fs.NodeSetattrer)((*MusicAppMetadataFile)(nil))
	_ = (fs.NodeUnlinker)((*MusicAppMetadataFile)(nil))
)

func (m *MusicAppMetadataFile) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	info, err := os.Stat(m.path)
	if err != nil {
		out.Mode = fuse.S_IFREG | 0o644
		out.Nlink = 1
		out.Ino = m.StableAttr().Ino
		out.Size = 0
		out.Mtime = uint64(time.Now().Unix())
		out.Atime = out.Mtime
		out.Ctime = out.Mtime
		out.Blocks = 1

		return 0
	}

	out.Mode = fuse.S_IFREG | uint32(info.Mode())
	out.Nlink = 1
	out.Ino = m.StableAttr().Ino
	out.Size = uint64(info.Size())
	out.Mtime = uint64(info.ModTime().Unix())
	out.Atime = out.Mtime
	out.Ctime = out.Mtime
	out.Blocks = (out.Size + 511) / 512

	return 0
}

func (m *MusicAppMetadataFile) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	info, err := os.Stat(m.path)
	if err != nil {
		out.Mode = fuse.S_IFREG | 0o644
		out.Nlink = 1
		out.Ino = m.StableAttr().Ino
		out.Size = 0
		out.Mtime = uint64(time.Now().Unix())
		out.Atime = out.Mtime
		out.Ctime = out.Mtime
		out.Blocks = 1
	} else {
		out.Mode = fuse.S_IFREG | uint32(info.Mode())
		out.Nlink = 1
		out.Ino = m.StableAttr().Ino
		out.Size = uint64(info.Size())
		out.Mtime = uint64(info.ModTime().Unix())
		out.Atime = out.Mtime
		out.Ctime = out.Mtime
		out.Blocks = (out.Size + 511) / 512
	}

	return 0
}

func (m *MusicAppMetadataFile) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	file, err := os.Create(m.path)
	if err != nil {
		return nil, nil, 0, syscall.EIO
	}

	ch := m.NewInode(ctx, &MusicAppMetadataFile{path: m.path}, fs.StableAttr{
		Mode: fuse.S_IFREG,
		Ino:  m.f.nextInode(),
	})

	out.Mode = fuse.S_IFREG | 0o644
	out.Nlink = 1
	out.Ino = ch.StableAttr().Ino
	out.Size = 0
	out.Mtime = uint64(time.Now().Unix())
	out.Atime = out.Mtime
	out.Ctime = out.Mtime
	out.Blocks = 1

	return ch, &File{file: file}, fuse.FOPEN_DIRECT_IO, 0
}

func (m *MusicAppMetadataFile) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		if err := os.WriteFile(m.path, []byte{}, 0o644); err != nil {
			return nil, 0, syscall.EIO
		}
	}

	file, err := os.OpenFile(m.path, int(flags), 0o644)
	if err != nil {
		return nil, 0, syscall.EIO
	}

	return &File{file: file}, fuse.FOPEN_DIRECT_IO, 0
}

func (m *MusicAppMetadataFile) Write(ctx context.Context, fh fs.FileHandle, data []byte, off int64) (written uint32, errno syscall.Errno) {
	handle, ok := fh.(*File)
	if !ok {
		return 0, syscall.EBADF
	}

	n, err := handle.file.WriteAt(data, off)
	if err != nil {
		return 0, syscall.EIO
	}

	return uint32(n), 0
}

func (m *MusicAppMetadataFile) Unlink(ctx context.Context, name string) syscall.Errno {
	if err := os.Remove(m.path); err != nil {
		return syscall.ENOENT
	}

	return 0
}

func (f *FS) isiTunesMetadata(name string) bool {
	name = strings.ToLower(name)

	return strings.HasPrefix(name, ".") ||
		strings.Contains(name, "albumart") ||
		strings.Contains(name, "folder") ||
		strings.Contains(name, "itunes") ||
		strings.HasSuffix(name, ".itl") ||
		strings.HasSuffix(name, ".xml") ||
		strings.HasSuffix(name, ".db")
}

func (f *FS) NewMusicAppMetadataFile(path string) *MusicAppMetadataFile {
	return &MusicAppMetadataFile{
		f:    f,
		path: path,
	}
}
