package converter

import (
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jdeng/goheif"
)

// ConvertResult contains the result of a conversion operation
type ConvertResult struct {
	HasEXIF bool // Whether EXIF data was found and preserved
}

// DiskFullError represents an error caused by insufficient disk space
type DiskFullError struct {
	OutputPath string
}

func (e *DiskFullError) Error() string {
	return fmt.Sprintf("Disk full - cannot write %s", e.OutputPath)
}

// humanizeFileError converts low-level file errors into user-friendly messages
func humanizeFileError(err error, path string) error {
	if err == nil {
		return nil
	}

	// Check for common file system errors
	if os.IsNotExist(err) {
		return fmt.Errorf("File not found: %s", path)
	}
	if os.IsPermission(err) {
		return fmt.Errorf("Permission denied: %s", path)
	}

	// Return original error wrapped with context
	return fmt.Errorf("failed to access %s: %w", filepath.Base(path), err)
}

// isNonHEICError checks if an error indicates the file is not a valid HEIC/HEIF image
func isNonHEICError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// Common error patterns from goheif when file is not HEIC
	return strings.Contains(errMsg, "box type") ||
		strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "unsupported")
}

// CheckEXIF checks if a HEIC file has EXIF data without converting it
func CheckEXIF(inputPath string) (bool, error) {
	fi, err := os.Open(inputPath)
	if err != nil {
		return false, humanizeFileError(err, inputPath)
	}
	defer fi.Close()

	exif, err := goheif.ExtractExif(fi)
	if err != nil {
		return false, nil // No EXIF, but not an error
	}
	return len(exif) > 0, nil
}

// Convert converts a HEIC file to JPEG format, preserving EXIF data
// Uses atomic file writes to prevent corrupted output files on crash/interruption
// maxSizeBytes: maximum file size to process (0 = no limit)
// Returns ConvertResult indicating whether EXIF data was preserved
func Convert(inputPath, outputPath string, maxSizeBytes int64) (*ConvertResult, error) {
	result := &ConvertResult{HasEXIF: false}

	// Check file size before opening (prevent OOM on huge files)
	if maxSizeBytes > 0 {
		fileInfo, err := os.Stat(inputPath)
		if err != nil {
			return nil, humanizeFileError(err, inputPath)
		}

		if fileInfo.Size() > maxSizeBytes {
			return nil, fmt.Errorf("file too large: %s (limit: %s, use --max-size to override)",
				formatSize(fileInfo.Size()),
				formatSize(maxSizeBytes))
		}
	}

	// Open input file
	fi, err := os.Open(inputPath)
	if err != nil {
		return nil, humanizeFileError(err, inputPath)
	}
	defer fi.Close()

	// Extract EXIF data
	exif, err := goheif.ExtractExif(fi)
	if err != nil {
		// Don't warn if this looks like a non-HEIC file (we'll catch it during decode)
		if !isNonHEICError(err) {
			// EXIF is optional, just log a warning for valid HEIC files
			fmt.Fprintf(os.Stderr, "Warning: no EXIF data found in %s: %v\n", inputPath, err)
		}
	} else if len(exif) > 0 {
		result.HasEXIF = true
	}

	// Reset file position to start after EXIF extraction
	// ExtractExif reads the file and leaves the cursor at an unknown position
	_, err = fi.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to start of file: %w", err)
	}

	// Decode HEIC image
	img, err := goheif.Decode(fi)
	if err != nil {
		// Check if this is a non-HEIC file
		if isNonHEICError(err) {
			return nil, fmt.Errorf("%s is not a HEIC/HEIF image", filepath.Base(inputPath))
		}
		return nil, fmt.Errorf("failed to decode HEIC image: %w", err)
	}

	// Create temp file for atomic write
	// Use .tmp suffix in same directory as final output
	tempPath := outputPath + ".tmp"
	fo, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", wrapDiskFullError(err, outputPath))
	}

	// Ensure temp file is cleaned up on error or panic
	var success bool
	defer func() {
		fo.Close()
		if !success {
			// Remove temp file if we didn't succeed
			os.Remove(tempPath)
		}
	}()

	// Create writer with EXIF support
	w, err := newWriterExif(fo, exif)
	if err != nil {
		return nil, fmt.Errorf("failed to create EXIF writer: %w", wrapDiskFullError(err, outputPath))
	}

	// Encode to JPEG
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", wrapDiskFullError(err, outputPath))
	}

	// Close the file before renaming (required on Windows)
	if err := fo.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", wrapDiskFullError(err, outputPath))
	}

	// Atomically rename temp file to final output path
	if err := os.Rename(tempPath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to rename temp file to output: %w", wrapDiskFullError(err, outputPath))
	}

	// Mark success so defer doesn't delete the file
	success = true
	return result, nil
}

// writerSkipper is a writer that skips the first n bytes
type writerSkipper struct {
	w     io.Writer
	skip  int
	wrote int
}

func (w *writerSkipper) Write(p []byte) (n int, err error) {
	if w.wrote >= w.skip {
		return w.w.Write(p)
	}

	if w.wrote+len(p) <= w.skip {
		w.wrote += len(p)
		return len(p), nil
	}

	skipRemaining := w.skip - w.wrote
	w.wrote += len(p)
	_, err = w.w.Write(p[skipRemaining:])
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// newWriterExif creates a writer that embeds EXIF data into JPEG
func newWriterExif(w io.Writer, exif []byte) (io.Writer, error) {
	if len(exif) == 0 {
		return w, nil
	}

	// Write JPEG SOI marker
	_, err := w.Write([]byte{0xff, 0xd8})
	if err != nil {
		return nil, err
	}

	// Write APP1 marker for EXIF
	app1Marker := []byte{0xff, 0xe1}
	_, err = w.Write(app1Marker)
	if err != nil {
		return nil, err
	}

	// Write EXIF segment length (including length bytes but not marker)
	exifLen := len(exif) + 2
	_, err = w.Write([]byte{byte(exifLen >> 8), byte(exifLen & 0xff)})
	if err != nil {
		return nil, err
	}

	// Write EXIF data
	_, err = w.Write(exif)
	if err != nil {
		return nil, err
	}

	// Return a writer that skips the SOI marker that jpeg.Encode will write
	return &writerSkipper{w: w, skip: 2}, nil
}

// formatSize formats a byte size into a human-readable string
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes >= GB {
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	} else if bytes >= MB {
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	} else if bytes >= KB {
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%dB", bytes)
}

// isDiskFullError checks if an error is caused by insufficient disk space
func isDiskFullError(err error) bool {
	return errors.Is(err, syscall.ENOSPC)
}

// wrapDiskFullError wraps an error with a user-friendly message if it's a disk full error
func wrapDiskFullError(err error, outputPath string) error {
	if isDiskFullError(err) {
		return &DiskFullError{OutputPath: outputPath}
	}
	return err
}

// IsDiskFullError checks if an error is a DiskFullError
func IsDiskFullError(err error) bool {
	var diskFullErr *DiskFullError
	return errors.As(err, &diskFullErr)
}
