package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckPathCollision_SamePath(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with identical paths
	err := checkPathCollision(testFile, testFile)
	if err == nil {
		t.Error("checkPathCollision() expected error for identical paths, got nil")
	}
	if err != nil && err.Error() != "Error: input and output are the same file" {
		t.Errorf("checkPathCollision() error = %v, want 'Error: input and output are the same file'", err)
	}
}

func TestCheckPathCollision_DifferentPaths(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.heic")
	outputFile := filepath.Join(tmpDir, "output.jpg")
	
	if err := os.WriteFile(inputFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with different paths
	err := checkPathCollision(inputFile, outputFile)
	if err != nil {
		t.Errorf("checkPathCollision() unexpected error for different paths: %v", err)
	}
}

func TestCheckPathCollision_RelativeVsAbsolute(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test with relative path vs absolute path to same file
	err = checkPathCollision("test.heic", testFile)
	if err == nil {
		t.Error("checkPathCollision() expected error for relative vs absolute path to same file, got nil")
	}
	if err != nil && err.Error() != "Error: input and output are the same file" {
		t.Errorf("checkPathCollision() error = %v, want 'Error: input and output are the same file'", err)
	}
}

func TestCheckPathCollision_Symlinks(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	symlinkFile := filepath.Join(tmpDir, "link.heic")
	
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create symlink
	if err := os.Symlink(testFile, symlinkFile); err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	// Test with symlink pointing to same file
	err := checkPathCollision(testFile, symlinkFile)
	if err == nil {
		t.Error("checkPathCollision() expected error for symlink to same file, got nil")
	}
	if err != nil && err.Error() != "Error: input and output are the same file" {
		t.Errorf("checkPathCollision() error = %v, want 'Error: input and output are the same file'", err)
	}
}

func TestCheckPathCollision_NonexistentOutput(t *testing.T) {
	// Create a temporary input file
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.heic")
	outputFile := filepath.Join(tmpDir, "output.jpg")
	
	if err := os.WriteFile(inputFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with nonexistent output file (should not error)
	err := checkPathCollision(inputFile, outputFile)
	if err != nil {
		t.Errorf("checkPathCollision() unexpected error for nonexistent output: %v", err)
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input       string
		expected    int64
		expectError bool
	}{
		// Valid cases
		{"0", 0, false},
		{"100", 100, false},
		{"1KB", 1024, false},
		{"1kb", 1024, false},
		{"500MB", 500 * 1024 * 1024, false},
		{"500mb", 500 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"1gb", 1024 * 1024 * 1024, false},
		{"2GB", 2 * 1024 * 1024 * 1024, false},
		{"10GB", 10 * 1024 * 1024 * 1024, false},
		{"100B", 100, false},
		{"  500MB  ", 500 * 1024 * 1024, false}, // with spaces

		// Invalid cases
		{"", 0, true},
		{"abc", 0, true},
		{"500XB", 0, true},
		{"-100MB", 0, true},
		{"1.5GB", 0, true}, // no decimal support
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseSize(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseSize(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseSize(%q) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}
