package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// TempFileManager manages temporary files for safe configuration editing
type TempFileManager struct {
	tempDir   string
	tempFiles map[string]*TempFile
	mu        sync.RWMutex
	
	// Metrics
	operations uint64
	errors     uint64
	
	// Configuration
	maxBackups  int
	maxTempAge  time.Duration
	bufferSize  int
}

// TempFile represents a temporary file with integrity tracking
type TempFile struct {
	OriginalPath   string
	TempPath       string
	BackupPath     string
	OriginalHash   string
	CreatedAt      time.Time
	LockFile       string
}

// TempFileOptions configures the temporary file manager
type TempFileOptions struct {
	TempDir    string
	MaxBackups int
	MaxTempAge time.Duration
	BufferSize int
}

// DefaultTempFileOptions returns sensible defaults
func DefaultTempFileOptions() *TempFileOptions {
	return &TempFileOptions{
		TempDir:    filepath.Join(os.TempDir(), fmt.Sprintf("zeroui-temp-%d", os.Getpid())),
		MaxBackups: 5,
		MaxTempAge: 24 * time.Hour,
		BufferSize: 32 * 1024, // 32KB buffer for file operations
	}
}

// NewTempFileManager creates a new temporary file manager
func NewTempFileManager() (*TempFileManager, error) {
	return NewTempFileManagerWithOptions(DefaultTempFileOptions())
}

// NewTempFileManagerWithOptions creates a new temporary file manager with custom options
func NewTempFileManagerWithOptions(opts *TempFileOptions) (*TempFileManager, error) {
	if opts == nil {
		opts = DefaultTempFileOptions()
	}
	
	// Ensure temp directory exists with proper permissions
	if err := os.MkdirAll(opts.TempDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create temp directory %q: %w", opts.TempDir, err)
	}
	
	// Test write permissions
	testFile := filepath.Join(opts.TempDir, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return nil, fmt.Errorf("temp directory %q is not writable: %w", opts.TempDir, err)
	}
	os.Remove(testFile)

	manager := &TempFileManager{
		tempDir:    opts.TempDir,
		tempFiles:  make(map[string]*TempFile),
		maxBackups: opts.MaxBackups,
		maxTempAge: opts.MaxTempAge,
		bufferSize: opts.BufferSize,
	}
	
	// Don't start periodic cleanup in tests or for short-lived instances
	// Users can call StartPeriodicCleanup() if needed
	
	return manager, nil
}

// CreateTempCopy creates a temporary copy of the original file
func (m *TempFileManager) CreateTempCopy(originalPath string) (*TempFile, error) {
	return m.CreateTempCopyWithContext(context.Background(), originalPath)
}

