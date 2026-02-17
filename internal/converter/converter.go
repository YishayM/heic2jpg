package converter

import (
	"fmt"
	"image/jpeg"
	"io"
	"os"

	"github.com/jdeng/goheif"
)

// Convert converts a HEIC file to JPEG format, preserving EXIF data
// Uses atomic file writes to prevent corrupted output files on crash/interruption
func Convert(inputPath, outputPath string) error {
	// Open input file
	fi, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer fi.Close()

	// Extract EXIF data
	exif, err := goheif.ExtractExif(fi)
	if err != nil {
		// EXIF is optional, just log a warning
		fmt.Fprintf(os.Stderr, "Warning: no EXIF data found in %s: %v\n", inputPath, err)
	}

	// Decode HEIC image
	img, err := goheif.Decode(fi)
	if err != nil {
		return fmt.Errorf("failed to decode HEIC image: %w", err)
	}

	// Create temp file for atomic write
	// Use .tmp suffix in same directory as final output
	tempPath := outputPath + ".tmp"
	fo, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
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
		return fmt.Errorf("failed to create EXIF writer: %w", err)
	}

	// Encode to JPEG
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: 95})
	if err != nil {
		return fmt.Errorf("failed to encode JPEG: %w", err)
	}

	// Close the file before renaming (required on Windows)
	if err := fo.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically rename temp file to final output path
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to rename temp file to output: %w", err)
	}

	// Mark success so defer doesn't delete the file
	success = true
	return nil
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

