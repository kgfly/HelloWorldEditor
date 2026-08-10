# HelloWorldEditor — Build & Run

Instructions assume a machine with **only the Go compiler installed**.

Project: `/home/k0g0kfq/data1/src/HelloWorldEditor`

| Target | Extra packages | cgo | Cross-compile from another OS? |
|---|---|---|---|
| **Windows** | none | no | yes, trivially |
| **macOS** | Xcode CLT only | yes | no (needs Apple toolchain) |
| **Linux X11** | 8 dev packages | yes | no (needs Linux headers) |
| **Linux Wayland** | 6 dev packages | yes | no |

Gio is pure Go for layout/widgets/GPU, but the window+input layer is cgo on
every platform **except Windows**. That is where the package lists come from.

---

## Windows

Nothing to install beyond Go. No C compiler, no headers — Gio drives Win32
through `golang.org/x/sys/windows`.

```powershell
cd path\to\HelloWorldEditor
go build -ldflags "-H windowsgui" -o helloworldeditor.exe .
.\helloworldeditor.exe notes.txt
```

`-H windowsgui` hides the console window that would otherwise sit behind the
GUI. Drop it if you want to see `stderr` while debugging.

Quick run without producing a binary: `go run . notes.txt`

The resulting `.exe` is self-contained and copies to any Windows box as-is.

---

## macOS

Only the Xcode command line tools (for `clang`; the full Xcode app is not needed):

```sh
xcode-select --install
```

```sh
cd path/to/HelloWorldEditor
go build -o helloworldeditor .
./helloworldeditor notes.txt
```

Everything else — Cocoa, Metal — ships with the OS.

### Voice typing (dictation)

Works with no extra code and no microphone permission: macOS performs the
speech-to-text itself and injects the finished text through `NSTextInputClient`,
which Gio implements. The app never touches the mic.

Enable it once in **System Settings → Keyboard → Dictation → On**, then press
`Fn` twice in the editor and talk.

### Optional: build a real `.app` bundle

A bare binary works, but a bundle gets a proper Dock icon and menu bar name:

```sh
go run gioui.org/cmd/gogio@latest -target macos -o HelloWorldEditor.app .
open HelloWorldEditor.app
```

### Note on cross-compiling

macOS builds require cgo, so they must be produced **on a Mac**. Building for
the other Mac architecture works if the toolchain supports it:

```sh
GOARCH=arm64 go build -o helloworldeditor-arm64 .   # Apple Silicon
GOARCH=amd64 go build -o helloworldeditor-amd64 .   # Intel
```

---

## Linux

Gio compiles **both** the X11 and Wayland backends by default and chooses at
runtime — Wayland is tried first, X11 is the fallback. A default build therefore
needs the headers for both.

Build tags let you drop a backend you do not want, along with its packages:

| Tag | Effect |
|---|---|
| `nowayland` | X11 only |
| `nox11` | Wayland only |
| `novulkan` | drop the Vulkan backend (OpenGL/EGL still used) |

### Option A — both backends (recommended: works everywhere)

Debian / Ubuntu:

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

Fedora / RHEL:

```sh
sudo dnf install -y pkgconf-pkg-config \
  libX11-devel libxcb-devel libXcursor-devel libXfixes-devel \
  libxkbcommon-devel libxkbcommon-x11-devel \
  wayland-devel mesa-libEGL-devel vulkan-loader-devel
```

Arch:

```sh
sudo pacman -S --needed pkgconf libx11 libxcursor libxfixes \
  libxkbcommon libxkbcommon-x11 wayland mesa vulkan-icd-loader
```

Alpine:

```sh
sudo apk add pkgconf gcc musl-dev libx11-dev libxcb-dev libxcursor-dev \
  libxfixes-dev libxkbcommon-dev wayland-dev mesa-dev vulkan-loader-dev
```

### Option B — X11 only

Skips all Wayland packages:

