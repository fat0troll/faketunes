package transcoder

import (
	"path/filepath"
	"strings"
)

func (t *Transcoder) escapeMetadata(item string) string {
	// Escape quotes and backslashes for FFmpeg metadata
	item = strings.ReplaceAll(item, `\`, `\\`)
	item = strings.ReplaceAll(item, `"`, `\"`)
	item = strings.ReplaceAll(item, `'`, `\'`)

	// Also escape semicolons and equals signs
	item = strings.ReplaceAll(item, `;`, `\;`)
	item = strings.ReplaceAll(item, `=`, `\=`)

	return item
}

func (t *Transcoder) extractAlbumArtist(filePath, sourceDir string) string {
	// Get relative path from source directory
	relPath, err := filepath.Rel(sourceDir, filePath)
	if err != nil {
		return "Unknown Artist"
	}

	// Split path into components
	parts := strings.Split(relPath, string(filepath.Separator))

	// Album artist is the first directory after source
	// e.g., /source/Artist/Album/01 - Track Name.flac
	if len(parts) >= 2 {
		artist := parts[0]
		artist = strings.TrimSpace(artist)

		return artist
	}

	return "Unknown Artist"
}
