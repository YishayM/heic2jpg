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

func TestE2E_OverwriteProtection_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	jpgFile := filepath.Join(tmpDir, "test.jpg")

	// Copy test HEIC file
	copyFile(t, testHEIC, testFile)

	// First conversion - should succeed
	cmd := exec.Command("./"+binaryName, testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("First conversion failed: %v\nOutput: %s", err, output)
	}

	// Verify JPG was created
	if !fileExists(jpgFile) {
		t.Fatalf("Expected output file %s does not exist", jpgFile)
	}

	// Get original file info
	originalInfo, err := os.Stat(jpgFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	// Second conversion without --force - should skip
	cmd = exec.Command("./"+binaryName, testFile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Second conversion failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Output file exists") {
		t.Errorf("Output should mention file exists, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "use --force") {
		t.Errorf("Output should mention --force flag, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "1 skipped") {
		t.Errorf("Output should mention 1 skipped, got: %s", outputStr)
	}

	// Verify file was NOT overwritten (same modification time)
	newInfo, err := os.Stat(jpgFile)
	if err != nil {
		t.Fatalf("Failed to stat output file after second run: %v", err)
	}
	if !newInfo.ModTime().Equal(originalInfo.ModTime()) {
		t.Errorf("File was modified when it should have been skipped")
	}
}

func TestE2E_OverwriteProtection_WithForce(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.heic")
	jpgFile := filepath.Join(tmpDir, "test.jpg")

	// Copy test HEIC file
	copyFile(t, testHEIC, testFile)

	// First conversion
	cmd := exec.Command("./"+binaryName, testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("First conversion failed: %v\nOutput: %s", err, output)
	}

	// Verify JPG was created
	if !fileExists(jpgFile) {
		t.Fatalf("Expected output file %s does not exist", jpgFile)
	}

	// Get original file info
	originalInfo, err := os.Stat(jpgFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	// Second conversion with --force - should overwrite
	cmd = exec.Command("./"+binaryName, "--force", testFile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Second conversion with --force failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "skipped") {
		t.Errorf("Output should not mention skipped with --force, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Converted 1 file") {
		t.Errorf("Output should mention converted 1 file, got: %s", outputStr)
	}

	// Verify file WAS overwritten (different modification time)
	newInfo, err := os.Stat(jpgFile)
	if err != nil {
		t.Fatalf("Failed to stat output file after second run: %v", err)
	}
	if newInfo.ModTime().Equal(originalInfo.ModTime()) {
		t.Errorf("File was not modified when it should have been overwritten with --force")
	}
}

func TestE2E_OverwriteProtection_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Copy multiple HEIC files
	copyFile(t, testHEIC, filepath.Join(tmpDir, "file1.heic"))
	copyFile(t, testHEIC, filepath.Join(tmpDir, "file2.heic"))
	copyFile(t, testHEIC, filepath.Join(tmpDir, "file3.heic"))

	// First conversion - should convert all 3
	cmd := exec.Command("./"+binaryName, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("First conversion failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Converted 3 file") {
		t.Errorf("Output should mention converted 3 files, got: %s", outputStr)
	}

	// Second conversion without --force - should skip all 3
	cmd = exec.Command("./"+binaryName, tmpDir)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Second conversion failed: %v\nOutput: %s", err, output)
	}

	outputStr = string(output)
	if !strings.Contains(outputStr, "3 skipped") {
		t.Errorf("Output should mention 3 skipped, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "use --force") {
		t.Errorf("Output should mention --force flag, got: %s", outputStr)
	}

	// Third conversion with --force - should convert all 3
	cmd = exec.Command("./"+binaryName, "--force", tmpDir)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Third conversion with --force failed: %v\nOutput: %s", err, output)
	}

	outputStr = string(output)
	if !strings.Contains(outputStr, "Converted 3 file") {
		t.Errorf("Output should mention converted 3 files with --force, got: %s", outputStr)
	}
	if strings.Contains(outputStr, "skipped") {
		t.Errorf("Output should not mention skipped with --force, got: %s", outputStr)
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

