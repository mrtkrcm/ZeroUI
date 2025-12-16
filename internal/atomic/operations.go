package atomic

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/recovery"
)

// Manager handles atomic operations with proper locking
type Manager struct {
	locks    map[string]*sync.RWMutex // Per-file locks
	locksMu  sync.RWMutex             // Protects the locks map
	recovery *recovery.BackupManager
}

// NewManager creates a new atomic operations manager
func NewManager() (*Manager, error) {
	backupManager, err := recovery.NewBackupManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create backup manager: %w", err)
	}

	return &Manager{
		locks:    make(map[string]*sync.RWMutex),
		recovery: backupManager,
	}, nil
}

// Operation represents an atomic operation
type Operation struct {
	manager  *Manager
	filePath string
	lock     *sync.RWMutex
	backupId string
	started  time.Time
}

// BeginOperation starts an atomic operation for a specific file
func (m *Manager) BeginOperation(filePath string) *Operation {
	// Get or create lock for this file
	m.locksMu.Lock()
	lock, exists := m.locks[filePath]
	if !exists {
		lock = &sync.RWMutex{}
		m.locks[filePath] = lock
	}
	m.locksMu.Unlock()

	// Acquire write lock
	lock.Lock()

	return &Operation{
		manager:  m,
		filePath: filePath,
		lock:     lock,
		started:  time.Now(),
	}
}

// BeginReadOperation starts a read-only operation for a specific file
func (m *Manager) BeginReadOperation(filePath string) *ReadOperation {
	// Get or create lock for this file
	m.locksMu.Lock()
	lock, exists := m.locks[filePath]
	if !exists {
		lock = &sync.RWMutex{}
		m.locks[filePath] = lock
	}
	m.locksMu.Unlock()

	// Acquire read lock
	lock.RLock()

	return &ReadOperation{
		manager:  m,
		filePath: filePath,
		lock:     lock,
		started:  time.Now(),
	}
}

// CreateBackup creates a backup before making changes
func (op *Operation) CreateBackup(appName string) error {
	if op.backupId != "" {
		return fmt.Errorf("backup already created for this operation")
	}

	// Check if file exists
	if _, err := os.Stat(op.filePath); os.IsNotExist(err) {
		// No file to backup, skip
		return nil
	}

	backupId, err := op.manager.recovery.CreateBackup(op.filePath, appName)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	op.backupId = backupId
	return nil
}

// WriteConfig writes configuration data atomically
func (op *Operation) WriteConfig(appConfig *appconfig.AppConfig, configData map[string]interface{}) error {
	if op.lock == nil {
		return fmt.Errorf("operation not properly initialized")
	}

	// Create temporary file in same directory
	tempPath := op.filePath + ".tmp." + fmt.Sprintf("%d", time.Now().UnixNano())

	// Ensure directory exists
	dir := filepath.Dir(op.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temporary file first
	loader := &appconfig.Loader{}
	k := koanf.New(".")
	for key, value := range configData {
		k.Set(key, value)
	}

	if err := loader.SaveTargetConfig(&appconfig.AppConfig{
		Path:   tempPath,
		Format: appConfig.Format,
	}, k); err != nil {
		// Clean up temp file
		os.Remove(tempPath)
		return fmt.Errorf("failed to write temporary config: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, op.filePath); err != nil {
		// Clean up temp file
		os.Remove(tempPath)
		return fmt.Errorf("failed to atomically replace config file: %w", err)
	}

	return nil
}

// Commit completes the operation successfully
func (op *Operation) Commit() {
	if op.lock != nil {
		op.lock.Unlock()
		op.lock = nil
	}

	// Operation completed successfully, backup can be kept for history
	// but we don't need to do rollback
}

// Rollback rolls back the operation in case of failure
func (op *Operation) Rollback() error {
	defer func() {
		if op.lock != nil {
			op.lock.Unlock()
			op.lock = nil
		}
	}()

	if op.backupId == "" {
		// No backup to rollback to
		return nil
	}

	// The backupId is actually the full backup path returned by CreateBackup
	// Restore from backup (backupId is the backup path)
	if err := op.manager.recovery.RestoreBackup(op.backupId, op.filePath); err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}

	return nil
}

// ReadOperation represents a read-only operation
type ReadOperation struct {
	manager  *Manager
	filePath string
	lock     *sync.RWMutex
	started  time.Time
}

// ReadConfig reads configuration data with read lock
func (rop *ReadOperation) ReadConfig(appConfig *appconfig.AppConfig) (map[string]interface{}, error) {
	if rop.lock == nil {
		return nil, fmt.Errorf("read operation not properly initialized")
	}

	loader := &appconfig.Loader{}
	configObj, err := loader.LoadTargetConfig(appConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return configObj.All(), nil
}

// Complete finishes the read operation
func (rop *ReadOperation) Complete() {
	if rop.lock != nil {
		rop.lock.RUnlock()
		rop.lock = nil
	}
}

// Transaction represents a multi-operation transaction
type Transaction struct {
	manager    *Manager
	operations []*Operation
	backupIds  []string
	committed  bool
	rolledBack bool
	mu         sync.Mutex
}

// BeginTransaction starts a new transaction
func (m *Manager) BeginTransaction() *Transaction {
	return &Transaction{
		manager:    m,
		operations: make([]*Operation, 0),
		backupIds:  make([]string, 0),
	}
}

// AddOperation adds an operation to the transaction
func (t *Transaction) AddOperation(filePath string) *Operation {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return nil // Transaction already finalized
	}

	op := t.manager.BeginOperation(filePath)
	t.operations = append(t.operations, op)
	return op
}

// CreateBackups creates backups for all operations
func (t *Transaction) CreateBackups(appNames []string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("transaction already finalized")
	}

	if len(appNames) != len(t.operations) {
		return fmt.Errorf("app names count doesn't match operations count")
	}

	// Create backups for all operations
	for i, op := range t.operations {
		if err := op.CreateBackup(appNames[i]); err != nil {
			// Rollback any backups we've already created
			t.rollbackInternal()
			return fmt.Errorf("failed to create backup for operation %d: %w", i, err)
		}
		if op.backupId != "" {
			t.backupIds = append(t.backupIds, op.backupId)
		}
	}

	return nil
}

