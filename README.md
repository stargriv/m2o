# M2O Stream Player

A macOS command-line player for Radio M2O's live HLS stream. The primary
implementation is a Go program that plays through macOS AVFoundation and
prints the current track when it changes.

## Go standalone player (macOS)

The Go command in `go/cmd/m2o/` uses the Go standard library plus macOS system
frameworks (AVFoundation and AppKit). Build it once and copy `bin/m2o` to any
directory; it does not need Python, FFmpeg, PortAudio, or a Go installation at
runtime.

Requirements:

- macOS with a default audio output device
- Go 1.27 or later, to build the binary
- Internet access

```sh
make go-build
./bin/m2o
```

Or build and run in one command:

```sh
make go-run
```

Track changes are printed in this form:

```text
11:50 16.10.2026 Zedd - Clarity (feat. Foxes)
```

Press `Ctrl+C` to stop. The binary is macOS-only because it uses Apple system
frameworks for audio output.

Run the Go tests with:

```sh
make go-test
```

## Releases

The GitHub Actions release workflow packages binaries for macOS, Linux, and
Windows on both amd64 and arm64. Pushing a tag such as `v0.0.1` creates the
matching GitHub release; it can also be run manually with `v0.0.1` as the
default release tag.

Only the macOS binaries provide playback today because the player uses
AVFoundation. The Linux and Windows builds are published for distribution and
will report that no audio backend is available when run.

## Legacy Python experiment

`python/` contains the original Python prototype. It remains available for
comparison, but the Go player above is the recommended way to run the stream.

Requirements:

- Python 3
- A working default audio output device
- PortAudio (required by SoundDevice)
- FFmpeg libraries if PyAV cannot use a prebuilt wheel for the platform

On macOS, the native libraries can be installed with Homebrew:

```sh
brew install portaudio ffmpeg
```

### Setup

Create and activate a virtual environment, then install the pinned Python
dependencies:

```sh
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -r python/requirements.txt
```

### Run

Run:

```sh
make run
```

The script opens the Radio M2O HLS endpoint, converts audio to 48 kHz stereo
32-bit floating point, and plays it with high output latency to reduce the
chance of dropouts. Press `Ctrl+C` to stop playback.

An internet connection is required, and playback depends on the public stream
URL remaining available.

## Project files

| File | Purpose |
| --- | --- |
| `go/go.mod` | Go module definition. |
| `go/cmd/m2o/` | Go command, AVFoundation bridge, and Go tests. |
| `bin/m2o` | Generated standalone macOS binary (not committed). |
| `python/main.py` | Legacy Python callback/buffered playback prototype. |
| `python/tests/` | Python unit tests. |
| `python/requirements.txt` | Pinned Python dependencies. |

## Troubleshooting

- **No sound from the Go player:** confirm macOS has an active default output
  device and that its volume is not muted. The player uses the same output as
  other macOS media applications.
- **`start audio playback` error:** check the network connection; the public
  M2O HLS endpoint may be unavailable or changed.
- **Python audio-device or installation errors:** install PortAudio and FFmpeg
  with `brew install portaudio ffmpeg`, then repeat the Python setup.

## Status

This is a focused audio-playback application. It has automated tests for the
HTTP metadata, redirect handling, and title formatting, but it currently has
no command-line configuration or reconnect logic.
