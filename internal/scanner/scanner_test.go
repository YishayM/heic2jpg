package scanner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindHEICFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func(string) error
		wantCount int
		wantFiles []string // relative filenames
	}{
		{
			name: "finds .heic files",
			setup: func(dir string) error {
				files := []string{"image1.heic", "image2.heic"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantCount: 2,
			wantFiles: []string{"image1.heic", "image2.heic"},
		},
		{
			name: "finds .HEIC files (uppercase)",
			setup: func(dir string) error {
				files := []string{"IMAGE1.HEIC", "IMAGE2.HEIC"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantCount: 2,
			wantFiles: []string{"IMAGE1.HEIC", "IMAGE2.HEIC"},
		},
		{
			name: "finds mixed case .heic files",
			setup: func(dir string) error {
				files := []string{"image1.heic", "IMAGE2.HEIC", "Image3.Heic"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantCount: 3,
			wantFiles: []string{"image1.heic", "IMAGE2.HEIC", "Image3.Heic"},
		},
		{
			name: "ignores non-HEIC files",
			setup: func(dir string) error {
				files := []string{"image1.heic", "image2.jpg", "image3.png", "document.txt"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantCount: 1,
			wantFiles: []string{"image1.heic"},
		},
		{
			name: "handles empty directory",
			setup: func(dir string) error {
				return nil
			},
			wantCount: 0,
			wantFiles: []string{},
		},
		{
			name: "handles directory with no HEIC files",
			setup: func(dir string) error {
				files := []string{"image1.jpg", "image2.png", "document.pdf"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantCount: 0,
			wantFiles: []string{},
		},
		{
			name: "ignores subdirectories",
			setup: func(dir string) error {
				// Create a subdirectory
				subdir := filepath.Join(dir, "subdir")
				if err := os.Mkdir(subdir, 0755); err != nil {
					return err
				}
				// Create a HEIC file in the subdirectory (should be ignored)
				if err := os.WriteFile(filepath.Join(subdir, "image.heic"), []byte{}, 0644); err != nil {
					return err
				}
				// Create a HEIC file in the main directory
				if err := os.WriteFile(filepath.Join(dir, "main.heic"), []byte{}, 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 1,
			wantFiles: []string{"main.heic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test case
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.Mkdir(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Setup test files
			if err := tt.setup(testDir); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Run the function
			result, err := FindHEICFiles(testDir)
			if err != nil {
				t.Fatalf("FindHEICFiles() error = %v", err)
			}

			// Check count
			if len(result) != tt.wantCount {
				t.Errorf("FindHEICFiles() returned %d files, want %d", len(result), tt.wantCount)
			}

			// Check that all expected files are present
			resultMap := make(map[string]bool)
			for _, path := range result {
				resultMap[filepath.Base(path)] = true
			}

			for _, wantFile := range tt.wantFiles {
				if !resultMap[wantFile] {
					t.Errorf("FindHEICFiles() missing expected file %s", wantFile)
				}
			}
		})
	}
}

func TestFindHEICFiles_NonexistentDirectory(t *testing.T) {
	_, err := FindHEICFiles("/nonexistent/directory")
	if err == nil {
		t.Error("FindHEICFiles() expected error for nonexistent directory, got nil")
	}
}

func TestFindHEICFilesRecursive(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		recursive bool
		setup     func(string) error
		wantCount int
		wantFiles []string // relative filenames
	}{
		{
			name:      "non-recursive ignores nested directories",
			recursive: false,
			setup: func(dir string) error {
				// Create a nested structure
				subdir := filepath.Join(dir, "subdir")
				if err := os.Mkdir(subdir, 0755); err != nil {
					return err
				}
				// Create HEIC file in root
				if err := os.WriteFile(filepath.Join(dir, "root.heic"), []byte{}, 0644); err != nil {
					return err
				}
				// Create HEIC file in subdir
				if err := os.WriteFile(filepath.Join(subdir, "nested.heic"), []byte{}, 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 1,
			wantFiles: []string{"root.heic"},
		},
		{
			name:      "recursive finds nested HEIC files",
			recursive: true,
			setup: func(dir string) error {
				// Create a nested structure
				subdir := filepath.Join(dir, "subdir")
				deepdir := filepath.Join(subdir, "deepdir")
				if err := os.MkdirAll(deepdir, 0755); err != nil {
					return err
				}
				// Create HEIC files at various levels
				if err := os.WriteFile(filepath.Join(dir, "root.heic"), []byte{}, 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(subdir, "nested.heic"), []byte{}, 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(deepdir, "deep.heic"), []byte{}, 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 3,
			wantFiles: []string{"root.heic", "nested.heic", "deep.heic"},
		},
		{
			name:      "recursive handles mixed extensions",
			recursive: true,
			setup: func(dir string) error {
				subdir := filepath.Join(dir, "photos")
				if err := os.Mkdir(subdir, 0755); err != nil {
					return err
				}
				// Create HEIC and HEIF files
				if err := os.WriteFile(filepath.Join(dir, "image.heic"), []byte{}, 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(subdir, "photo.HEIF"), []byte{}, 0644); err != nil {
					return err
				}
				// Create non-matching files
				if err := os.WriteFile(filepath.Join(subdir, "doc.txt"), []byte{}, 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 2,
			wantFiles: []string{"image.heic", "photo.HEIF"},
		},
		{
			name:      "recursive with empty subdirectories",
			recursive: true,
			setup: func(dir string) error {
				// Create empty subdirectory
				subdir := filepath.Join(dir, "empty")
				if err := os.Mkdir(subdir, 0755); err != nil {
					return err
				}
				// Create HEIC file in root only
				if err := os.WriteFile(filepath.Join(dir, "only.heic"), []byte{}, 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 1,
			wantFiles: []string{"only.heic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test case
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.Mkdir(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Setup test files
			if err := tt.setup(testDir); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Run the function
			result, err := FindHEICFilesRecursive(testDir, tt.recursive)
			if err != nil {
				t.Fatalf("FindHEICFilesRecursive() error = %v", err)
			}

			// Check count
			if len(result) != tt.wantCount {
				t.Errorf("FindHEICFilesRecursive() returned %d files, want %d", len(result), tt.wantCount)
			}

			// Check that all expected files are present
			resultMap := make(map[string]bool)
			for _, path := range result {
				resultMap[filepath.Base(path)] = true
			}

			for _, wantFile := range tt.wantFiles {
				if !resultMap[wantFile] {
					t.Errorf("FindHEICFilesRecursive() missing expected file %s", wantFile)
				}
			}
		})
	}
}

func TestFindHEICFiles_SkipsSymlinks(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a real HEIC file
	realFile := filepath.Join(tmpDir, "real.heic")
	if err := os.WriteFile(realFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a symlink to the real file
	symlinkFile := filepath.Join(tmpDir, "symlink.heic")
	if err := os.Symlink(realFile, symlinkFile); err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Capture stderr to check for warning message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the function
	result, err := FindHEICFiles(tmpDir)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	if err != nil {
		t.Fatalf("FindHEICFiles() error = %v", err)
	}

	// Should only find the real file, not the symlink
	if len(result) != 1 {
		t.Errorf("FindHEICFiles() returned %d files, want 1", len(result))
	}

	// Check that the real file is in the results
	found := false
	for _, path := range result {
		if filepath.Base(path) == "real.heic" {
			found = true
		}
		if filepath.Base(path) == "symlink.heic" {
			t.Errorf("FindHEICFiles() should not include symlink")
		}
	}
	if !found {
		t.Errorf("FindHEICFiles() missing real.heic")
	}

	// Check that warning was logged
	if !strings.Contains(stderrOutput, "Skipping symlink:") {
		t.Errorf("Expected warning about skipping symlink, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "symlink.heic") {
		t.Errorf("Expected warning to mention symlink.heic, got: %s", stderrOutput)
	}
}

func TestFindHEICFilesRecursive_SkipsSymlinkDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a subdirectory with a HEIC file
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	subdirFile := filepath.Join(subdir, "nested.heic")
	if err := os.WriteFile(subdirFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Create a symlink to the parent directory (potential infinite loop)
	symlinkDir := filepath.Join(subdir, "loop")
	if err := os.Symlink(tmpDir, symlinkDir); err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Create a file in the root
	rootFile := filepath.Join(tmpDir, "root.heic")
	if err := os.WriteFile(rootFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	// Capture stderr to check for warning message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the function with recursive=true
	result, err := FindHEICFilesRecursive(tmpDir, true)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	if err != nil {
		t.Fatalf("FindHEICFilesRecursive() error = %v", err)
	}

	// Should find exactly 2 files (root.heic and nested.heic), not infinite loop
	if len(result) != 2 {
		t.Errorf("FindHEICFilesRecursive() returned %d files, want 2", len(result))
	}

	// Check that warning was logged
	if !strings.Contains(stderrOutput, "Skipping symlink:") {
		t.Errorf("Expected warning about skipping symlink, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "loop") {
		t.Errorf("Expected warning to mention loop symlink, got: %s", stderrOutput)
	}
}

