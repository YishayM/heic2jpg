package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindHEICFiles scans a directory and returns all HEIC files found
func FindHEICFiles(dirPath string) ([]string, error) {
	return FindHEICFilesRecursive(dirPath, false)
}

// FindHEICFilesRecursive scans a directory and returns all HEIC files found
// If recursive is true, it walks through all subdirectories
func FindHEICFilesRecursive(dirPath string, recursive bool) ([]string, error) {
	var heicFiles []string

	if recursive {
		// Use filepath.WalkDir for recursive scanning
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Check if this is a symlink using Lstat to avoid following it
			info, err := os.Lstat(path)
			if err != nil {
				return err
			}

			// Skip symlinks to prevent infinite loops
			if info.Mode()&os.ModeSymlink != 0 {
				fmt.Fprintf(os.Stderr, "Skipping symlink: %s\n", path)
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip directories themselves, but continue walking into them
			if d.IsDir() {
				return nil
			}

			name := d.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".heic" || ext == ".heif" {
				heicFiles = append(heicFiles, path)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Read directory entries (non-recursive)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}

		// Filter for HEIC files
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// Check if this is a symlink
			fullPath := filepath.Join(dirPath, entry.Name())
			info, err := os.Lstat(fullPath)
			if err != nil {
				continue // Skip entries we can't stat
			}

			// Skip symlinks to prevent following them
			if info.Mode()&os.ModeSymlink != 0 {
				fmt.Fprintf(os.Stderr, "Skipping symlink: %s\n", fullPath)
				continue
			}

			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".heic" || ext == ".heif" {
				heicFiles = append(heicFiles, fullPath)
			}
		}
	}

	return heicFiles, nil
}

