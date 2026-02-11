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

type RootDirectory struct {
	fs.Inode

	f *FS
}

var (
	_ = (fs.NodeGetattrer)((*RootDirectory)(nil))
	_ = (fs.NodeLookuper)((*RootDirectory)(nil))
	_ = (fs.NodeReaddirer)((*RootDirectory)(nil))
	_ = (fs.NodeCreater)((*RootDirectory)(nil))
	_ = (fs.NodeGetxattrer)((*RootDirectory)(nil))
	_ = (fs.NodeSetxattrer)((*RootDirectory)(nil))
	_ = (fs.NodeRemovexattrer)((*RootDirectory)(nil))
	_ = (fs.NodeListxattrer)((*RootDirectory)(nil))
)

func (r *RootDirectory) Create(
	ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut,
) (*fs.Inode, fs.FileHandle, uint32, syscall.Errno) {
	if r.f.isiTunesMetadata(name) {
		metaPath := filepath.Join(r.f.metadataDir, name)

		file, err := os.Create(metaPath)
		if err != nil {
			return nil, nil, 0, syscall.EIO
		}

		ch := r.NewInode(
			ctx,
			r.f.NewMusicAppMetadataFile(metaPath),
			fs.StableAttr{
				Mode: fuse.S_IFREG,
				Ino:  r.f.nextInode(),
			},
		)

		out.Mode = fuse.S_IFREG | 0644
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

func (r *RootDirectory) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if r.f.isiTunesMetadata(name) {
		metaPath := filepath.Join(r.f.metadataDir, name)
		ch := r.NewInode(
			ctx,
			r.f.NewMusicAppMetadataFile(metaPath),
			fs.StableAttr{
				Mode: fuse.S_IFREG,
				Ino:  r.f.nextInode(),
			},
		)

		out.Mode = fuse.S_IFREG | 0644
		out.Nlink = 1
		out.Ino = ch.StableAttr().Ino

		if info, err := os.Stat(metaPath); err == nil {
			out.Size = uint64(info.Size())
			out.Mtime = uint64(info.ModTime().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
		} else {
			out.Size = 0
			out.Mtime = uint64(time.Now().Unix())
			out.Atime = out.Mtime
			out.Ctime = out.Mtime
		}

		// Calculate blocks
		out.Blocks = (out.Size + 511) / 512

		return ch, 0
	}

	// Handle .m4a virtual files
	if strings.HasSuffix(strings.ToLower(name), ".m4a") {
		flacName := name[:len(name)-4] + ".flac"
		flacPath := filepath.Join(r.f.sourceDir, flacName)

		if _, err := os.Stat(flacPath); err == nil {
			ch := r.NewInode(
				ctx,
				r.f.NewMusicFile(flacPath, name, false),
				fs.StableAttr{
					Mode: fuse.S_IFREG,
					Ino:  r.f.nextInode(),
				},
			)

			out.Mode = fuse.S_IFREG | 0444
			out.Nlink = 1
			out.Ino = ch.StableAttr().Ino

			if size, err := r.f.cacher.GetStat(flacPath); err == nil {
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
	fullPath := filepath.Join(r.f.sourceDir, name)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, syscall.ENOENT
	}

	if info.IsDir() {
		ch := r.NewInode(ctx, r.f.NewMusicDirectory(fullPath), fs.StableAttr{
			Mode: fuse.S_IFDIR,
			Ino:  r.f.nextInode(),
		})

		out.Mode = fuse.S_IFDIR | 0755
		out.Nlink = 2 // Minimum . and ..
		out.Ino = ch.StableAttr().Ino
		out.Size = 4096
		out.Mtime = uint64(info.ModTime().Unix())
		out.Atime = out.Mtime
		out.Ctime = out.Mtime
		out.Blocks = 1

		return ch, 0
	}

	// Regular file (non-FLAC)
	isMeta := r.f.isiTunesMetadata(name)
	ch := r.NewInode(ctx, r.f.NewMusicFile(fullPath, name, isMeta), fs.StableAttr{
		Mode: fuse.S_IFREG,
		Ino:  r.f.nextInode(),
	})

	if isMeta {
		out.Mode = fuse.S_IFREG | 0644
	} else {
		out.Mode = fuse.S_IFREG | 0444
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

func (r *RootDirectory) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r.f.app.Logger().WithField("path", r.f.sourceDir).Debug("Readdir called on directory")

	var dirEntries []fuse.DirEntry

	// Always include . and .. first
	dirEntries = append(dirEntries, fuse.DirEntry{
		Name: ".",
		Mode: fuse.S_IFDIR | 0755,
		Ino:  1, // Root inode
	})
	dirEntries = append(dirEntries, fuse.DirEntry{
		Name: "..",
		Mode: fuse.S_IFDIR | 0755,
		Ino:  1,
	})

	// Read actual directory contents
	entries, err := os.ReadDir(r.f.sourceDir)
	if err != nil {
		r.f.app.Logger().WithError(err).WithField("path", r.f.sourceDir).Error(
			"Error reading directory",
		)
		return fs.NewListDirStream(dirEntries), 0
	}

	for _, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, ".") && !r.f.isiTunesMetadata(name) {
			continue
		}

		mode := fuse.S_IFREG | 0444
		if entry.IsDir() {
			mode = fuse.S_IFDIR | 0755
		}

		// Convert .flac to .m4a in directory listing
		if strings.HasSuffix(strings.ToLower(name), ".flac") {
			name = name[:len(name)-5] + ".m4a"
		}

		mode = fuse.S_IFREG | 0644

		dirEntries = append(dirEntries, fuse.DirEntry{
			Name: name,
			Mode: uint32(mode),
			Ino:  r.f.nextInode(),
		})
	}

	r.f.app.Logger().WithFields(logrus.Fields{
		"path":              r.f.sourceDir,
		"directory entries": len(dirEntries),
	}).Debug("Returning directory entries")

	return fs.NewListDirStream(dirEntries), 0
}

func (r *RootDirectory) Getattr(
	ctx context.Context, f fs.FileHandle, out *fuse.AttrOut,
) syscall.Errno {
	// Set basic directory attributes
	out.Mode = fuse.S_IFDIR | 0755

	// Set nlink to at least 2 (for . and ..)
	out.Nlink = 2

	// Root directory typically has inode 1
	out.Ino = 1

	// Set size to typical directory size
	out.Size = 4096

	// Set timestamps
	now := uint64(time.Now().Unix())
	out.Mtime = now
	out.Atime = now
	out.Ctime = now

	// Set blocks (1 block of 512 bytes each = 512 bytes)
	out.Blocks = 1

	// Set block size
	out.Blksize = 512

	// Count actual subdirectories for accurate nlink
	if entries, err := os.ReadDir(r.f.sourceDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				out.Nlink++
			}
		}
	}

	return 0
}

func (r *RootDirectory) Getxattr(
	ctx context.Context, attr string, dest []byte,
) (uint32, syscall.Errno) {
	// Handle common macOS/Netatalk xattrs
	switch attr {
	case "user.org.netatalk.Metadata":
		fallthrough
	case "com.apple.FinderInfo":
		fallthrough
	case "com.apple.ResourceFork":
		// Return empty data
		if len(dest) > 0 {
			return 0, 0
		}

		return 0, 0
	default:
		return 0, syscall.ENODATA
	}
}

func (r *RootDirectory) Setxattr(ctx context.Context, attr string, data []byte, flags uint32) syscall.Errno {
	// Silently accept xattr writes (ignore them)
	return 0
}

func (r *RootDirectory) Removexattr(ctx context.Context, attr string) syscall.Errno {
	return 0
}

func (r *RootDirectory) Listxattr(ctx context.Context, dest []byte) (uint32, syscall.Errno) {
	// Return empty xattr list
	return 0, 0
}

func (f *FS) NewRootDirectory() *RootDirectory {
	return &RootDirectory{
		f: f,
	}
}
