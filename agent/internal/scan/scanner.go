package scan

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Entry mirrors the rclone lsjson output format.
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

// Scanner performs directory scanning via rclone (if available) or native Go walker.
type Scanner struct {
	rcloneBin string // path to rclone binary; empty means native walker
}

func New(rcloneBin string) *Scanner {
	return &Scanner{rcloneBin: rcloneBin}
}

// Scan scans the given sourcePath and returns all entries.
// onProgress is called after each directory batch (approx every 100ms).
func (s *Scanner) Scan(sourcePath string, onProgress ProgressFunc) (*Result, error) {
	if s.rcloneBin != "" {
		if _, err := os.Stat(s.rcloneBin); err == nil {
			return s.rcloneScan(sourcePath, onProgress)
		}
	}
	// Fallback to native Go walker
	return s.nativeScan(sourcePath, onProgress)
}

// rcloneScan uses `rclone lsjson --recursive` to scan.
func (s *Scanner) rcloneScan(sourcePath string, onProgress ProgressFunc) (*Result, error) {
	cmd := exec.Command(s.rcloneBin, "lsjson", sourcePath, "--recursive", "--no-modtime=false")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("rclone lsjson: %w", err)
	}

	type rcloneEntry struct {
		Path    string `json:"Path"`
		Name    string `json:"Name"`
		IsDir   bool   `json:"IsDir"`
		Size    int64  `json:"Size"`
		ModTime string `json:"ModTime"`
	}
	var raw []rcloneEntry
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing rclone output: %w", err)
	}

	result := &Result{}
	for i, r := range raw {
		entry := Entry{
			Path:    r.Path,
			IsDir:   r.IsDir,
			Size:    r.Size,
			ModTime: r.ModTime,
		}
		result.Entries = append(result.Entries, entry)
		if r.IsDir {
			result.TotalFolders++
		} else {
			result.TotalFiles++
		}
		if onProgress != nil && i%20 == 0 {
			onProgress(result.TotalFiles)
		}
	}
	if onProgress != nil {
		onProgress(result.TotalFiles)
	}
	return result, nil
}

// nativeScan uses filepath.WalkDir as rclone fallback.
func (s *Scanner) nativeScan(sourcePath string, onProgress ProgressFunc) (*Result, error) {
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
