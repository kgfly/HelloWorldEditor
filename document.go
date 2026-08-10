// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Document owns everything about "which file are we editing and is it saved".
// The UI layer never touches os.* directly; it goes through here.
type Document struct {
	path  string // "" means the buffer has never been written to disk
	saved string // last known on-disk content, for dirty detection
	dirty bool
}

// Path returns the file's location on disk, or "" for an unsaved buffer.
func (d *Document) Path() string { return d.path }

// Dirty reports whether there are unsaved changes.
func (d *Document) Dirty() bool { return d.dirty }

// Sync compares the live buffer against the last on-disk content and updates
// the dirty flag. Diffing beats trusting change events: loading a file emits a
// change event without dirtying anything, and undoing back to the saved state
// correctly clears the marker.
func (d *Document) Sync(text string) { d.dirty = text != d.saved }

// Name is the short, human-friendly file name for the title bar.
func (d *Document) Name() string {
	if d.path == "" {
		return "untitled.txt"
	}
	return filepath.Base(d.path)
}

// Title renders the window/status caption, with the conventional "•" dirty marker.
func (d *Document) Title() string {
	if d.dirty {
		return d.Name() + " •"
	}
	return d.Name()
}

// Load reads path from disk and adopts it as the current document.
// A path that does not exist yet is treated as a brand-new empty file so that
// `helloworldeditor notes.txt` works the way every other editor behaves.
func (d *Document) Load(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", path, err)
	}
	data, err := os.ReadFile(abs)
	if os.IsNotExist(err) {
		d.path, d.saved, d.dirty = abs, "", false
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("open %s: %w", filepath.Base(abs), err)
	}
	d.path, d.saved, d.dirty = abs, string(data), false
	return string(data), nil
}

// Save writes text to the document's path. It fails loudly rather than
// silently inventing a filename when the buffer has never been saved.
func (d *Document) Save(text string) error {
	if d.path == "" {
		return fmt.Errorf("no filename: run the editor with a path, e.g. helloworldeditor notes.txt")
	}
	if err := os.WriteFile(d.path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("save %s: %w", d.Name(), err)
	}
	d.saved, d.dirty = text, false
	return nil
}
