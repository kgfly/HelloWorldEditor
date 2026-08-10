package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnsavedBufferHasFriendlyTitle(t *testing.T) {
	var d Document
	if got := d.Title(); got != "untitled.txt" {
		t.Fatalf("Title() = %q, want untitled.txt", got)
	}
	if err := d.Save("hi"); err == nil {
		t.Fatal("Save() on a pathless buffer should fail, but it succeeded")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	var d Document
	got, err := d.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("Load() = %q, want hello", got)
	}

	d.Sync("hello there")
	if d.Title() != "notes.txt •" {
		t.Fatalf("dirty Title() = %q", d.Title())
	}

	if err := d.Save("world"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if d.Dirty() {
		t.Fatal("Save() should clear the dirty flag")
	}

	data, _ := os.ReadFile(path)
	if string(data) != "world" {
		t.Fatalf("file contains %q, want world", data)
	}
}

// Regression: loading a file pushes text into the editor, which emits a change
// event. That must not mark a freshly-opened, unmodified file as dirty.
func TestFreshlyLoadedFileIsNotDirty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("hello from disk"), 0o644); err != nil {
		t.Fatal(err)
	}

	var d Document
	content, err := d.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate the editor's change event fired by SetText during load.
	d.Sync(content)
	if d.Dirty() {
		t.Fatalf("freshly loaded file reported dirty; Title() = %q", d.Title())
	}

	// Editing dirties it, and reverting cleans it again.
	d.Sync("hello from disk!")
	if !d.Dirty() {
		t.Fatal("edited buffer should be dirty")
	}
	d.Sync("hello from disk")
	if d.Dirty() {
		t.Fatal("reverting to saved content should clear dirty")
	}
}

func TestLoadMissingFileStartsEmptyBuffer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brand-new.txt")

	var d Document
	got, err := d.Load(path)
	if err != nil {
		t.Fatalf("Load() of a missing file should succeed, got: %v", err)
	}
	if got != "" {
		t.Fatalf("Load() = %q, want empty", got)
	}
	if err := d.Save("first words"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file was not created: %v", err)
	}
}
