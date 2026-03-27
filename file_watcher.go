package main

// Module: file_watcher.go
// Purpose: Live file system watching via fsnotify
// Responsibilities:
// - Creating and managing fsnotify watcher instances
// - Listening for file system events (create, write, remove, rename)
// - Debouncing rapid changes (atomic writes: delete+recreate within 100ms)
// - Per-file deduplication (skip duplicate events within 200ms for same file)
// - Timer-based batching (collect events over 500ms window, send single refresh)
// - Delivering fileChangedMsg to Bubbletea via blocking tea.Cmd pattern
// - Restarting watcher on directory change, cleaning up on quit
//
// Ported from: ~/projects/markdown-themes/backend/websocket/filewatcher.go
// Key pattern: atomic write handling (editors often delete then recreate files)
//
// Architecture: Uses a bridging goroutine that reads fsnotify events, applies
// per-file deduplication, batching, and coalescing, then writes to a Go channel.
// A blocking tea.Cmd reads from that channel one message at a time (idiomatic
// Bubbletea subscription).
//
// Debounce strategy (handles AI agents & git operations generating many events):
//   1. Per-file dedup: skip events for the same file if < 200ms since last event
//   2. Batch window: first event starts a 500ms timer; new events reset the timer
//   3. Max batch delay: timer resets are capped at 2s from first event to prevent
//      indefinite starvation during sustained change storms
//   4. Atomic write detection: delete events wait 100ms to check for file re-creation

import (
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// watcherDebounceInterval is the time to wait after a delete event before
// concluding the file is truly gone (handles atomic writes by editors)
const watcherDebounceInterval = 100 * time.Millisecond

// watcherPerFileDedup is the minimum interval between events for the same file.
// Events arriving faster than this for a given path are silently dropped.
const watcherPerFileDedup = 200 * time.Millisecond

// watcherBatchWindow is the duration of the batching timer. On the first event,
// a timer is started. New events reset the timer (up to watcherMaxBatchDelay).
// When the timer fires, a single fileChangedMsg is sent.
const watcherBatchWindow = 500 * time.Millisecond

// watcherMaxBatchDelay is the maximum time from the first event in a batch to
// when the batch is flushed. This prevents indefinite starvation when events
// arrive continuously (e.g., git checkout touching hundreds of files).
const watcherMaxBatchDelay = 2 * time.Second

// watcherChanSize is the buffer size for the watcher event channel.
// Small buffer prevents backpressure from blocking fsnotify reads.
const watcherChanSize = 4

// initWatcher creates a new fsnotify watcher, the bridge channel, and stores
// them on the model. Does not start watching any path yet -- call startWatcher().
func (m *model) initWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// Non-fatal: TFE works fine without live watching
		return
	}
	m.watcher = watcher
	m.watcherChan = make(chan fileChangedMsg, watcherChanSize)
}

// startWatcher begins watching the given directory path for changes.
// It stops any existing watch, adds the new path to fsnotify, and spawns
// the bridge goroutine that processes events. Returns a tea.Cmd that
// listens for the first event on the bridge channel.
//
// The bridge goroutine handles:
//   - Per-file deduplication (skip events for same file within 200ms)
//   - Timer-based batching (500ms window, reset on new events, 2s max)
//   - Debouncing delete events (wait 100ms to check for atomic writes)
//   - Clean shutdown when the watcher is closed
func (m *model) startWatcher(path string) tea.Cmd {
	if m.watcher == nil {
		return nil
	}

	// Stop watching the previous path if any
	m.stopWatcher()

	// Add the new path to fsnotify
	if err := m.watcher.Add(path); err != nil {
		return nil
	}

	m.watchedPath = path
	m.watcherActive = true

	// Start the bridge goroutine that reads fsnotify events and writes
	// coalesced fileChangedMsg values to the channel
	go runWatcherBridge(m.watcher, m.watcherChan)

	// Return a tea.Cmd that blocks until the next event arrives
	return waitForWatcherEvent(m.watcherChan)
}

// waitForWatcherEvent returns a tea.Cmd that blocks until a fileChangedMsg
// arrives on the channel. This is the idiomatic Bubbletea subscription pattern:
// handle one message, then return another waitForWatcherEvent to keep listening.
func waitForWatcherEvent(ch <-chan fileChangedMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			// Channel closed -- watcher was shut down
			return nil
		}
		return msg
	}
}

// runWatcherBridge reads from fsnotify's event channel, applies per-file
// deduplication, timer-based batching, and atomic write detection, then writes
// a single coalesced fileChangedMsg to the output channel per batch.
//
// Debounce pipeline:
//   1. Per-file dedup: if the same file changed < 200ms ago, skip the event.
//   2. Batch timer: first accepted event starts a 500ms timer. Subsequent
//      events reset the timer (extending the window). When the timer fires,
//      one fileChangedMsg is sent representing the batch.
//   3. Max delay cap: timer resets are capped at 2s from the first event in
//      the batch so sustained storms still get periodic refreshes.
//   4. Atomic write: delete events wait 100ms before being classified.
//
// This goroutine exits when watcher.Events is closed (i.e., watcher.Close()
// is called). It also closes the output channel on exit so that any blocking
// waitForWatcherEvent unblocks with ok=false.
func runWatcherBridge(watcher *fsnotify.Watcher, out chan<- fileChangedMsg) {
	runWatcherBridgeWithConfig(watcher, out, watcherPerFileDedup, watcherBatchWindow, watcherMaxBatchDelay)
}

