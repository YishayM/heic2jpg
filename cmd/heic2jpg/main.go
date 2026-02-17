package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
		convertDirectory(inputPath, *pattern, *outputDir, *dryRun, forceOverwrite)
	} else {
		// Single file
		convertSingleFile(inputPath, *pattern, *outputDir, *dryRun, forceOverwrite, 1)
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
	fmt.Println("  --dry-run                   Show what would be converted without doing it")
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
	fmt.Println("  heic2jpg --dry-run                    # Preview conversions")
}



func convertSingleFile(inputPath, pattern, outputDir string, dryRun, force bool, index int) {
	// Generate output path using naming pattern
	outputPath, err := naming.GenerateOutputPath(inputPath, pattern, index, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating output path: %v\n", err)
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

	err = converter.Convert(inputPath, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Converted 1 file")
}

func convertDirectory(dirPath, pattern, outputDir string, dryRun, force bool) {
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

	for i, inputPath := range heicFiles {
		// Generate output path using naming pattern (index is 1-based)
		outputPath, err := naming.GenerateOutputPath(inputPath, pattern, i+1, outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating output path for %s: %v\n", inputPath, err)
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

		err = converter.Convert(inputPath, outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			failCount++
		} else {
			successCount++
		}
	}

	// Print summary
	if dryRun {
		fmt.Printf("✓ Would convert %d file(s) (dry-run)\n", successCount)
	} else if failCount == 0 && skippedCount == 0 {
		fmt.Printf("✓ Converted %d file(s)\n", successCount)
	} else if skippedCount > 0 && failCount == 0 {
		fmt.Printf("✓ Converted %d file(s), %d skipped (use --force)\n", successCount, skippedCount)
	} else if skippedCount > 0 && failCount > 0 {
		fmt.Printf("✓ Converted %d file(s), %d skipped (use --force), %d failed\n", successCount, skippedCount, failCount)
	} else {
		fmt.Printf("✓ Converted %d file(s), %d failed\n", successCount, failCount)
	}
}

