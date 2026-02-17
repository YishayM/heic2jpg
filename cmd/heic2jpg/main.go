package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"heic2jpg/internal/converter"
	"heic2jpg/internal/scanner"
)

func main() {
	switch len(os.Args) {
	case 1:
		// No args: convert all HEIC in current directory
		convertDirectory(".")
	case 2:
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" {
			printUsage()
			return
		}

		// Check if arg is a directory or file
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if info.IsDir() {
			// Directory: convert all HEIC files in it
			convertDirectory(arg)
		} else {
			// Single file: auto-generate output name
			outputPath := generateOutputPath(arg)
			convertSingleFile(arg, outputPath)
		}
	case 3:
		// Two args: single file with custom output
		inputPath := os.Args[1]
		outputPath := os.Args[2]
		convertSingleFile(inputPath, outputPath)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  heic2jpg                    Convert all HEIC files in current directory")
	fmt.Println("  heic2jpg <directory>        Convert all HEIC files in specified directory")
	fmt.Println("  heic2jpg <input.heic>       Convert single file (auto-generate output name)")
	fmt.Println("  heic2jpg <input> <output>   Convert single file with custom output name")
}

func generateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	return strings.TrimSuffix(inputPath, ext) + ".jpg"
}

func convertSingleFile(inputPath, outputPath string) {
	fmt.Printf("Converting %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))

	err := converter.Convert(inputPath, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Converted 1 file")
}

func convertDirectory(dirPath string) {
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

	// Convert each file
	successCount := 0
	failCount := 0

	for _, inputPath := range heicFiles {
		outputPath := generateOutputPath(inputPath)
		fmt.Printf("Converting %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))

		err := converter.Convert(inputPath, outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			failCount++
		} else {
			successCount++
		}
	}

	// Print summary
	if failCount == 0 {
		fmt.Printf("✓ Converted %d file(s)\n", successCount)
	} else {
		fmt.Printf("✓ Converted %d file(s), %d failed\n", successCount, failCount)
	}
}