// runWatcherBridgeWithConfig is the configurable implementation of the bridge
// goroutine. It accepts tunable intervals so tests can use shorter durations.
func runWatcherBridgeWithConfig(
	watcher *fsnotify.Watcher,
	out chan<- fileChangedMsg,
	perFileDedup time.Duration,
	batchWindow time.Duration,
	maxBatchDelay time.Duration,
) {
	var (
		mu             sync.Mutex
		lastChangeTime = make(map[string]time.Time) // per-file dedup timestamps
		batchTimer     *time.Timer                  // fires to flush the current batch
		batchStartTime time.Time                    // when the current batch started
		// Track the "latest" event in the batch for the outgoing message.
		// For a directory watcher the exact path matters less than the fact
		// that something changed, but we keep the last path/op for context.
		batchPath string
		batchOp   fsnotify.Op
	)

	// acceptEvent checks per-file dedup and, if accepted, adds the event to
	// the current batch. Returns true if the event was accepted (not a dup).
	acceptEvent := func(path string, op fsnotify.Op) {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()

		// --- Per-file deduplication ---
		if lastTime, exists := lastChangeTime[path]; exists {
			if now.Sub(lastTime) < perFileDedup {
				// Duplicate for this file -- skip
				return
			}
		}
		lastChangeTime[path] = now

		// --- Batch timer management ---
		batchPath = path
		batchOp = op

		if batchTimer == nil {
			// First event in a new batch -- start the timer
			batchStartTime = now
			batchTimer = time.AfterFunc(batchWindow, func() {
				mu.Lock()
				p := batchPath
				o := batchOp
				batchTimer = nil
				// Clear per-file map to allow future events for these files
				// after the batch is flushed (prevents permanent suppression)
				for k := range lastChangeTime {
					delete(lastChangeTime, k)
				}
				mu.Unlock()
				select {
				case out <- fileChangedMsg{path: p, op: o}:
				default:
					// Channel full -- drop (next batch will pick up changes)
				}
			})
		} else {
			// Subsequent event in current batch -- reset timer if within max delay
			elapsed := now.Sub(batchStartTime)
			if elapsed < maxBatchDelay {
				// Reset: extend the batch window
				batchTimer.Reset(batchWindow)
			}
			// If elapsed >= maxBatchDelay, let the existing timer fire naturally
			// (it was set at most batchWindow ago, so it will fire soon)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				// Watcher closed -- clean up and exit
				mu.Lock()
				if batchTimer != nil {
					batchTimer.Stop()
				}
				mu.Unlock()
				close(out)
				return
			}

			// Filter out Chmod-only events (not interesting for file list refresh)
			if event.Op == fsnotify.Chmod {
				continue
			}

			// Handle delete events with atomic-write detection.
			// Editors like vim, nano, and many others save files by:
			//   1. Write to temp file
			//   2. Delete original
			//   3. Rename temp -> original
			// This causes a Delete event followed quickly by a Create.
			// We wait briefly before treating a delete as final.
			if event.Op&fsnotify.Remove != 0 {
				go func(path string) {
					time.Sleep(watcherDebounceInterval)

					// Check if file came back (atomic write pattern)
					if _, err := os.Stat(path); err == nil {
						// File exists again -- treat as a write/create
						acceptEvent(path, fsnotify.Write)
					} else {
						// File is really gone
						acceptEvent(path, fsnotify.Remove)
					}
				}(event.Name)
				continue
			}

			// For Create, Write, Rename -- add to batch
			acceptEvent(event.Name, event.Op)

		case _, ok := <-watcher.Errors:
			if !ok {
				return // Watcher closed
			}
			// Errors are non-fatal for a file explorer; silently continue.
			// Common errors: too many open files, permission denied on inotify.
		}
	}
}

// stopWatcher removes the currently watched path and marks the watcher as inactive.
// Does NOT close the underlying fsnotify.Watcher (use closeWatcher for full cleanup).
func (m *model) stopWatcher() {
	if m.watcher == nil || !m.watcherActive {
		return
	}

	if m.watchedPath != "" {
		m.watcher.Remove(m.watchedPath)
	}

	m.watcherActive = false
	m.watchedPath = ""
}

// closeWatcher fully shuts down the fsnotify watcher and releases resources.
// Call this on application quit.
func (m *model) closeWatcher() {
	if m.watcher == nil {
		return
	}

	m.stopWatcher()
	m.watcher.Close()
	m.watcher = nil
}

// switchWatchPath swaps the fsnotify watch from the current path to a new path
// without restarting the bridge goroutine. The goroutine continues reading from
// the same fsnotify watcher -- only the watched directory changes.
// This is used by navigateToPath() and other directory-change code paths.
func (m *model) switchWatchPath(newPath string) {
	if m.watcher == nil || !m.watcherActive {
		return
	}

	// Only switch if the path actually changed
	if m.watchedPath == newPath {
		return
	}

	// Remove old path
	if m.watchedPath != "" {
		m.watcher.Remove(m.watchedPath)
	}

	// Add new path
	if err := m.watcher.Add(newPath); err != nil {
		// Non-fatal: continue without watching
		m.watchedPath = ""
		return
	}

	m.watchedPath = newPath
}
