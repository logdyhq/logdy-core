package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RotatingWriter implements a process-restart-safe log rotation
type RotatingWriter struct {
	filename    string
	maxSize     int64
	maxBackups  int
	currentFile *os.File
	currentSize int64
	mu          sync.Mutex
}

// NewRotatingWriter creates a new rotating writer with restart safety
func NewRotatingWriter(filename string, maxSize int64, maxBackups int) (*RotatingWriter, error) {
	rw := &RotatingWriter{
		filename:   filename,
		maxSize:    maxSize,
		maxBackups: maxBackups,
	}

	// Recover from any incomplete rotation and open file
	if err := rw.recoverAndOpenFile(); err != nil {
		return nil, err
	}

	return rw, nil
}

// Write implements the io.Writer interface with restart safety
func (rw *RotatingWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Re-check file size in case of external modifications
	if err := rw.refreshFileSize(); err != nil {
		return 0, err
	}

	// Check if we need to rotate before writing
	if rw.currentSize+int64(len(p)) > rw.maxSize {
		if err := rw.safeRotate(); err != nil {
			return 0, err
		}
	}

	// Write to current file
	n, err = rw.currentFile.Write(p)
	if err != nil {
		return n, err
	}

	// Sync to disk for durability (optional but safer)
	if err := rw.currentFile.Sync(); err != nil {
		return n, err
	}

	rw.currentSize += int64(n)
	return n, nil
}

// recoverAndOpenFile handles restart recovery and file opening
func (rw *RotatingWriter) recoverAndOpenFile() error {
	// Check for incomplete rotation (temp files left behind)
	tempFile := rw.filename + ".tmp"
	if _, err := os.Stat(tempFile); err == nil {
		// Remove incomplete temp file
		os.Remove(tempFile)
	}

	// Check for incomplete backup rotation
	rw.cleanupIncompleteRotation()

	// Open the main log file
	return rw.openFile()
}

// cleanupIncompleteRotation cleans up any partial rotation state
func (rw *RotatingWriter) cleanupIncompleteRotation() {
	// Look for .rotating files which indicate incomplete rotation
	rotatingFile := rw.filename + ".rotating"
	if _, err := os.Stat(rotatingFile); err == nil {
		// Found incomplete rotation marker, clean up
		os.Remove(rotatingFile)

		// The main file might be in an inconsistent state
		// Check if backup.1 exists and is newer
		backup1 := rw.backupName(1)
		if mainInfo, err1 := os.Stat(rw.filename); err1 == nil {
			if backupInfo, err2 := os.Stat(backup1); err2 == nil {
				// If backup is newer than main, rotation was likely interrupted
				if backupInfo.ModTime().After(mainInfo.ModTime()) {
					// Restore the backup as main file
					os.Remove(rw.filename)
					os.Rename(backup1, rw.filename)
				}
			}
		}
	}
}

// refreshFileSize updates the current file size (handles external changes)
func (rw *RotatingWriter) refreshFileSize() error {
	if rw.currentFile == nil {
		return fmt.Errorf("no open file")
	}

	info, err := rw.currentFile.Stat()
	if err != nil {
		return err
	}

	rw.currentSize = info.Size()
	return nil
}

// openFile opens the current log file and gets its size
func (rw *RotatingWriter) openFile() error {
	file, err := os.OpenFile(rw.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	rw.currentFile = file
	rw.currentSize = info.Size()
	return nil
}

// safeRotate performs atomic log rotation with crash safety
func (rw *RotatingWriter) safeRotate() error {
	// Create rotation marker file
	rotatingMarker := rw.filename + ".rotating"
	markerFile, err := os.Create(rotatingMarker)
	if err != nil {
		return err
	}
	markerFile.Close()

	// Close current file
	if rw.currentFile != nil {
		rw.currentFile.Sync() // Ensure all data is written
		rw.currentFile.Close()
	}

	// Perform rotation atomically
	if err := rw.performAtomicRotation(); err != nil {
		// Clean up marker on failure
		os.Remove(rotatingMarker)
		// Try to reopen the original file
		rw.openFile()
		return err
	}

	// Remove rotation marker (rotation completed successfully)
	os.Remove(rotatingMarker)

	// Open new log file
	return rw.openFile()
}

// performAtomicRotation does the actual file rotation
func (rw *RotatingWriter) performAtomicRotation() error {
	// First, shift all backup files
	for i := rw.maxBackups; i > 0; i-- {
		oldPath := rw.backupName(i)
		newPath := rw.backupName(i + 1)

		if i == rw.maxBackups {
			// Remove the oldest backup
			os.Remove(oldPath)
		} else {
			// Rename backup files (ignore errors for non-existent files)
			if _, err := os.Stat(oldPath); err == nil {
				if err := os.Rename(oldPath, newPath); err != nil {
					return fmt.Errorf("failed to rotate backup %s to %s: %v", oldPath, newPath, err)
				}
			}
		}
	}

	// Move current log to backup.1
	if _, err := os.Stat(rw.filename); err == nil {
		if err := os.Rename(rw.filename, rw.backupName(1)); err != nil {
			return fmt.Errorf("failed to move current log to backup: %v", err)
		}
	}

	return nil
}

// backupName generates backup file names
func (rw *RotatingWriter) backupName(n int) string {
	ext := filepath.Ext(rw.filename)
	base := rw.filename[:len(rw.filename)-len(ext)]
	return fmt.Sprintf("%s.%d%s", base, n, ext)
}

// Close closes the current log file safely
func (rw *RotatingWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.currentFile != nil {
		// Ensure all data is written before closing
		rw.currentFile.Sync()
		return rw.currentFile.Close()
	}
	return nil
}
