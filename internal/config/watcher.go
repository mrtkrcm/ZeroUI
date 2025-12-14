package config

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher provides debounced file watching capabilities
type FileWatcher struct {
	watcher      *fsnotify.Watcher
	debounceTime time.Duration
	callbacks    map[string]func(string)
	mu           sync.RWMutex
	timers       map[string]*time.Timer
	timerMu      sync.Mutex
}

// NewFileWatcher creates a new file watcher with debouncing
func NewFileWatcher(debounceTime time.Duration) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:      watcher,
		debounceTime: debounceTime,
		callbacks:    make(map[string]func(string)),
		timers:       make(map[string]*time.Timer),
	}

	// Start the event processor
	go fw.processEvents()

	return fw, nil
}

// Watch adds a file or directory to watch with a callback
func (fw *FileWatcher) Watch(path string, callback func(string)) error {
	// Add to watcher
	if err := fw.watcher.Add(path); err != nil {
		return err
	}

	// Store callback
	fw.mu.Lock()
	fw.callbacks[path] = callback
	fw.mu.Unlock()

	return nil
}

// Unwatch removes a file or directory from watching
func (fw *FileWatcher) Unwatch(path string) error {
	// Remove from watcher
	if err := fw.watcher.Remove(path); err != nil {
		return err
	}

	// Remove callback
	fw.mu.Lock()
	delete(fw.callbacks, path)
	fw.mu.Unlock()

	// Cancel any pending timer
	fw.timerMu.Lock()
	if timer, exists := fw.timers[path]; exists {
		timer.Stop()
		delete(fw.timers, path)
	}
	fw.timerMu.Unlock()

	return nil
}

// processEvents processes file system events with debouncing
func (fw *FileWatcher) processEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err // Handle error appropriately in production
		}
	}
}

// handleEvent handles a single file system event with debouncing
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Ignore certain events
	if event.Op&fsnotify.Chmod == fsnotify.Chmod {
		return // Ignore permission changes
	}

	// Get the callback for this path
	fw.mu.RLock()
	callback, exists := fw.callbacks[event.Name]
	if !exists {
		// Check if it's a file in a watched directory
		dir := filepath.Dir(event.Name)
		callback, exists = fw.callbacks[dir]
	}
	fw.mu.RUnlock()

	if !exists {
		return
	}

	// Debounce the event
	fw.debounceEvent(event.Name, callback)
}

// debounceEvent debounces events for a specific path
func (fw *FileWatcher) debounceEvent(path string, callback func(string)) {
	fw.timerMu.Lock()
	defer fw.timerMu.Unlock()

	// Cancel existing timer if present
	if timer, exists := fw.timers[path]; exists {
		timer.Stop()
	}

	// Create new timer
	fw.timers[path] = time.AfterFunc(fw.debounceTime, func() {
		// Execute callback
		callback(path)

		// Clean up timer
		fw.timerMu.Lock()
		delete(fw.timers, path)
		fw.timerMu.Unlock()
	})
}

// Close closes the file watcher
func (fw *FileWatcher) Close() error {
	// Stop all timers
	fw.timerMu.Lock()
	for _, timer := range fw.timers {
		timer.Stop()
	}
	fw.timers = nil
	fw.timerMu.Unlock()

	// Close watcher
	return fw.watcher.Close()
}

// DebouncedWatcher provides a simpler interface for debounced file watching
type DebouncedWatcher struct {
	watcher  *FileWatcher
	onChange func(string)
}

// NewDebouncedWatcher creates a new debounced watcher with default settings
func NewDebouncedWatcher(onChange func(string)) (*DebouncedWatcher, error) {
	// Use 100ms debounce by default
	watcher, err := NewFileWatcher(100 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	return &DebouncedWatcher{
		watcher:  watcher,
		onChange: onChange,
	}, nil
}

// Watch adds a path to watch
func (dw *DebouncedWatcher) Watch(path string) error {
	return dw.watcher.Watch(path, dw.onChange)
}

// Unwatch removes a path from watching
func (dw *DebouncedWatcher) Unwatch(path string) error {
	return dw.watcher.Unwatch(path)
}

// Close closes the watcher
func (dw *DebouncedWatcher) Close() error {
	return dw.watcher.Close()
}
