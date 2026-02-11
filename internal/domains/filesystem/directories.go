package filesystem

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

func (f *FS) prepareDirectories() error {
	if _, err := os.Stat(f.sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %w (%w)", ErrFilesystem, ErrNoSource, err)
	}

	f.app.Logger().WithField("path", f.sourceDir).Info("Got source directory")

	// Clean destination directory
	if _, err := os.Stat(f.destinationDir); err == nil {
		f.app.Logger().WithField("path", f.destinationDir).Info(
			"Cleaning up the destination mountpoint",
		)

		// Try to unmount the destination FS if that was mounted before.
		exec.Command("fusermount3", "-u", f.destinationDir).Run()
		time.Sleep(5 * time.Second)

		// Clean the destination
		err := os.RemoveAll(f.destinationDir)
		if err != nil {
			return fmt.Errorf("%w: %w (%w)", ErrFilesystem, ErrFailedToCleanupDestination, err)
		}
	}

	// Create the structure for the virtual filesystem.
	for _, dir := range []string{f.destinationDir, f.cacheDir, f.metadataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			f.app.Logger().WithField("path", dir).Error("Operation on directory was unsuccessful")

			return fmt.Errorf("%w: %w (%w)", ErrFilesystem, ErrFailedToCreateDestinationDirectory, err)
		}
	}

	f.app.Logger().WithFields(logrus.Fields{
		"source directory":         f.sourceDir,
		"virtual filesystem mount": f.destinationDir,
		"cache directory":          f.cacheDir,
		"metadata directory":       f.metadataDir,
	}).Debug("Filesystem directories prepared")

	return nil
}
