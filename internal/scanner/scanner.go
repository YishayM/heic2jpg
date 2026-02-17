package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// FindHEICFiles scans a directory and returns all HEIC files found
func FindHEICFiles(dirPath string) ([]string, error) {
	var heicFiles []string

	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	// Filter for HEIC files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".heic" {
			heicFiles = append(heicFiles, filepath.Join(dirPath, name))
		}
	}

	return heicFiles, nil
}

