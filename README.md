# HelloWorldEditor

A minimal cross-platform text editor in ~290 lines of Go, built with [Gio](https://gioui.org).
One source tree → macOS, Linux, Windows. **macOS voice dictation works out of the box.**

| | |
|---|---|
| Open a file | `helloworldeditor notes.txt` |
| Save | `Ctrl+S` (Windows/Linux) · `Cmd+S` (macOS) |
| Dictate | Press `Fn` twice (macOS) |
| Unsaved marker | `notes.txt •` in the status bar |

Also free from Gio's editor widget: selection, copy/paste/cut, select-all, undo/redo
(`Ctrl/Cmd+Z`, `Shift+Ctrl/Cmd+Z`), word-wise delete, mouse selection.

---

## Build & run

```sh
go mod download
go build -o helloworldeditor .
./helloworldeditor notes.txt
```

Or without producing a binary: `go run . notes.txt`

### macOS

No extra dependencies — just Xcode command line tools (`xcode-select --install`).

```sh
go build -o helloworldeditor .
./helloworldeditor notes.txt
```

Cross-compiling for the other Mac architecture requires cgo, so build natively on
each, or use `GOARCH=arm64` / `GOARCH=amd64` with a suitable toolchain.

<details>
<summary>Optional: bundle as a real <code>.app</code></summary>

A plain binary works fine, but a bundle gets you a Dock icon and a proper app name:

```sh
go run gioui.org/cmd/gogio@latest -target macos -o HelloWorldEditor.app .
open HelloWorldEditor.app
```
</details>

### Linux

Needs X11/Wayland dev headers (Debian/Ubuntu):

```sh
sudo apt-get install -y --no-install-recommends \
  pkg-config \
  libx11-dev libx11-xcb-dev libxcursor-dev libxfixes-dev \
  libxkbcommon-dev libxkbcommon-x11-dev \
  libwayland-dev libwayland-egl-backend-dev \
  libegl-dev libvulkan-dev

go build -o helloworldeditor .
./helloworldeditor notes.txt
```

Gio compiles both the X11 and Wayland backends and picks at runtime, which is
why both sets of headers are needed. Build tags trim this: `-tags nowayland`
(X11 only), `-tags nox11` (Wayland only), `-tags novulkan` (skip `libvulkan-dev`).

Fedora: `sudo dnf install pkgconf-pkg-config libX11-devel libxcb-devel libXcursor-devel libXfixes-devel libxkbcommon-devel libxkbcommon-x11-devel wayland-devel mesa-libEGL-devel vulkan-loader-devel`

See [`~/doc/build-run.md`](../../../doc/build-run.md) for per-distro lists,
minimal per-backend package sets, and troubleshooting.

### Windows

No dependencies, and no cgo required.

```powershell
go build -o helloworldeditor.exe .
.\helloworldeditor.exe notes.txt
```

Hide the console window that appears behind the GUI:

```powershell
go build -ldflags "-H windowsgui" -o helloworldeditor.exe .
```

Cross-compile from Linux/macOS:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o helloworldeditor.exe .
```

---

## How macOS voice typing works

Nothing special was written for it. Gio's macOS backend implements the
`NSTextInputClient` protocol — the same input path used by IMEs and by system
Dictation. Text spoken into the mic is delivered through `insertText:` /
`setMarkedText:`, which Gio forwards to the focused editor exactly like keystrokes.

Two details in this app make it work smoothly on launch:

- The editor grabs keyboard focus on the first frame, so dictation has a target
  without needing a click first.
- Dictated text arrives as an ordinary `ChangeEvent`, so the dirty marker and
  save logic treat voice and keyboard input identically.

**Turn it on:** System Settings → Keyboard → Dictation → On, then press `Fn` twice
in the editor and start talking.

---

## Layout

| File | Role |
|---|---|
| `main.go` | Window creation and the event loop |
| `ui.go` | Layout, shortcuts, status bar |
| `document.go` | File I/O and dirty tracking — no UI types |
| `document_test.go` | Tests for the file logic |

`Document` knows nothing about Gio, which is why the file logic is testable
headlessly:

```sh
go test ./...
```

Dirty state is computed by diffing the buffer against the last saved content
rather than by listening for change events. Loading a file emits a change event
but must not mark it dirty, and undoing back to the saved text should clear the
marker — a boolean flag gets both cases wrong.

## Extending

- **Save As / Open dialogs** — Gio has no native file picker; use
  [`sqweek/dialog`](https://github.com/sqweek/dialog) or `gioui.org/x/explorer`.
- **Line numbers** — wrap the editor in a `layout.Flex` row and count `\n`.
- **Tabs** — make `Document` + `widget.Editor` a pair and hold a slice of them.