// Commit commits the entire transaction
func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("transaction already finalized")
	}

	// Commit all operations
	for _, op := range t.operations {
		op.Commit()
	}

	t.committed = true
	return nil
}

// Rollback rolls back the entire transaction
func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("transaction already finalized")
	}

	return t.rollbackInternal()
}

// rollbackInternal performs the actual rollback (must be called with lock held)
func (t *Transaction) rollbackInternal() error {
	var lastError error

	// Rollback all operations in reverse order
	for i := len(t.operations) - 1; i >= 0; i-- {
		if err := t.operations[i].Rollback(); err != nil {
			lastError = err
		}
	}

	t.rolledBack = true
	return lastError
}

// SafeOperation wraps an operation with automatic rollback on error
type SafeOperation struct {
	operation *Operation
	completed bool
}

// NewSafeOperation creates a new safe operation
func (m *Manager) NewSafeOperation(filePath string) *SafeOperation {
	return &SafeOperation{
		operation: m.BeginOperation(filePath),
	}
}

// Execute executes a function within the safe operation
func (so *SafeOperation) Execute(appName string, fn func(*Operation) error) (err error) {
	// Create backup
	if backupErr := so.operation.CreateBackup(appName); backupErr != nil {
		so.operation.Commit() // Release lock
		return fmt.Errorf("failed to create backup: %w", backupErr)
	}

	// Use defer to ensure rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			so.operation.Rollback()
			err = fmt.Errorf("operation panicked: %v", r)
		} else if err != nil && !so.completed {
			so.operation.Rollback()
		}
	}()

	// Execute the function
	err = fn(so.operation)
	if err != nil {
		return err
	}

	// Success - commit the operation
	so.operation.Commit()
	so.completed = true
	return nil
}

// LockManager provides higher-level locking utilities
type LockManager struct {
	manager *Manager
}

// NewLockManager creates a new lock manager
func NewLockManager() (*LockManager, error) {
	manager, err := NewManager()
	if err != nil {
		return nil, err
	}

	return &LockManager{
		manager: manager,
	}, nil
}

// WithReadLock executes a function with read lock on the specified file
func (lm *LockManager) WithReadLock(filePath string, fn func() error) error {
	readOp := lm.manager.BeginReadOperation(filePath)
	defer readOp.Complete()

	return fn()
}

// WithWriteLock executes a function with write lock on the specified file
func (lm *LockManager) WithWriteLock(filePath string, appName string, fn func(*Operation) error) error {
	safeOp := lm.manager.NewSafeOperation(filePath)
	return safeOp.Execute(appName, fn)
}

// WithMultipleLocks executes a function with write locks on multiple files
func (lm *LockManager) WithMultipleLocks(filePaths []string, appNames []string, fn func([]*Operation) error) error {
	if len(filePaths) != len(appNames) {
		return fmt.Errorf("file paths and app names must have same length")
	}

	tx := lm.manager.BeginTransaction()

	// Add all operations
	operations := make([]*Operation, len(filePaths))
	for i, filePath := range filePaths {
		operations[i] = tx.AddOperation(filePath)
	}

	// Create backups
	if err := tx.CreateBackups(appNames); err != nil {
		tx.Rollback()
		return err
	}

	// Execute function
	err := fn(operations)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// HealthCheck checks the health of the atomic operations system
func (m *Manager) HealthCheck() error {
	// Check if we can create a backup manager
	if m.recovery == nil {
		return fmt.Errorf("backup manager not initialized")
	}

	// Check if backup directory is accessible
	return m.recovery.HealthCheck()
}

// Stats returns statistics about the atomic operations system
func (m *Manager) Stats() map[string]interface{} {
	m.locksMu.RLock()
	defer m.locksMu.RUnlock()

	return map[string]interface{}{
		"active_locks": len(m.locks),
		"backup_stats": m.recovery.GetStats(),
	}
}
