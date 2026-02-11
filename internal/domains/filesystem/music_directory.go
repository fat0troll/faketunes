package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

// Any non-root directory is a MusicDirectory

type MusicDir struct {
	fs.Inode

	f    *FS
	path string
}

var (
	_ = (fs.NodeGetattrer)((*MusicDir)(nil))
	_ = (fs.NodeLookuper)((*MusicDir)(nil))
	_ = (fs.NodeReaddirer)((*MusicDir)(nil))
	_ = (fs.NodeCreater)((*MusicDir)(nil))
	_ = (fs.NodeGetxattrer)((*MusicDir)(nil))
	_ = (fs.NodeSetxattrer)((*MusicDir)(nil))
	_ = (fs.NodeRemovexattrer)((*MusicDir)(nil))
	_ = (fs.NodeListxattrer)((*MusicDir)(nil))
)

func (d *MusicDir) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFDIR | 0o755
	out.Nlink = 2 // Minimum . and ..
	out.Ino = d.StableAttr().Ino
	out.Size = 4096

	// Get actual mod time from filesystem if possible
	if info, err := os.Stat(d.path); err == nil {
		out.Mtime = uint64(info.ModTime().Unix())
		out.Atime = out.Mtime
		out.Ctime = out.Mtime

		// Count actual subdirectories for accurate nlink
		if entries, err := os.ReadDir(d.path); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					out.Nlink++
				}
			}
		}
	} else {
		now := uint64(time.Now().Unix())
		out.Mtime = now
		out.Atime = now
		out.Ctime = now
	}

	out.Blocks = 1
	out.Blksize = 512

	return 0
}

func (d *MusicDir) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
	// Same implementation as RootDir
	switch attr {
	case "user.org.netatalk.Metadata":
		fallthrough
	case "com.apple.FinderInfo":
		fallthrough
	case "com.apple.ResourceFork":
		if len(dest) > 0 {
			return 0, 0
		}

		return 0, 0
	default:
		return 0, syscall.ENODATA
	}
}

func (d *MusicDir) Setxattr(ctx context.Context, attr string, data []byte, flags uint32) syscall.Errno {
	return 0
}

func (d *MusicDir) Removexattr(ctx context.Context, attr string) syscall.Errno {
	return 0
}

func (d *MusicDir) Listxattr(ctx context.Context, dest []byte) (uint32, syscall.Errno) {
	return 0, 0
}

func (d *MusicDir) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (*fs.Inode, fs.FileHandle, uint32, syscall.Errno) {
	if d.f.isiTunesMetadata(name) {
		metaPath := filepath.Join(d.f.metadataDir, name)

		file, err := os.Create(metaPath)
		if err != nil {
			return nil, nil, 0, syscall.EIO
		}

		ch := d.NewInode(
			ctx,
			d.f.NewMusicAppMetadataFile(metaPath),
			fs.StableAttr{
				Mode: fuse.S_IFREG,
				Ino:  d.f.nextInode(),
			},
		)

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

	return nil, nil, 0, syscall.EPERM
}

func (d *MusicDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Handle .m4a virtual files
	if strings.HasSuffix(strings.ToLower(name), ".m4a") {
		flacName := name[:len(name)-4] + ".flac"
		flacPath := filepath.Join(d.path, flacName)

		if _, err := os.Stat(flacPath); err == nil {
			ch := d.NewInode(
				ctx,
				d.f.NewMusicFile(flacPath, name, false),
				fs.StableAttr{
					Mode: fuse.S_IFREG,
					Ino:  d.f.nextInode(),
				},
			)

			out.Mode = fuse.S_IFREG | 0o444
			out.Nlink = 1
			out.Ino = ch.StableAttr().Ino

			if size, err := d.f.cacher.GetStat(flacPath); err == nil {
				out.Size = uint64(size)
			} else {
				out.Size = 0
			}

			out.Mtime = uint64(time.Now().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
			out.Blocks = (out.Size + 511) / 512

			return ch, 0
		}
	}

	// Check real file or directory
	fullPath := filepath.Join(d.path, name)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, syscall.ENOENT
	}

	if info.IsDir() {
		ch := d.NewInode(
			ctx, d.f.NewMusicDirectory(fullPath),
			fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  d.f.nextInode(),
			},
		)

		out.Mode = fuse.S_IFDIR | 0o755
		out.Nlink = 2
		out.Ino = ch.StableAttr().Ino
		out.Size = 4096
		out.Mtime = uint64(info.ModTime().Unix())
		out.Atime = out.Mtime
		out.Ctime = out.Mtime
		out.Blocks = 1

		return ch, 0
	}

	// Regular file (non-FLAC)
	isMeta := d.f.isiTunesMetadata(name)
	ch := d.NewInode(ctx, d.f.NewMusicFile(fullPath, name, isMeta),
		fs.StableAttr{
			Mode: fuse.S_IFREG,
			Ino:  d.f.nextInode(),
		},
	)

	if isMeta {
		out.Mode = fuse.S_IFREG | 0o644
	} else {
		out.Mode = fuse.S_IFREG | 0o444
	}

	out.Nlink = 1
	out.Ino = ch.StableAttr().Ino
	out.Size = uint64(info.Size())
	out.Mtime = uint64(info.ModTime().Unix())
	out.Atime = out.Mtime
	out.Ctime = out.Mtime
	out.Blocks = (out.Size + 511) / 512

	return ch, 0
}

func (d *MusicDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	d.f.app.Logger().WithField("path", d.path).Debug("Readdir called on directory")

	var dirEntries []fuse.DirEntry

	dirEntries = append(dirEntries, fuse.DirEntry{
		Name: ".",
		Mode: fuse.S_IFDIR | 0o755,
		Ino:  d.StableAttr().Ino,
	})
	dirEntries = append(dirEntries, fuse.DirEntry{
		Name: "..",
		Mode: fuse.S_IFDIR | 0o755,
		Ino:  1, // Parent (root) inode
	})

	entries, err := os.ReadDir(d.path)
	if err != nil {
		d.f.app.Logger().WithError(err).WithField("path", d.path).Error(
			"Error reading directory",
		)

		return fs.NewListDirStream(dirEntries), 0
	}

	for _, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, ".") && !d.f.isiTunesMetadata(name) {
			continue
		}

		mode := fuse.S_IFREG | 0o444
		if entry.IsDir() {
			mode = fuse.S_IFDIR | 0o755
		}

		// Convert .flac to .m4a in directory listing
		if strings.HasSuffix(strings.ToLower(name), ".flac") {
			name = name[:len(name)-5] + ".m4a"
			if !d.f.isiTunesMetadata(name) {
				mode = fuse.S_IFREG | 0o644
			}
		} else if !d.f.isiTunesMetadata(name) {
			mode = fuse.S_IFREG | 0o644
		}

		dirEntries = append(dirEntries, fuse.DirEntry{
			Name: name,
			Mode: uint32(mode),
			Ino:  d.f.nextInode(),
		})
	}

	d.f.app.Logger().WithFields(logrus.Fields{
		"path":              d.path,
		"directory entries": len(dirEntries),
	}).Debug("Returning directory entries")

	return fs.NewListDirStream(dirEntries), 0
}

func (f *FS) NewMusicDirectory(path string) *MusicDir {
	return &MusicDir{
		f:    f,
		path: path,
	}
}
