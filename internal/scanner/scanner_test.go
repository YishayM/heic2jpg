package scanner

import (
	"os"
	"path/filepath"
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

