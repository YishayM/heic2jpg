package naming

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// calculateIndexWidth determines the number of digits needed to represent totalCount
// Returns at least 3 for consistency with existing behavior
func calculateIndexWidth(totalCount int) int {
	if totalCount <= 0 {
		return 3 // Default minimum width
	}
	// Calculate number of digits needed
	width := int(math.Log10(float64(totalCount))) + 1
	// Ensure minimum width of 3
	if width < 3 {
		return 3
	}
	return width
}

// ApplyPattern applies a naming pattern to generate an output filename
// Supported patterns:
//   {name}  - original filename without extension
//   {date}  - file modification date in YYYY-MM-DD format
//   {index} - numbered sequence (001, 002, 003...) - width adapts to totalCount
// totalCount: total number of files being processed (used to calculate index width)
func ApplyPattern(inputPath, pattern string, index int, totalCount int) (string, error) {
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

	// Calculate dynamic index width based on total count
	indexWidth := calculateIndexWidth(totalCount)
	indexFormat := fmt.Sprintf("%%0%dd", indexWidth)

	// Apply pattern replacements
	result := pattern
	result = strings.ReplaceAll(result, "{name}", nameWithoutExt)
	result = strings.ReplaceAll(result, "{date}", dateStr)
	result = strings.ReplaceAll(result, "{index}", fmt.Sprintf(indexFormat, index))

	// Add .jpg extension
	return result + ".jpg", nil
}

// GenerateOutputPath creates the full output path given input path, pattern, index, and output directory
// totalCount: total number of files being processed (used to calculate index width)
func GenerateOutputPath(inputPath, pattern string, index int, outputDir string, totalCount int) (string, error) {
	// Apply the naming pattern
	outputName, err := ApplyPattern(inputPath, pattern, index, totalCount)
	if err != nil {
		return "", err
	}

	// Security: Validate the output name before joining with directory
	// Check for absolute paths or Windows drive letters
	if filepath.IsAbs(outputName) {
		return "", fmt.Errorf("invalid pattern: output path would escape target directory")
	}

	// Check for Windows-style paths (C:\, D:\, etc.) or UNC paths (\\server\share)
	if len(outputName) >= 2 {
		// Windows drive letter (C:, D:, etc.)
		if outputName[1] == ':' && ((outputName[0] >= 'A' && outputName[0] <= 'Z') || (outputName[0] >= 'a' && outputName[0] <= 'z')) {
			return "", fmt.Errorf("invalid pattern: output path would escape target directory")
		}
		// Windows UNC path (\\server\share)
		if len(outputName) >= 2 && outputName[0] == '\\' && outputName[1] == '\\' {
			return "", fmt.Errorf("invalid pattern: output path would escape target directory")
		}
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
	outputPath := filepath.Join(targetDir, outputName)

	// Security: Validate that the output path doesn't escape the target directory
	if err := validateOutputPath(outputPath, targetDir); err != nil {
		return "", err
	}

	return outputPath, nil
}

// validateOutputPath ensures the output path stays within the target directory
// and doesn't contain path traversal sequences
func validateOutputPath(outputPath, targetDir string) error {
	// Clean both paths to resolve any . or .. components
	cleanOutput := filepath.Clean(outputPath)
	cleanTarget := filepath.Clean(targetDir)

	// Make target directory absolute for consistent comparison
	absTarget := cleanTarget
	if !filepath.IsAbs(absTarget) {
		var err error
		absTarget, err = filepath.Abs(absTarget)
		if err != nil {
			return fmt.Errorf("invalid pattern: output path would escape target directory")
		}
	}

	// Make output path absolute for consistent comparison
	absOutput := cleanOutput
	if !filepath.IsAbs(absOutput) {
		var err error
		absOutput, err = filepath.Abs(absOutput)
		if err != nil {
			return fmt.Errorf("invalid pattern: output path would escape target directory")
		}
	}

	// Get the relative path from target to output
	relPath, err := filepath.Rel(absTarget, absOutput)
	if err != nil {
		return fmt.Errorf("invalid pattern: output path would escape target directory")
	}

	// Check if the relative path tries to escape (starts with ..)
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
		return fmt.Errorf("invalid pattern: output path would escape target directory")
	}

	return nil
}

