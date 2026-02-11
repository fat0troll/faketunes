package filesystem

import (
	"log"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

func (f *FS) mount() {
	rootDir := f.NewRootDirectory()

	// Populate mount options
	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Name:          "faketunes",
			FsName:        "faketunes",
			DisableXAttrs: false, // Enable xattr support for macOS
			Debug:         false,
			// AllowOther:    true,
			Options: []string{
				"default_permissions",
				"fsname=flac2alac",
				"nosuid",
				"nodev",
				"noexec",
				"ro",
			},
		},
		NullPermissions: false,
		Logger:          log.New(os.Stdout, "FUSE: ", log.LstdFlags),
	}

	// Redirect FUSE logs to logrus
	log.SetOutput(f.app.Logger().WithField("fuse debug logs", true).WriterLevel(logrus.DebugLevel))

	// Do an actual mount
	server, err := fs.Mount(f.destinationDir, rootDir, opts)
	if err != nil {
		f.app.Logger().WithError(err).Fatal("Failed to start filesystem")
	}
	defer server.Unmount()

	select {
	case <-f.app.Context().Done():
		return
	default:
		server.Wait()
	}
}