```sh
sudo apt-get install -y --no-install-recommends \
  pkg-config \
  libx11-dev libx11-xcb-dev libxcursor-dev libxfixes-dev \
  libxkbcommon-dev libxkbcommon-x11-dev \
  libegl-dev libvulkan-dev

go build -tags nowayland -o helloworldeditor .
./helloworldeditor notes.txt
```

### Option C — Wayland only

The smallest dependency set — no X11 packages at all:

```sh
sudo apt-get install -y --no-install-recommends \
  pkg-config libwayland-dev libwayland-egl-backend-dev \
  libxkbcommon-dev libegl-dev libvulkan-dev

go build -tags nox11 -o helloworldeditor .
./helloworldeditor notes.txt
```

`libxkbcommon-dev` is still required — Wayland uses it for keyboard layouts.

### Dropping Vulkan

`libvulkan-dev` is needed by default. If you would rather not install it,
build with `novulkan` (rendering falls back to OpenGL ES / EGL):

```sh
go build -tags novulkan -o helloworldeditor .
# combine freely:
go build -tags "nowayland,novulkan" -o helloworldeditor .
```

### Which backend am I actually running?

Wayland wins if `WAYLAND_DISPLAY` is set. Force X11 on a Wayland desktop
(XWayland) for a build that has both:

```sh
WAYLAND_DISPLAY= DISPLAY=:0 ./helloworldeditor notes.txt
```

### Runtime dependencies

A Linux build is **not** a self-contained binary — it links ~14 shared
libraries. To run it on another machine, that machine needs the runtime
libraries (the non-`-dev` packages: `libx11-6`, `libwayland-client0`,
`libxkbcommon0`, `libegl1`, …). Normal desktop installs already have them;
minimal containers do not.

Check what a build needs:

```sh
ldd helloworldeditor
```

---

## Cross-compiling summary

Only Windows cross-compiles cleanly, because it is the one target that needs no cgo:

```sh
# From Linux or macOS -> Windows. Works out of the box.
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -ldflags "-H windowsgui" -o helloworldeditor.exe .
```

Linux and macOS targets need their platform's C headers, so build them on the
platform itself (or in a matching container / CI runner).

---

## Tests

The file-handling logic has no Gio dependency, so tests run headlessly with no
display and no GUI packages:

```sh
go test ./...
```

Note: `go mod tidy` may fail behind a restrictive proxy because a **test-only**
dependency of Gio's text shaper (`typesetting-utils`) can be blocked. This does
not affect building. Use `go mod download` or `go get ./...` instead.

---

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `exec: "pkg-config": executable file not found` | install `pkg-config` |
| `Package 'x11-xcb' was not found` | install `libx11-xcb-dev` |
| `fatal error: vulkan/vulkan.h` | install `libvulkan-dev`, or build `-tags novulkan` |
| `fatal error: X11/Xlib.h` | install `libx11-dev`, or build `-tags nox11` |
| `fatal error: wayland-client.h` | install `libwayland-dev`, or build `-tags nowayland` |
| `build constraints exclude all Go files in gioui.org/internal/vk` | `CGO_ENABLED=0` on Linux — cgo is mandatory there |
| Console window behind the GUI on Windows | build with `-ldflags "-H windowsgui"` |
| Dictation does nothing on macOS | enable System Settings → Keyboard → Dictation; click into the text area first |

---

## Verified

Package lists were derived from Gio v0.10.1's `#cgo pkg-config` directives and
confirmed empirically on Ubuntu by removing packages until builds broke:

- Linux, both backends — builds and runs
- Linux, `-tags nowayland` — builds and runs; links no Wayland libraries
- Linux, `-tags nox11` — builds; links no X11 libraries
- Linux, `-tags novulkan` — builds without `libvulkan-dev`
- Windows cross-compile with `CGO_ENABLED=0` — builds
- `go test ./...` — passes

`wayland-protocols` and `wayland-scanner` are **not** required: Gio ships the
generated protocol C files in its source tree.

macOS instructions were derived from Gio's `os_macos.m` / `os_macos.go` source
and were **not** executed on a Mac.
