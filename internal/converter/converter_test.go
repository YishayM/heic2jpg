package converter

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriterSkipper(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		input    []byte
		expected []byte
	}{
		{
			name:     "skip 2 bytes",
			skip:     2,
			input:    []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			expected: []byte{0x03, 0x04, 0x05},
		},
		{
			name:     "skip 0 bytes",
			skip:     0,
			input:    []byte{0x01, 0x02, 0x03},
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "skip all bytes",
			skip:     5,
			input:    []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			expected: []byte{},
		},
		{
			name:     "skip more than input",
			skip:     10,
			input:    []byte{0x01, 0x02, 0x03},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ws := &writerSkipper{
				w:    &buf,
				skip: tt.skip,
			}

			_, err := ws.Write(tt.input)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			result := buf.Bytes()
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Write() result = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWriterSkipper_MultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	ws := &writerSkipper{
		w:    &buf,
		skip: 3,
	}

	// First write: 2 bytes (all skipped)
	_, err := ws.Write([]byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("First Write() error = %v", err)
	}

	// Second write: 3 bytes (1 skipped, 2 written)
	_, err = ws.Write([]byte{0x03, 0x04, 0x05})
	if err != nil {
		t.Fatalf("Second Write() error = %v", err)
	}

	// Third write: 2 bytes (all written)
	_, err = ws.Write([]byte{0x06, 0x07})
	if err != nil {
		t.Fatalf("Third Write() error = %v", err)
	}

	expected := []byte{0x04, 0x05, 0x06, 0x07}
	result := buf.Bytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Final result = %v, want %v", result, expected)
	}
}

func TestNewWriterExif_NoExif(t *testing.T) {
	var buf bytes.Buffer
	w, err := newWriterExif(&buf, []byte{})
	if err != nil {
		t.Fatalf("newWriterExif() error = %v", err)
	}

	// Should return the original writer unchanged
	if w != &buf {
		t.Error("newWriterExif() should return original writer when no EXIF data")
	}

	// Buffer should be empty
	if buf.Len() != 0 {
		t.Errorf("Buffer should be empty, got %d bytes", buf.Len())
	}
}

func TestNewWriterExif_WithExif(t *testing.T) {
	var buf bytes.Buffer
	exifData := []byte{0x45, 0x78, 0x69, 0x66, 0x00, 0x00} // "Exif\0\0"

	w, err := newWriterExif(&buf, exifData)
	if err != nil {
		t.Fatalf("newWriterExif() error = %v", err)
	}

	// Should return a writerSkipper
	if _, ok := w.(*writerSkipper); !ok {
		t.Error("newWriterExif() should return a writerSkipper when EXIF data is present")
	}

	// Check that JPEG SOI marker was written
	result := buf.Bytes()
	if len(result) < 2 || result[0] != 0xff || result[1] != 0xd8 {
		t.Errorf("Expected JPEG SOI marker (0xff 0xd8) at start, got %v", result[:2])
	}

	// Check that APP1 marker was written
	if len(result) < 4 || result[2] != 0xff || result[3] != 0xe1 {
		t.Errorf("Expected APP1 marker (0xff 0xe1) after SOI, got %v", result[2:4])
	}

	// Check EXIF segment length
	exifLen := len(exifData) + 2
	expectedLenHigh := byte(exifLen >> 8)
	expectedLenLow := byte(exifLen & 0xff)
	if len(result) < 6 || result[4] != expectedLenHigh || result[5] != expectedLenLow {
		t.Errorf("Expected EXIF length bytes (%d %d), got %v", expectedLenHigh, expectedLenLow, result[4:6])
	}

	// Check that EXIF data was written
	if len(result) < 6+len(exifData) {
		t.Fatalf("Buffer too short, expected at least %d bytes, got %d", 6+len(exifData), len(result))
	}
	actualExif := result[6 : 6+len(exifData)]
	if !bytes.Equal(actualExif, exifData) {
		t.Errorf("EXIF data mismatch, got %v, want %v", actualExif, exifData)
	}
}

func TestConvert_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.jpg")

	_, err := Convert("/nonexistent/file.heic", outputPath, 0)
	if err == nil {
		t.Error("Convert() expected error for nonexistent file, got nil")
	}
}

