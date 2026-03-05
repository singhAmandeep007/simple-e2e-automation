// Package scan provides directory scanning via the native Go filepath.WalkDir.
// It produces a structured list of all files and subdirectories under a given root path,
// reporting progress periodically without blocking the caller.
package scan

import (
	"io/fs"
	"path/filepath"
	"time"
)

// Entry mirrors the expected JSON output format for directory items.
type Entry struct {
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// Result is returned after a full scan.
type Result struct {
	Entries      []Entry
	TotalFiles   int
	TotalFolders int
}

// ProgressFunc is called periodically during scanning.
type ProgressFunc func(filesScanned int)

// Scanner performs directory scanning via the native Go walker.
// This abstract struct ensures the Control Plane receives consistent JSON structures.
type Scanner struct{}

// New constructs a Scanner using the standard library dir walker.
func New() *Scanner {
	return &Scanner{}
}

// Scan scans the given sourcePath and returns all entries.
// onProgress is called periodically (approx every 200ms) without blocking.
func (s *Scanner) Scan(sourcePath string, onProgress ProgressFunc) (*Result, error) {
	result := &Result{}
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		// Compute relative path
		rel, relErr := filepath.Rel(sourcePath, path)
		if relErr != nil || rel == "." {
			return nil
		}

		info, infoErr := d.Info()
		var size int64
		var modTime string
		if infoErr == nil {
			size = info.Size()
			modTime = info.ModTime().UTC().Format(time.RFC3339)
		}

		entry := Entry{
			Path:    filepath.ToSlash(rel),
			IsDir:   d.IsDir(),
			Size:    size,
			ModTime: modTime,
		}
		result.Entries = append(result.Entries, entry)

		if d.IsDir() {
			result.TotalFolders++
		} else {
			result.TotalFiles++
		}

		// Progress callback via ticker (non-blocking)
		select {
		case <-ticker.C:
			if onProgress != nil {
				onProgress(result.TotalFiles)
			}
		default:
		}
		return nil
	})

	if onProgress != nil {
		onProgress(result.TotalFiles)
	}
	return result, err
}
