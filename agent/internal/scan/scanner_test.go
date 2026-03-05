package scan_test

import (
	"os"
	"path/filepath"
	"testing"

	"agent/internal/scan"
)

func TestScanner_NativeWalk(t *testing.T) {
	// Setup a temporary directory hierarchy
	tmpDir := t.TempDir()

	// Create 2 folders, 3 files
	folderA := filepath.Join(tmpDir, "folderA")
	folderB := filepath.Join(tmpDir, "folderB")
	_ = os.MkdirAll(folderA, 0o755)
	_ = os.MkdirAll(folderB, 0o755)

	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("data"), 0o644)
	_ = os.WriteFile(filepath.Join(folderA, "file2.txt"), []byte("data2"), 0o644)
	_ = os.WriteFile(filepath.Join(folderB, "file3.txt"), []byte("data3"), 0o644)

	scanner := scan.New()

	progressCalls := 0
	onProgress := func(filesScanned int) {
		progressCalls++
	}

	result, err := scanner.Scan(tmpDir, onProgress)
	if err != nil {
		t.Fatalf("unexpected error during scan: %v", err)
	}

	// Verify stats
	if result.TotalFiles != 3 {
		t.Errorf("expected 3 files, got %d", result.TotalFiles)
	}
	// WalkDir skips explicitly appending the `.` (root) directory via the rel == "." check
	if result.TotalFolders != 2 {
		t.Errorf("expected 2 folders (excluding root), got %d", result.TotalFolders)
	}

	// Verify progress was called at least once
	if progressCalls == 0 {
		t.Errorf("expected ProgressFunc to be called at least once")
	}

	// Verify shape
	if len(result.Entries) != 5 {
		t.Errorf("expected 5 total entries (3 files + 2 sub-folders), got %d", len(result.Entries))
	}
}