func TestConvert_InvalidOutputPath(t *testing.T) {
	// Create a temporary input file (even though it's not a valid HEIC)
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.heic")
	if err := os.WriteFile(inputPath, []byte("not a heic file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to write to an invalid output path
	_, err := Convert(inputPath, "/nonexistent/directory/output.jpg", 0)
	if err == nil {
		t.Error("Convert() expected error for invalid output path, got nil")
	}
}

func TestConvert_AtomicWrite_NoOrphanedTempFiles(t *testing.T) {
	// This test verifies that temp files are cleaned up on error
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.heic")
	outputPath := filepath.Join(tmpDir, "output.jpg")
	tempPath := outputPath + ".tmp"

	// Create an invalid HEIC file that will fail during decode
	if err := os.WriteFile(inputPath, []byte("not a valid heic file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Attempt conversion (should fail during decode)
	_, err := Convert(inputPath, outputPath, 0)
	if err == nil {
		t.Error("Convert() expected error for invalid HEIC file, got nil")
	}

	// Verify no temp file was left behind
	if _, err := os.Stat(tempPath); err == nil {
		t.Error("Temp file should have been cleaned up on error")
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking temp file: %v", err)
	}

	// Verify no output file was created
	if _, err := os.Stat(outputPath); err == nil {
		t.Error("Output file should not exist when conversion fails")
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking output file: %v", err)
	}
}

func TestConvert_FileSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "large.heic")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	// Create a file larger than 100 bytes
	largeContent := make([]byte, 200)
	if err := os.WriteFile(inputPath, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with size limit of 100 bytes (should fail)
	_, err := Convert(inputPath, outputPath, 100)
	if err == nil {
		t.Error("Convert() expected error for file exceeding size limit, got nil")
	}
	if err != nil && !contains(err.Error(), "file too large") {
		t.Errorf("Expected 'file too large' error, got: %v", err)
	}

	// Test with size limit of 0 (disabled, should still fail but for different reason - invalid HEIC)
	_, err = Convert(inputPath, outputPath, 0)
	if err == nil {
		t.Error("Convert() expected error for invalid HEIC file, got nil")
	}
	// Should fail during decode, not size check
	if err != nil && contains(err.Error(), "file too large") {
		t.Errorf("Should not fail with size limit when limit is 0, got: %v", err)
	}

	// Test with size limit larger than file (should fail during decode, not size check)
	_, err = Convert(inputPath, outputPath, 1000)
	if err == nil {
		t.Error("Convert() expected error for invalid HEIC file, got nil")
	}
	if err != nil && contains(err.Error(), "file too large") {
		t.Errorf("Should not fail with size limit when file is smaller, got: %v", err)
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{100, "100B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{1024*1024 + 512*1024, "1.5MB"},
		{500 * 1024 * 1024, "500.0MB"},
		{1024 * 1024 * 1024, "1.0GB"},
		{2*1024*1024*1024 + 512*1024*1024, "2.5GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestIsDiskFullError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ENOSPC error",
			err:      syscall.ENOSPC,
			expected: true,
		},
		{
			name:     "wrapped ENOSPC error",
			err:      errors.New("write failed: " + syscall.ENOSPC.Error()),
			expected: false, // String matching doesn't work with errors.Is
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDiskFullError(tt.err)
			if result != tt.expected {
				t.Errorf("isDiskFullError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestWrapDiskFullError(t *testing.T) {
	outputPath := "/tmp/output.jpg"

	tests := []struct {
		name           string
		err            error
		expectDiskFull bool
	}{
		{
			name:           "ENOSPC error",
			err:            syscall.ENOSPC,
			expectDiskFull: true,
		},
		{
			name:           "other error",
			err:            errors.New("some other error"),
			expectDiskFull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := wrapDiskFullError(tt.err, outputPath)

			if tt.expectDiskFull {
				var diskFullErr *DiskFullError
				if !errors.As(wrapped, &diskFullErr) {
					t.Errorf("wrapDiskFullError() should return DiskFullError for ENOSPC, got %T", wrapped)
				}
				if diskFullErr != nil && diskFullErr.OutputPath != outputPath {
					t.Errorf("DiskFullError.OutputPath = %s, want %s", diskFullErr.OutputPath, outputPath)
				}
				expectedMsg := "Disk full - cannot write " + outputPath
				if wrapped.Error() != expectedMsg {
					t.Errorf("DiskFullError.Error() = %s, want %s", wrapped.Error(), expectedMsg)
				}
			} else {
				if wrapped != tt.err {
					t.Errorf("wrapDiskFullError() should return original error for non-ENOSPC errors")
				}
			}
		})
	}
}

func TestIsDiskFullErrorFunc(t *testing.T) {
	outputPath := "/tmp/output.jpg"

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "DiskFullError",
			err:      &DiskFullError{OutputPath: outputPath},
			expected: true,
		},
		{
			name:     "wrapped DiskFullError",
			err:      errors.New("failed: " + (&DiskFullError{OutputPath: outputPath}).Error()),
			expected: false, // String wrapping doesn't preserve type
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDiskFullError(tt.err)
			if result != tt.expected {
				t.Errorf("IsDiskFullError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
