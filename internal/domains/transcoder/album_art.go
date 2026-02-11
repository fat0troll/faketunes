package transcoder

import (
	"os"
	"path/filepath"
)

func (t *Transcoder) findAlbumArt(path string) string {
	// Common album art filenames (in order of preference)
	artFiles := []string{
		"albumart.jpg",
		"AlbumArt.jpg",
		"cover.jpg",
		"Cover.jpg",
		"folder.jpg",
		"Folder.jpg",
		"albumart.jpeg",
		"cover.jpeg",
		"folder.jpeg",
		"albumart.png",
		"cover.png",
		"folder.png",
		"albumart.gif",
		"cover.gif",
		".albumart.jpg",
		".cover.jpg",
		"AlbumArtwork.jpg",
		"album.jpg",
		"Album.jpg",
	}

	for _, artFile := range artFiles {
		fullPath := filepath.Join(path, artFile)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}
