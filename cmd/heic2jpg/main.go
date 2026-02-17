package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"heic2jpg/internal/converter"
	"heic2jpg/internal/naming"
	"heic2jpg/internal/scanner"
)

const version = "1.0.0"

func main() {
	// Define flags
	var (
		pattern    = flag.String("pattern", "{name}", "Output naming pattern (supports {name}, {date}, {index})")
		outputDir  = flag.String("output", "", "Output directory (default: same as input)")
		dryRun     = flag.Bool("dry-run", false, "Show what would be converted without doing it")
		force      = flag.Bool("force", false, "Overwrite existing files")
		forceShort = flag.Bool("f", false, "Overwrite existing files (short)")
		maxSize    = flag.String("max-size", "500MB", "Maximum file size to process (e.g., 500MB, 1GB, 2GB; use 0 to disable)")
		strict     = flag.Bool("strict", false, "Fail if EXIF data is missing")
		showVersion = flag.Bool("version", false, "Show version information")
		versionShort = flag.Bool("v", false, "Show version information (short)")
	)

	// Custom usage message
	flag.Usage = printUsage

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *showVersion || *versionShort {
		fmt.Printf("heic2jpg version %s\n", version)
		return
	}

	// Parse max size
	maxSizeBytes, err := parseSize(*maxSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid --max-size value: %v\n", err)
		os.Exit(1)
	}

	// Get remaining arguments (input paths)
	args := flag.Args()

	// Determine input path
	var inputPath string
	if len(args) == 0 {
		// No args: use current directory
		inputPath = "."
	} else if len(args) == 1 {
		inputPath = args[0]
	} else {
		fmt.Fprintf(os.Stderr, "Error: too many arguments\n")
		printUsage()
		os.Exit(1)
	}

	// Check if input is a directory or file
	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Combine force flags
	forceOverwrite := *force || *forceShort

	if info.IsDir() {
		// Directory: convert all HEIC files in it
		convertDirectory(inputPath, *pattern, *outputDir, *dryRun, forceOverwrite, *strict, maxSizeBytes)
	} else {
		// Single file
		convertSingleFile(inputPath, *pattern, *outputDir, *dryRun, forceOverwrite, *strict, 1, maxSizeBytes)
	}
}

func printUsage() {
	fmt.Println("heic2jpg - Convert HEIC images to JPEG format")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  heic2jpg [flags] [input]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  input                       File or directory to convert (default: current directory)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --pattern <pattern>         Output naming pattern (default: \"{name}\")")
	fmt.Println("                              Supported placeholders:")
	fmt.Println("                                {name}  - original filename without extension")
	fmt.Println("                                {date}  - file modification date (YYYY-MM-DD)")
	fmt.Println("                                {index} - numbered sequence (001, 002, 003...)")
	fmt.Println("  --output <dir>              Output directory (default: same as input)")
	fmt.Println("  --max-size <size>           Maximum file size to process (default: 500MB)")
	fmt.Println("                              Examples: 500MB, 1GB, 2GB")
	fmt.Println("                              Use 0 to disable limit (at your own risk)")
	fmt.Println("  --dry-run                   Show what would be converted without doing it")
	fmt.Println("  --strict                    Fail if EXIF data is missing")
	fmt.Println("  -f, --force                 Overwrite existing files")
	fmt.Println("  -v, --version               Show version information")
	fmt.Println("  -h, --help                  Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  heic2jpg                              # Convert all HEIC in current directory")
	fmt.Println("  heic2jpg photos/                      # Convert all HEIC in photos/")
	fmt.Println("  heic2jpg photo.heic                   # Convert single file")
	fmt.Println("  heic2jpg --pattern \"{name}_web\"       # Add suffix to filenames")
	fmt.Println("  heic2jpg --pattern \"IMG_{index}\"      # Sequential numbering")
	fmt.Println("  heic2jpg --output ./converted         # Output to specific directory")
	fmt.Println("  heic2jpg --max-size 2GB               # Allow files up to 2GB")
	fmt.Println("  heic2jpg --dry-run                    # Preview conversions")
}

// parseSize parses a human-readable size string (e.g., "500MB", "1GB", "2GB")
// Returns size in bytes, or 0 if the input is "0" (disabled)
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)

	// Special case: 0 means disabled
	if s == "0" {
		return 0, nil
	}

	// Convert to uppercase for case-insensitive matching
	s = strings.ToUpper(s)

	// Parse size with unit
	var multiplier int64
	var numStr string

	if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		multiplier = 1
		numStr = strings.TrimSuffix(s, "B")
	} else {
		// Try to parse as plain number (bytes)
		num, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid size format: %s (use format like 500MB, 1GB, 2GB)", s)
		}
		return num, nil
	}

	// Parse the numeric part
	numStr = strings.TrimSpace(numStr)
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s (use format like 500MB, 1GB, 2GB)", s)
	}

	if num < 0 {
		return 0, fmt.Errorf("size cannot be negative: %s", s)
	}

	return num * multiplier, nil
}



