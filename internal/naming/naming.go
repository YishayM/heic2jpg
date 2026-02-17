package naming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ApplyPattern applies a naming pattern to generate an output filename
// Supported patterns:
//   {name}  - original filename without extension
//   {date}  - file modification date in YYYY-MM-DD format
//   {index} - numbered sequence (001, 002, 003...)
func ApplyPattern(inputPath, pattern string, index int) (string, error) {
	// Get file info for metadata
	info, err := os.Stat(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Extract base name without extension
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// Get modification date
	modTime := info.ModTime()
	dateStr := modTime.Format("2006-01-02")

	// Apply pattern replacements
	result := pattern
	result = strings.ReplaceAll(result, "{name}", nameWithoutExt)
	result = strings.ReplaceAll(result, "{date}", dateStr)
	result = strings.ReplaceAll(result, "{index}", fmt.Sprintf("%03d", index))

	// Add .jpg extension
	return result + ".jpg", nil
}

// GenerateOutputPath creates the full output path given input path, pattern, index, and output directory
func GenerateOutputPath(inputPath, pattern string, index int, outputDir string) (string, error) {
	// Apply the naming pattern
	outputName, err := ApplyPattern(inputPath, pattern, index)
	if err != nil {
		return "", err
	}

	// Determine output directory
	var targetDir string
	if outputDir != "" {
		targetDir = outputDir
	} else {
		// Use same directory as input file
		targetDir = filepath.Dir(inputPath)
	}

	// Combine directory and filename
	return filepath.Join(targetDir, outputName), nil
}