// CreateTempCopyWithContext creates a temporary copy with context support
func (m *TempFileManager) CreateTempCopyWithContext(ctx context.Context, originalPath string) (*TempFile, error) {
	atomic.AddUint64(&m.operations, 1)
	
	// Normalize path for consistency
	originalPath = filepath.Clean(originalPath)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check context before proceeding
	select {
	case <-ctx.Done():
		atomic.AddUint64(&m.errors, 1)
		return nil, ctx.Err()
	default:
	}

	// Check if file is already being edited
	if existing, exists := m.tempFiles[originalPath]; exists {
		if m.isLocked(existing.LockFile) {
			atomic.AddUint64(&m.errors, 1)
			return nil, fmt.Errorf("file %q is already being edited (lock: %s)", originalPath, existing.LockFile)
		}
		// Clean up stale entry
		m.cleanup(existing)
		delete(m.tempFiles, originalPath)
	}

	// Calculate original file hash (if file exists)
	hash, err := m.calculateFileHashWithContext(ctx, originalPath)
	if err != nil && !os.IsNotExist(err) {
		atomic.AddUint64(&m.errors, 1)
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Create unique temporary file name
	timestamp := time.Now().UnixNano()
	randomSuffix := fmt.Sprintf("%d_%d", os.Getpid(), timestamp)
	tempFileName := fmt.Sprintf("config_%s.tmp", randomSuffix)
	tempPath := filepath.Join(m.tempDir, tempFileName)

	// Copy original to temp with context
	if err := m.copyFileWithContext(ctx, originalPath, tempPath); err != nil {
		atomic.AddUint64(&m.errors, 1)
		return nil, fmt.Errorf("failed to create temp copy: %w", err)
	}

	// Create lock file with process info
	lockFile := tempPath + ".lock"
	lockInfo := fmt.Sprintf("%d:%s:%d", os.Getpid(), runtime.GOOS, time.Now().Unix())
	if err := os.WriteFile(lockFile, []byte(lockInfo), 0600); err != nil {
		os.Remove(tempPath)
		atomic.AddUint64(&m.errors, 1)
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	// Create backup path
	backupPath := originalPath + ".backup"

	tempFile := &TempFile{
		OriginalPath: originalPath,
		TempPath:     tempPath,
		BackupPath:   backupPath,
		OriginalHash: hash,
		CreatedAt:    time.Now(),
		LockFile:     lockFile,
	}

	m.tempFiles[originalPath] = tempFile
	return tempFile, nil
}

// ValidateTemp validates the temporary file before saving
func (m *TempFileManager) ValidateTemp(tempFile *TempFile) error {
	// Check if temp file exists
	if _, err := os.Stat(tempFile.TempPath); err != nil {
		return fmt.Errorf("temporary file not found: %w", err)
	}

	// Check if temp file is not empty
	info, err := os.Stat(tempFile.TempPath)
	if err != nil {
		return fmt.Errorf("failed to stat temp file: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("temporary file is empty")
	}

	return nil
}

// CommitTemp commits the temporary file to the original location
func (m *TempFileManager) CommitTemp(tempFile *TempFile) error {
	return m.CommitTempWithContext(context.Background(), tempFile)
}

// CommitTempWithContext commits with context support and improved error handling
func (m *TempFileManager) CommitTempWithContext(ctx context.Context, tempFile *TempFile) error {
	if tempFile == nil {
		return fmt.Errorf("tempFile cannot be nil")
	}
	
	atomic.AddUint64(&m.operations, 1)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check context
	select {
	case <-ctx.Done():
		atomic.AddUint64(&m.errors, 1)
		return ctx.Err()
	default:
	}

	// Validate temp file first
	if err := m.ValidateTemp(tempFile); err != nil {
		atomic.AddUint64(&m.errors, 1)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create backup of original if it exists
	originalExists := false
	if info, err := os.Stat(tempFile.OriginalPath); err == nil {
		originalExists = true
		// Ensure we're not overwriting a directory
		if info.IsDir() {
			atomic.AddUint64(&m.errors, 1)
			return fmt.Errorf("cannot overwrite directory %q", tempFile.OriginalPath)
		}
		
		if err := m.createBackupWithRotation(tempFile.OriginalPath, tempFile.BackupPath); err != nil {
			atomic.AddUint64(&m.errors, 1)
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(tempFile.OriginalPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		atomic.AddUint64(&m.errors, 1)
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Platform-specific atomic rename
	commitErr := m.atomicRename(tempFile.TempPath, tempFile.OriginalPath)
	if commitErr != nil {
		// Try to restore backup if rename failed and we had an original
		if originalExists {
			if restoreErr := m.atomicRename(tempFile.BackupPath, tempFile.OriginalPath); restoreErr != nil {
				// Critical error: couldn't restore backup
				atomic.AddUint64(&m.errors, 1)
				return fmt.Errorf("commit failed and backup restore failed: commit=%v, restore=%v", commitErr, restoreErr)
			}
		}
		atomic.AddUint64(&m.errors, 1)
		return fmt.Errorf("failed to commit changes: %w", commitErr)
	}

	// Verify the committed file
	if _, err := os.Stat(tempFile.OriginalPath); err != nil {
		// Try to restore from backup
		if originalExists {
			m.atomicRename(tempFile.BackupPath, tempFile.OriginalPath)
		}
		atomic.AddUint64(&m.errors, 1)
		return fmt.Errorf("commit verification failed: %w", err)
	}

	// Clean up
	m.cleanup(tempFile)
	delete(m.tempFiles, tempFile.OriginalPath)

	return nil
}

// Rollback discards changes and restores from backup if needed
func (m *TempFileManager) Rollback(tempFile *TempFile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clean up temp file
	m.cleanup(tempFile)
	delete(m.tempFiles, tempFile.OriginalPath)

	// Restore from backup if it exists
	if _, err := os.Stat(tempFile.BackupPath); err == nil {
		if err := os.Rename(tempFile.BackupPath, tempFile.OriginalPath); err != nil {
			return fmt.Errorf("failed to restore from backup: %w", err)
		}
	}

	return nil
}

// GetTempFile returns the temporary file for the given original path
func (m *TempFileManager) GetTempFile(originalPath string) (*TempFile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tempFile, exists := m.tempFiles[originalPath]
	return tempFile, exists
}

// CleanupStale removes stale temporary files older than duration
func (m *TempFileManager) CleanupStale(maxAge time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var toDelete []string

	for path, tempFile := range m.tempFiles {
		if now.Sub(tempFile.CreatedAt) > maxAge {
			// For stale files, we force cleanup even if locked
			// since the process might have died
			toDelete = append(toDelete, path)
			m.cleanup(tempFile)
		}
	}

	for _, path := range toDelete {
		delete(m.tempFiles, path)
	}

	return nil
}

// CleanupAll removes all temporary files
func (m *TempFileManager) CleanupAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tempFile := range m.tempFiles {
		m.cleanup(tempFile)
	}
	m.tempFiles = make(map[string]*TempFile)
}

// calculateFileHash calculates SHA-256 hash of a file
func (m *TempFileManager) calculateFileHash(path string) (string, error) {
	return m.calculateFileHashWithContext(context.Background(), path)
}

// calculateFileHashWithContext calculates hash with context support
func (m *TempFileManager) calculateFileHashWithContext(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // New file, no hash
		}
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	
	// Use buffered copy with context checking
	buf := make([]byte, m.bufferSize)
	for {
		// Check context before reading
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		
		n, err := file.Read(buf)
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// copyFile copies a file from src to dst
func (m *TempFileManager) copyFile(src, dst string) error {
	return m.copyFileWithContext(context.Background(), src, dst)
}

// copyFileWithContext copies with context support and optimized buffering
func (m *TempFileManager) copyFileWithContext(ctx context.Context, src, dst string) error {
	// If source doesn't exist, create empty temp file
	srcInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		return os.WriteFile(dst, []byte{}, 0600)
	}
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	
	// Don't copy directories
	if srcInfo.IsDir() {
		return fmt.Errorf("cannot copy directory %q", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer source.Close()

	// Create destination with same permissions as source
	destination, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	
	// Ensure cleanup on error
	success := false
	defer func() {
		destination.Close()
		if !success {
			os.Remove(dst)
		}
	}()

	// Copy with context support
	buf := make([]byte, m.bufferSize)
	for {
		// Check context before reading
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		n, readErr := source.Read(buf)
		if n > 0 {
			if _, writeErr := destination.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read error: %w", readErr)
		}
	}
	
	// Ensure all data is written to disk
	if err := destination.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination: %w", err)
	}
	
	success = true
	return nil
}

// createBackup creates a backup with rotation
func (m *TempFileManager) createBackup(src, dst string) error {
	return m.createBackupWithRotation(src, dst)
}

// createBackupWithRotation creates a backup with configurable rotation
func (m *TempFileManager) createBackupWithRotation(src, dst string) error {
	// Rotate existing backups (keep maxBackups versions)
	for i := m.maxBackups - 1; i > 0; i-- {
		oldBackup := fmt.Sprintf("%s.%d", dst, i)
		newBackup := fmt.Sprintf("%s.%d", dst, i+1)
		if _, err := os.Stat(oldBackup); err == nil {
			if err := os.Rename(oldBackup, newBackup); err != nil {
				// Log but don't fail on rotation errors
				continue
			}
		}
	}

	// Rename current backup to .backup.1
	if _, err := os.Stat(dst); err == nil {
		backupOne := dst + ".1"
		os.Rename(dst, backupOne)
	}

	// Create new backup
	return m.copyFile(src, dst)
}

// atomicRename performs an atomic rename operation with platform-specific handling
func (m *TempFileManager) atomicRename(src, dst string) error {
	// On Windows, we need to remove the destination first
	if runtime.GOOS == "windows" {
		// Try to remove destination if it exists
		if _, err := os.Stat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				// If we can't remove, try to move it aside
				tempDst := dst + ".old"
				os.Rename(dst, tempDst)
				defer os.Remove(tempDst)
			}
		}
	}
	
	// Perform the rename
	if err := os.Rename(src, dst); err != nil {
		// If rename fails, try copy and delete
		if copyErr := m.copyFile(src, dst); copyErr != nil {
			return fmt.Errorf("rename failed, copy also failed: rename=%v, copy=%v", err, copyErr)
		}
		// Remove source after successful copy
		os.Remove(src)
		return nil
	}
	
	return nil
}

// StartPeriodicCleanup starts a goroutine that periodically cleans up stale files
func (m *TempFileManager) StartPeriodicCleanup(ctx context.Context, interval time.Duration) {
	go m.periodicCleanup(ctx, interval)
}

// periodicCleanup runs periodic cleanup of stale files
func (m *TempFileManager) periodicCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.CleanupStale(m.maxTempAge); err != nil {
				// Log error but continue
				continue
			}
		}
	}
}

// cleanup removes temporary files and lock
func (m *TempFileManager) cleanup(tempFile *TempFile) {
	os.Remove(tempFile.TempPath)
	os.Remove(tempFile.LockFile)
}

// isLocked checks if a lock file is valid
func (m *TempFileManager) isLocked(lockFile string) bool {
	data, err := os.ReadFile(lockFile)
	if err != nil {
		return false
	}

	// Parse lock info: pid:os:timestamp
	var pid int
	var osName string
	var timestamp int64
	fmt.Sscanf(string(data), "%d:%s:%d", &pid, &osName, &timestamp)
	
	// Check if it's our own process
	if pid == os.Getpid() {
		return true // We own the lock
	}
	
	// Check if lock is too old (stale)
	if timestamp > 0 {
		lockTime := time.Unix(timestamp, 0)
		if time.Since(lockTime) > m.maxTempAge {
			return false // Lock is stale
		}
	}
	
	// For other processes, we consider the lock valid if it has a valid PID
	return pid > 0
}

// GetMetrics returns current metrics for monitoring
func (m *TempFileManager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	tempFileCount := len(m.tempFiles)
	m.mu.RUnlock()
	
	return map[string]interface{}{
		"operations":       atomic.LoadUint64(&m.operations),
		"errors":          atomic.LoadUint64(&m.errors),
		"temp_files":      tempFileCount,
		"temp_dir":        m.tempDir,
		"max_backups":     m.maxBackups,
		"max_temp_age":    m.maxTempAge.String(),
		"buffer_size":     m.bufferSize,
	}
}

// Close cleans up the temp manager
func (m *TempFileManager) Close() error {
	m.CleanupAll()
	// Remove temp directory if empty
	os.Remove(m.tempDir)
	return nil
}