func convertSingleFile(inputPath, pattern, outputDir string, dryRun, force, strict bool, index int, maxSizeBytes int64) {
	// Generate output path using naming pattern
	outputPath, err := naming.GenerateOutputPath(inputPath, pattern, index, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating output path: %v\n", err)
		os.Exit(1)
	}

	// Check for input/output path collision to prevent data corruption
	if err := checkPathCollision(inputPath, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists (unless dry-run)
	if !dryRun && outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Converting %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))

	if dryRun {
		fmt.Println("  (dry-run, skipping actual conversion)")
		return
	}

	// Check if output file exists and force is not set
	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("  Output file exists: %s (use --force to overwrite)\n", filepath.Base(outputPath))
			fmt.Println("✓ Converted 0 files, 1 skipped (use --force)")
			return
		}
	}

	result, err := converter.Convert(inputPath, outputPath, maxSizeBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if strict && !result.HasEXIF {
		fmt.Fprintf(os.Stderr, "Error: No EXIF data in %s (use without --strict to convert anyway)\n", filepath.Base(inputPath))
		os.Exit(1)
	}

	fmt.Println("✓ Converted 1 file")
}

func convertDirectory(dirPath, pattern, outputDir string, dryRun, force, strict bool, maxSizeBytes int64) {
	// Find all HEIC files
	heicFiles, err := scanner.FindHEICFiles(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(heicFiles) == 0 {
		fmt.Println("No HEIC files found")
		return
	}

	// Ensure output directory exists (unless dry-run)
	if !dryRun && outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Convert each file
	successCount := 0
	failCount := 0
	skippedCount := 0
	noExifCount := 0

	for i, inputPath := range heicFiles {
		// Generate output path using naming pattern (index is 1-based)
		outputPath, err := naming.GenerateOutputPath(inputPath, pattern, i+1, outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating output path for %s: %v\n", inputPath, err)
			failCount++
			continue
		}

		// Check for input/output path collision to prevent data corruption
		if err := checkPathCollision(inputPath, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("Converting %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))

		if dryRun {
			fmt.Println("  (dry-run, skipping actual conversion)")
			successCount++
			continue
		}

		// Check if output file exists and force is not set
		if !force {
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Printf("  Output file exists: %s (use --force to overwrite)\n", filepath.Base(outputPath))
				skippedCount++
				continue
			}
		}

		result, err := converter.Convert(inputPath, outputPath, maxSizeBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			failCount++
		} else if strict && !result.HasEXIF {
			fmt.Fprintf(os.Stderr, "  Error: No EXIF data in %s (use without --strict to convert anyway)\n", filepath.Base(inputPath))
			failCount++
		} else {
			successCount++
			if !result.HasEXIF {
				noExifCount++
			}
		}
	}

	// Print summary
	if dryRun {
		fmt.Printf("✓ Would convert %d file(s) (dry-run)\n", successCount)
	} else {
		// Build summary message
		summary := fmt.Sprintf("✓ Converted %d file(s)", successCount)

		// Add EXIF-less count if any
		if noExifCount > 0 {
			summary += fmt.Sprintf(" (%d without EXIF)", noExifCount)
		}

		// Add skipped count if any
		if skippedCount > 0 {
			summary += fmt.Sprintf(", %d skipped (use --force)", skippedCount)
		}

		// Add failed count if any
		if failCount > 0 {
			summary += fmt.Sprintf(", %d failed", failCount)
		}

		fmt.Println(summary)
	}
}

// checkPathCollision checks if input and output paths resolve to the same file
// This prevents data corruption when converting a file to itself
func checkPathCollision(inputPath, outputPath string) error {
	// Resolve input path to absolute path and follow symlinks
	inputAbs, err := filepath.Abs(inputPath)
	if err != nil {
		return fmt.Errorf("failed to resolve input path: %w", err)
	}
	inputResolved, err := filepath.EvalSymlinks(inputAbs)
	if err != nil {
		// If symlink evaluation fails, use absolute path
		// (file might not exist yet, or symlink might be broken)
		inputResolved = inputAbs
	}

	// Resolve output path to absolute path and follow symlinks
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}
	outputResolved, err := filepath.EvalSymlinks(outputAbs)
	if err != nil {
		// If symlink evaluation fails, use absolute path
		// (file might not exist yet, or symlink might be broken)
		outputResolved = outputAbs
	}

	// Compare resolved paths
	if inputResolved == outputResolved {
		return fmt.Errorf("Error: input and output are the same file")
	}

	return nil
}

