package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"

	"heic2jpg/internal/converter"
	"heic2jpg/internal/naming"
	"heic2jpg/internal/scanner"

	"golang.org/x/term"
)

const version = "1.0.0"

func main() {
	// Define flags
	var (
		pattern        = flag.String("pattern", "{name}", "Output naming pattern (supports {name}, {date}, {index})")
		outputDir      = flag.String("output", "", "Output directory (default: same as input)")
		dryRun         = flag.Bool("dry-run", false, "Show what would be converted without doing it")
		force          = flag.Bool("force", false, "Overwrite existing files")
		forceShort     = flag.Bool("f", false, "Overwrite existing files (short)")
		recursive      = flag.Bool("recursive", false, "Scan directories recursively for nested HEIC files")
		recursiveShort = flag.Bool("r", false, "Scan directories recursively for nested HEIC files (short)")
		maxSize        = flag.String("max-size", "500MB", "Maximum file size to process (e.g., 500MB, 1GB, 2GB; use 0 to disable)")
		strict         = flag.Bool("strict", false, "Fail if EXIF data is missing")
		yes            = flag.Bool("yes", false, "Skip confirmation prompts")
		yesShort       = flag.Bool("y", false, "Skip confirmation prompts (short)")
		showVersion    = flag.Bool("version", false, "Show version information")
		versionShort   = flag.Bool("v", false, "Show version information (short)")
	)

	// Custom usage message
	flag.Usage = printUsage

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *showVersion || *versionShort {
		fmt.Printf("heic2jpg %s (%s/%s, %s)\n", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
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
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", inputPath)
		} else if os.IsPermission(err) {
			fmt.Fprintf(os.Stderr, "Error: Permission denied: %s\n", inputPath)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Combine force flags
	forceOverwrite := *force || *forceShort
	// Combine recursive flags
	recursiveScan := *recursive || *recursiveShort
	// Combine yes flags
	skipConfirmation := *yes || *yesShort

	if info.IsDir() {
		// Directory: convert all HEIC files in it
		convertDirectory(inputPath, *pattern, *outputDir, *dryRun, forceOverwrite, *strict, maxSizeBytes, recursiveScan, skipConfirmation)
	} else {
		// Single file
		convertSingleFile(inputPath, *pattern, *outputDir, *dryRun, forceOverwrite, *strict, 1, maxSizeBytes)
	}
}

func printUsage() {
	fmt.Println("heic2jpg - Convert HEIC images to JPEG format")
	fmt.Println()
	fmt.Println("Supported extensions: .heic, .HEIC, .heif, .HEIF")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  heic2jpg [flags] [input]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  input                       File or directory to convert (default: current directory)")
	fmt.Println()
	fmt.Println("Common Flags:")
	fmt.Println("  --pattern <pattern>         Output naming pattern (default: \"{name}\")")
	fmt.Println("                              Supported placeholders:")
	fmt.Println("                                {name}  - original filename without extension")
	fmt.Println("                                {date}  - file modification date (YYYY-MM-DD)")
	fmt.Println("                                {index} - numbered sequence (001, 002, 003...)")
	fmt.Println("  --output <dir>              Output directory (default: same as input)")
	fmt.Println("  -r, --recursive             Scan directories recursively for nested HEIC files")
	fmt.Println()
	fmt.Println("Safety Flags:")
	fmt.Println("  --dry-run                   Show what would be converted without doing it")
	fmt.Println("  -f, --force                 Overwrite existing files")
	fmt.Println("  --strict                    Fail if EXIF data is missing")
	fmt.Println("  -y, --yes                   Skip confirmation prompts for large batches")
	fmt.Println()
	fmt.Println("Advanced Flags:")
	fmt.Println("  --max-size <size>           Maximum file size to process (default: 500MB)")
	fmt.Println("                              Examples: 500MB, 1GB, 2GB")
	fmt.Println("                              Use 0 to disable limit (at your own risk)")
	fmt.Println()
	fmt.Println("Other Flags:")
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
	fmt.Println("  heic2jpg --recursive ./photos         # Convert all HEIC in nested folders")
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
	// Generate output path using naming pattern (totalCount=1 for single file)
	outputPath, err := naming.GenerateOutputPath(inputPath, pattern, index, outputDir, 1)
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

	if dryRun {
		fmt.Printf("[DRY-RUN] Would convert %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))
		return
	}

	fmt.Printf("Converting %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))

	// Check if output file exists and force is not set
	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("  Output file exists: %s (use --force to overwrite)\n", filepath.Base(outputPath))
			fmt.Println("✓ Converted 0 files, 1 skipped (use --force)")
			return
		}
	}

	// In strict mode, check EXIF before converting
	if strict {
		hasEXIF, err := converter.CheckEXIF(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking EXIF: %v\n", err)
			os.Exit(1)
		}
		if !hasEXIF {
			fmt.Fprintf(os.Stderr, "Error: No EXIF data in %s (use without --strict to convert anyway)\n", filepath.Base(inputPath))
			os.Exit(1)
		}
	}

	_, err = converter.Convert(inputPath, outputPath, maxSizeBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		// Exit with code 28 (ENOSPC) for disk full errors
		if converter.IsDiskFullError(err) {
			os.Exit(28)
		}
		os.Exit(1)
	}

	fmt.Println("✓ Converted 1 file")
}

// conversionError holds details about a failed conversion
type conversionError struct {
	filePath string
	err      error
}

func convertDirectory(dirPath, pattern, outputDir string, dryRun, force, strict bool, maxSizeBytes int64, recursive bool, skipConfirmation bool) {
	// Find all HEIC files
	heicFiles, err := scanner.FindHEICFilesRecursive(dirPath, recursive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(heicFiles) == 0 {
		fmt.Println("No HEIC files found")
		return
	}

	// Show file count
	fileCount := len(heicFiles)
	if fileCount == 1 {
		fmt.Println("Found 1 HEIC file.")
	} else {
		fmt.Printf("Found %d HEIC files.\n", fileCount)
	}

	// Prompt for confirmation if batch is large (>50 files), we're in a TTY, and --yes not set
	if fileCount > 50 && isTTY() && !skipConfirmation {
		fmt.Print("Continue? [Y/n] ")
		var response string
		fmt.Scanln(&response)
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "n" || response == "no" {
			fmt.Println("Cancelled.")
			return
		}
	}

	// Convert each file
	successCount := 0
	failCount := 0
	skippedCount := 0
	noExifCount := 0
	var errors []conversionError

	// Check if we should show progress bar (TTY and not dry-run)
	showProgress := isTTY() && !dryRun
	totalFiles := len(heicFiles)

	// Set up signal handling for graceful cancellation
	var interrupted atomic.Bool
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle signals in a goroutine
	go func() {
		<-sigChan
		interrupted.Store(true)
		fmt.Println("\n\nInterrupted! Finishing current file and showing summary...")
	}()

	// Show cancel hint for large batches (10+ files)
	if totalFiles >= 10 {
		if dryRun {
			fmt.Printf("[DRY-RUN] Would convert %d files... (Ctrl+C to cancel)\n", totalFiles)
		} else {
			fmt.Printf("Converting %d files... (Ctrl+C to cancel)\n", totalFiles)
		}
	}

	for i, inputPath := range heicFiles {
		// Check if we've been interrupted
		if interrupted.Load() {
			break
		}

		// Calculate the relative path from the input directory for recursive mode
		var relPath string
		if recursive {
			relPath, _ = filepath.Rel(dirPath, inputPath)
		} else {
			relPath = filepath.Base(inputPath)
		}

		// Generate the actual output directory, preserving structure in recursive mode
		actualOutputDir := outputDir
		if recursive && outputDir != "" {
			// Get the relative directory of the input file
			relDir := filepath.Dir(relPath)
			if relDir != "." {
				actualOutputDir = filepath.Join(outputDir, relDir)
			}
		}

		// Ensure output directory exists (unless dry-run)
		if !dryRun && actualOutputDir != "" {
			err := os.MkdirAll(actualOutputDir, 0755)
			if err != nil {
				errors = append(errors, conversionError{filePath: relPath, err: fmt.Errorf("creating output directory: %w", err)})
				failCount++
				continue
			}
		}

		// Generate output path using naming pattern (index is 1-based)
		outputPath, err := naming.GenerateOutputPath(inputPath, pattern, i+1, actualOutputDir, totalFiles)
		if err != nil {
			errors = append(errors, conversionError{filePath: relPath, err: fmt.Errorf("generating output path: %w", err)})
			failCount++
			continue
		}

		// Check for input/output path collision to prevent data corruption
		if err := checkPathCollision(inputPath, outputPath); err != nil {
			errors = append(errors, conversionError{filePath: relPath, err: err})
			failCount++
			continue
		}

		// Show relative path for recursive mode, filename only for non-recursive
		var displayInput, displayOutput string
		if recursive {
			displayInput = relPath
			// Calculate relative output path if output dir is set
			if outputDir != "" {
				displayOutput, _ = filepath.Rel(outputDir, outputPath)
			} else {
				displayOutput = relPath[:len(relPath)-len(filepath.Ext(relPath))] + ".jpg"
			}
		} else {
			displayInput = filepath.Base(inputPath)
			displayOutput = filepath.Base(outputPath)
		}

		// Show progress bar or individual file message
		if showProgress {
			drawProgressBar(i+1, totalFiles)
		} else {
			if dryRun {
				fmt.Printf("[DRY-RUN] Would convert %s → %s\n", displayInput, displayOutput)
			} else {
				fmt.Printf("Converting %s → %s\n", displayInput, displayOutput)
			}
		}

		if dryRun {
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

		// In strict mode, check EXIF before converting
		if strict {
			hasEXIF, err := converter.CheckEXIF(inputPath)
			if err != nil {
				errors = append(errors, conversionError{filePath: relPath, err: fmt.Errorf("checking EXIF: %w", err)})
				failCount++
				continue
			}
			if !hasEXIF {
				errors = append(errors, conversionError{filePath: relPath, err: fmt.Errorf("no EXIF data (use without --strict to convert anyway)")})
				failCount++
				continue
			}
		}

		result, err := converter.Convert(inputPath, outputPath, maxSizeBytes)
		if err != nil {
			// Exit immediately with code 28 for disk full errors
			if converter.IsDiskFullError(err) {
				os.Exit(28)
			}
			errors = append(errors, conversionError{filePath: relPath, err: err})
			failCount++
		} else {
			successCount++
			if !result.HasEXIF {
				noExifCount++
			}
		}
	}

	// Clear progress bar before printing summary
	if showProgress {
		// Clear the progress line by printing spaces and carriage return
		fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
	}

	// Print summary
	if dryRun {
		if interrupted.Load() {
			fmt.Printf("✓ Would convert %d file(s) (dry-run, interrupted)\n", successCount)
		} else {
			fmt.Printf("✓ Would convert %d file(s) (dry-run)\n", successCount)
		}
	} else {
		// Build summary message with color
		checkmark := colorize("✓", colorGreen)
		summary := fmt.Sprintf("%s Converted %d file(s)", checkmark, successCount)

		// Add interruption note if interrupted
		if interrupted.Load() {
			summary += " (interrupted)"
		}

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

		// Display error details if any failures occurred
		if len(errors) > 0 {
			fmt.Println()
			crossmark := colorize("✗", colorRed)
			fmt.Printf("%s %d file(s) failed:\n", crossmark, len(errors))
			for _, e := range errors {
				errorMsg := colorize(fmt.Sprintf("  %s: %v", e.filePath, e.err), colorRed)
				fmt.Println(errorMsg)
			}
		}
	}

	// Exit with error code if any failures occurred
	if failCount > 0 {
		os.Exit(1)
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

// isTTY checks if stdout is a terminal (not piped or redirected)
func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
)

// colorize wraps text in ANSI color codes if output is a TTY
func colorize(text, color string) string {
	if isTTY() {
		return color + text + colorReset
	}
	return text
}

// drawProgressBar renders a progress bar with the format:
// Converting... [████████░░░░░░░░] 47/100 files
func drawProgressBar(current, total int) {
	const barWidth = 20

	// Calculate progress
	progress := float64(current) / float64(total)
	filledWidth := int(progress * float64(barWidth))

	// Build the bar
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	// Print with carriage return to update in-place
	fmt.Printf("\rConverting... %s %d/%d files", bar, current, total)
}

