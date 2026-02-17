package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	binaryName = "heic2jpg"
	testHEIC   = "testrun/camel.heic"
)

// TestMain builds the binary before running tests
func TestMain(m *testing.M) {
	// Build the binary
	build := exec.Command("go", "build", "-o", binaryName, "./cmd/heic2jpg")
	if err := build.Run(); err != nil {
		os.Stderr.WriteString("Failed to build binary: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove(binaryName)
	os.Exit(code)
}

func TestE2E_SingleFileConversion(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Copy test HEIC file to temp dir
	testFile := filepath.Join(tmpDir, "test.heic")
	copyFile(t, testHEIC, testFile)
	
	// Run conversion
	cmd := exec.Command("./"+binaryName, testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify JPG was created
	jpgFile := filepath.Join(tmpDir, "test.jpg")
	if !fileExists(jpgFile) {
		t.Errorf("Expected output file %s does not exist", jpgFile)
	}
	
	// Verify it's a valid JPEG
	if !isValidJPEG(t, jpgFile) {
		t.Errorf("Output file is not a valid JPEG")
	}
}

func TestE2E_DirectoryConversion(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Copy multiple HEIC files
	copyFile(t, testHEIC, filepath.Join(tmpDir, "file1.heic"))
	copyFile(t, testHEIC, filepath.Join(tmpDir, "file2.heic"))
	
	// Run conversion on directory
	cmd := exec.Command("./"+binaryName, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify both JPGs were created
	if !fileExists(filepath.Join(tmpDir, "file1.jpg")) {
		t.Errorf("Expected file1.jpg does not exist")
	}
	if !fileExists(filepath.Join(tmpDir, "file2.jpg")) {
		t.Errorf("Expected file2.jpg does not exist")
	}
}

func TestE2E_NamingPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "photo.heic")
	copyFile(t, testHEIC, testFile)
	
	// Run with naming pattern
	cmd := exec.Command("./"+binaryName, "--pattern", "{name}_web", testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify output file with pattern
	expectedFile := filepath.Join(tmpDir, "photo_web.jpg")
	if !fileExists(expectedFile) {
		t.Errorf("Expected output file %s does not exist", expectedFile)
	}
}

func TestE2E_OutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	copyFile(t, testHEIC, testFile)
	
	outDir := filepath.Join(tmpDir, "output")
	
	// Run with output directory
	cmd := exec.Command("./"+binaryName, "--output", outDir, testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify output in specified directory
	expectedFile := filepath.Join(outDir, "test.jpg")
	if !fileExists(expectedFile) {
		t.Errorf("Expected output file %s does not exist", expectedFile)
	}
}

func TestE2E_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	copyFile(t, testHEIC, testFile)
	
	// Run with dry-run flag
	cmd := exec.Command("./"+binaryName, "--dry-run", testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}
	
	// Verify NO JPG was created
	jpgFile := filepath.Join(tmpDir, "test.jpg")
	if fileExists(jpgFile) {
		t.Errorf("JPG file should not exist in dry-run mode")
	}
	
	// Verify output mentions dry-run
	outputStr := string(output)
	if !strings.Contains(outputStr, "dry-run") {
		t.Errorf("Output should mention dry-run, got: %s", outputStr)
	}
}

func TestE2E_VersionFlag(t *testing.T) {
	// Run with --version flag
	cmd := exec.Command("./"+binaryName, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains version
	outputStr := string(output)
	if !strings.Contains(outputStr, "1.0.0") {
		t.Errorf("Output should contain version 1.0.0, got: %s", outputStr)
	}
}

func TestE2E_HelpFlag(t *testing.T) {
	// Run with --help flag
	cmd := exec.Command("./"+binaryName, "--help")
	output, _ := cmd.CombinedOutput()
	// Note: --help may exit with 0 or non-zero, both are acceptable

	// Verify output contains usage information
	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("Output should contain 'Usage:', got: %s", outputStr)
	}
}

func TestE2E_NoHEICFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Run on empty directory
	cmd := exec.Command("./"+binaryName, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify output mentions no HEIC files found
	outputStr := string(output)
	if !strings.Contains(outputStr, "No HEIC files found") {
		t.Errorf("Output should mention 'No HEIC files found', got: %s", outputStr)
	}
}

func TestE2E_InvalidFile(t *testing.T) {
	// Run with non-existent file
	cmd := exec.Command("./"+binaryName, "nonexistent.heic")
	output, err := cmd.CombinedOutput()

	// Should exit with error
	if err == nil {
		t.Errorf("Command should fail for non-existent file")
	}

	// Verify error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "Error") && !strings.Contains(outputStr, "error") {
		t.Errorf("Output should contain error message, got: %s", outputStr)
	}
}

// Helper functions

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isValidJPEG(t *testing.T, path string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	// Check JPEG magic bytes (FF D8 FF)
	if len(data) < 3 {
		return false
	}
	return data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
}

