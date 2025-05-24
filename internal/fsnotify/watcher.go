package fsnotify

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// EventType represents the type of file system event
type EventType int

const (
	// EventCreate represents a file creation event
	EventCreate EventType = iota
	// EventModify represents a file modification event
	EventModify
	// EventRemove represents a file removal event
	EventRemove
	// EventRename represents a file rename event
	EventRename
)

// Event represents a file system event
type Event struct {
	Type EventType
	Path string
}

// EventHandler is a function that handles file system events
type EventHandler func(event Event)

// Watcher watches directories for file system events
type Watcher struct {
	watcher     *fsnotify.Watcher
	handlers    []EventHandler
	directories map[string]bool
	mu          sync.RWMutex
	done        chan struct{}
}

// NewWatcher creates a new file system watcher
func NewWatcher() (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		watcher:     fsWatcher,
		handlers:    make([]EventHandler, 0),
		directories: make(map[string]bool),
		done:        make(chan struct{}),
	}

	go w.watch()

	return w, nil
}

// AddHandler adds an event handler
func (w *Watcher) AddHandler(handler EventHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers = append(w.handlers, handler)
}

// AddDirectory adds a directory to watch
func (w *Watcher) AddDirectory(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if the directory exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Add the directory to the watcher
	if err := w.watcher.Add(path); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	// Add subdirectories recursively
	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := w.watcher.Add(subpath); err != nil {
				log.Printf("Warning: failed to add subdirectory to watcher: %v", err)
			}
			w.directories[subpath] = true
		}
		return nil
	})

	w.directories[path] = true
	return err
}

// RemoveDirectory removes a directory from the watcher
func (w *Watcher) RemoveDirectory(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Remove the directory from the watcher
	if err := w.watcher.Remove(path); err != nil {
		return fmt.Errorf("failed to remove directory from watcher: %w", err)
	}

	// Remove subdirectories
	for dir := range w.directories {
		if filepath.HasPrefix(dir, path) {
			if err := w.watcher.Remove(dir); err != nil {
				log.Printf("Warning: failed to remove subdirectory from watcher: %v", err)
			}
			delete(w.directories, dir)
		}
	}

	delete(w.directories, path)
	return nil
}

// Close closes the watcher
func (w *Watcher) Close() error {
	close(w.done)
	return w.watcher.Close()
}

// watch watches for file system events
func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching file system: %v", err)
		case <-w.done:
			return
		}
	}
}

// handleEvent handles a file system event
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// If a directory is created, add it to the watcher
	if event.Op&fsnotify.Create == fsnotify.Create {
		info, err := os.Stat(event.Name)
		if err == nil && info.IsDir() {
			w.mu.Lock()
			if err := w.watcher.Add(event.Name); err != nil {
				log.Printf("Warning: failed to add new directory to watcher: %v", err)
			} else {
				w.directories[event.Name] = true
			}
			w.mu.Unlock()
		}
	}

	// Convert fsnotify event to our event type
	var eventType EventType
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = EventCreate
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = EventModify
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = EventRemove
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = EventRename
	default:
		return // Ignore other events
	}

	// Create our event
	e := Event{
		Type: eventType,
		Path: event.Name,
	}

	// Call handlers
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, handler := range w.handlers {
		handler(e)
	}
}
