package transcoder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	defaultSampleRate = 48000
	defaultBitDepth   = 16
)

// Convert converts the file from FLAC to ALAC using ffmpeg.
// It embeds all required metadata and places the file in the desired destination.
// On success, it returns the transcoded file's size.
func (t *Transcoder) Convert(sourcePath, destinationPath string) (int64, error) {
	t.app.Logger().WithFields(logrus.Fields{
		"source file": sourcePath,
		"destination": destinationPath,
	}).Info("Transcoding file using ffmpeg...")

	sourceAlbumDir := filepath.Dir(sourcePath)
	albumArt := t.findAlbumArt(sourceAlbumDir)
	hasAlbumArt := albumArt != ""
	sortArtist := t.extractAlbumArtist(sourcePath, sourceAlbumDir)
	sampleRate := defaultSampleRate
	bitDepth := defaultBitDepth

	if hasAlbumArt {
		t.app.Logger().WithField("album art path", albumArt).Debug("Found album art")
	}

	t.app.Logger().WithField("sort artist", sortArtist).Debug(
		"Setting sorting artist for iTunes",
	)

	sourceAnalyzeCmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-show_streams",
		"-select_streams", "a:0",
		"-of", "csv=p=0",
		sourcePath,
	)

	analyzeOutput, err := sourceAnalyzeCmd.Output()
	if err == nil {
		// Investiage bit depth and sample rate from ffprobe output.
		// We need that to make sure we don't oversample files that are lower
		// than the default sample rate and bit depth.
		lines := strings.Split(strings.TrimSpace(string(analyzeOutput)), "\n")
		for _, line := range lines {
			if strings.Contains(line, "audio") {
				parts := strings.Split(line, ",")
				if len(parts) >= 6 {
					// Get sample rate
					if sr, err := strconv.Atoi(parts[2]); err == nil && sr > 0 {
						sampleRate = sr
					}

					// Get bit depth from sample_fmt or bits_per_raw_sample
					sampleFmt := parts[4]
					if strings.Contains(sampleFmt, "s32") || strings.Contains(sampleFmt, "flt") {
						bitDepth = 32
					} else if strings.Contains(sampleFmt, "s64") || strings.Contains(sampleFmt, "dbl") {
						bitDepth = 64
					} else if len(parts) >= 6 && parts[5] != "N/A" && parts[5] != "" {
						if bd, err := strconv.Atoi(parts[5]); err == nil && bd > 0 {
							bitDepth = bd
						}
					}
				}

				break // We only need the first audio stream
			}
		}
	}

	t.app.Logger().WithFields(logrus.Fields{
		"bit depth":   bitDepth,
		"sample rate": sampleRate,
	}).Info("Detected source file sample rate and bit depth")

	needsDownsample := sampleRate > defaultSampleRate
	needsBitReduce := bitDepth > defaultBitDepth

	if needsDownsample {
		t.app.Logger().WithFields(logrus.Fields{
			"new sample rate": defaultSampleRate,
			"old sample rate": sampleRate,
		}).Info("Sample rate of the destination file will be changed")
	}

	if needsBitReduce {
		t.app.Logger().WithFields(logrus.Fields{
			"new bit depth": defaultBitDepth,
			"old bit depth": bitDepth,
		}).Info("Bit depth of the destination file will be changed")
	}

	ffmpegArgs := make([]string, 0)

	// Add sources
	ffmpegArgs = append(ffmpegArgs, "-i", sourcePath)

	if hasAlbumArt {
		ffmpegArgs = append(ffmpegArgs, "-i", albumArt)
	}

	// Map streams and set codecs
	if hasAlbumArt {
		ffmpegArgs = append(ffmpegArgs,
			"-map", "0:a", // Map audio from first input
			"-map", "1", // Map image from second input
			"-c:a", "alac", // ALAC codec for audio
			"-c:v", "copy", // Copy image without re-encoding
			"-disposition:v", "attached_pic",
		)
	} else {
		ffmpegArgs = append(ffmpegArgs,
			"-map", "0:a",
			"-c:a", "alac",
		)
	}

	// Handle downsampling
	if needsDownsample {
		ffmpegArgs = append(
			ffmpegArgs,
			"-af", "aresample=48000:resampler=soxr:precision=28",
		)
	} else {
		ffmpegArgs = append(ffmpegArgs, "-ar", fmt.Sprintf("%d", sampleRate))
	}

	if needsBitReduce {
		// Reduce to 16-bit with good dithering
		ffmpegArgs = append(ffmpegArgs,
			"-sample_fmt", "s16p",
			"-dither_method", "triangular",
		)
	}

	// Handle metadata copying and sort_artist filling
	ffmpegArgs = append(ffmpegArgs,
		"-map_metadata", "0",
		"-metadata", fmt.Sprintf("sort_artist=%s", t.escapeMetadata(sortArtist)),
		"-write_id3v2", "1",
		"-id3v2_version", "3",
		destinationPath,
		"-y",
		"-loglevel", "error",
		"-stats",
	)

	t.app.Logger().WithField(
		"ffmpeg command", "ffmpeg "+strings.Join(ffmpegArgs, " "),
	).Debug("FFMpeg parameters")

	ffmpeg := exec.Command("ffmpeg", ffmpegArgs...)
	var stderr bytes.Buffer
	ffmpeg.Stderr = &stderr

	if err := ffmpeg.Run(); err != nil {
		t.app.Logger().WithError(err).Error("Failed to invoke ffmpeg!")
		t.app.Logger().WithField("ffmpeg stderr", stderr.String()).Debug("Got ffmpeg stderr")

		return 0, fmt.Errorf("%w: %w (%w)", ErrTranscoder, ErrTranscodeError, err)
	}

	// Verify that the result file is saved to cache directory
	transcodedFileStat, err := os.Stat(destinationPath)
	if err != nil {
		t.app.Logger().WithError(err).WithFields(logrus.Fields{
			"source file": sourcePath,
			"destination": destinationPath,
		}).Error("Transcoded file not found (transcode error?). Check the logs for details")

		return 0, fmt.Errorf("%w: %w (%w)", ErrTranscoder, ErrTranscodedFileNotFound, err)
	}

	// Discard the file if it's less than 1 kilobyte: it's probably a transcode
	// error
	if transcodedFileStat.Size() < 1024 {
		t.app.Logger().WithFields(logrus.Fields{
			"source file":          sourcePath,
			"destination":          destinationPath,
			"transcoded file size": transcodedFileStat.Size(),
		}).Error("Transcoded file not found (transcode error?). Check the logs for details")

		return 0, fmt.Errorf(
			"%w: %w (%s)",
			ErrTranscoder, ErrTranscodedFileNotFound,
			fmt.Sprintf("size is %d bytes, less than 1 kilobyte", transcodedFileStat.Size()),
		)
	}

	t.app.Logger().WithFields(logrus.Fields{
		"source file":      sourcePath,
		"destination":      destinationPath,
		"destination size": transcodedFileStat.Size(),
	}).Info("File transcoded successfully")

	return transcodedFileStat.Size(), nil
}
