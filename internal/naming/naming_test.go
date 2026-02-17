package naming

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApplyPattern(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_image.heic")
	
	// Create the file with a known modification time
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()
	
	// Set a known modification time
	testTime := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(testFile, testTime, testTime); err != nil {
		t.Fatalf("Failed to set file time: %v", err)
	}

	tests := []struct {
		name     string
		pattern  string
		index    int
		expected string
	}{
		{
			name:     "name pattern",
			pattern:  "{name}",
			index:    1,
			expected: "test_image.jpg",
		},
		{
			name:     "date pattern",
			pattern:  "{date}",
			index:    1,
			expected: "2024-03-15.jpg",
		},
		{
			name:     "index pattern",
			pattern:  "{index}",
			index:    1,
			expected: "001.jpg",
		},
		{
			name:     "index pattern with larger number",
			pattern:  "{index}",
			index:    42,
			expected: "042.jpg",
		},
		{
			name:     "index pattern with three digits",
			pattern:  "{index}",
			index:    999,
			expected: "999.jpg",
		},
		{
			name:     "combined name and date",
			pattern:  "{name}_{date}",
			index:    1,
			expected: "test_image_2024-03-15.jpg",
		},
		{
			name:     "combined name and index",
			pattern:  "IMG_{index}",
			index:    5,
			expected: "IMG_005.jpg",
		},
		{
			name:     "all patterns combined",
			pattern:  "{name}_{date}_{index}",
			index:    10,
			expected: "test_image_2024-03-15_010.jpg",
		},
		{
			name:     "no pattern variables",
			pattern:  "output",
			index:    1,
			expected: "output.jpg",
		},
		{
			name:     "empty pattern",
			pattern:  "",
			index:    1,
			expected: ".jpg",
		},
		{
			name:     "custom prefix with index",
			pattern:  "photo_{index}",
			index:    123,
			expected: "photo_123.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyPattern(testFile, tt.pattern, tt.index)
			if err != nil {
				t.Fatalf("ApplyPattern() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("ApplyPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApplyPattern_FileNotFound(t *testing.T) {
	_, err := ApplyPattern("/nonexistent/file.heic", "{name}", 1)
	if err == nil {
		t.Error("ApplyPattern() expected error for nonexistent file, got nil")
	}
}

func TestGenerateOutputPath(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_image.heic")
	
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()

	tests := []struct {
		name      string
		inputPath string
		pattern   string
		index     int
		outputDir string
		wantName  string // just the filename part
		wantDir   string // expected directory
	}{
		{
			name:      "with output directory",
			inputPath: testFile,
			pattern:   "{name}",
			index:     1,
			outputDir: "/custom/output",
			wantName:  "test_image.jpg",
			wantDir:   "/custom/output",
		},
		{
			name:      "without output directory (same as input)",
			inputPath: testFile,
			pattern:   "{name}",
			index:     1,
			outputDir: "",
			wantName:  "test_image.jpg",
			wantDir:   tmpDir,
		},
		{
			name:      "with pattern and custom directory",
			inputPath: testFile,
			pattern:   "IMG_{index}",
			index:     5,
			outputDir: "/photos",
			wantName:  "IMG_005.jpg",
			wantDir:   "/photos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateOutputPath(tt.inputPath, tt.pattern, tt.index, tt.outputDir)
			if err != nil {
				t.Fatalf("GenerateOutputPath() error = %v", err)
			}
			
			gotDir := filepath.Dir(result)
			gotName := filepath.Base(result)
			
			if gotName != tt.wantName {
				t.Errorf("GenerateOutputPath() filename = %v, want %v", gotName, tt.wantName)
			}
			if gotDir != tt.wantDir {
				t.Errorf("GenerateOutputPath() directory = %v, want %v", gotDir, tt.wantDir)
			}
		})
	}
}

