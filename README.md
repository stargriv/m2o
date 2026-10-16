# M2O Stream Player

A small Python experiment that plays the Radio M2O live stream through the
computer's default audio output. It uses [PyAV](https://pyav.org/) to decode
the HLS stream and [SoundDevice](https://python-sounddevice.readthedocs.io/)
to send resampled audio to the local sound device.

## Requirements

- Python 3
- A working default audio output device
- PortAudio (required by SoundDevice)
- FFmpeg libraries if PyAV cannot use a prebuilt wheel for the platform

On macOS, the native libraries can be installed with Homebrew:

```sh
brew install portaudio ffmpeg
```

## Setup

Create and activate a virtual environment, then install the pinned Python
dependencies:

```sh
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements.txt
```

## Play the stream

Run:

```sh
python stream_test.py
```

The script opens the Radio M2O HLS endpoint, converts audio to 48 kHz stereo
32-bit floating point, and plays it with high output latency to reduce the
chance of dropouts. Press `Ctrl+C` to stop playback.

An internet connection is required, and playback depends on the public stream
URL remaining available.

## Project files

| File | Purpose |
| --- | --- |
| `stream_test.py` | Current working streaming-player experiment. |
| `main.py` | Early callback/buffered playback prototype; it is incomplete. |
| `local.py` | Local transport-stream playback prototype; it is incomplete. |
| `requirements.txt` | Pinned Python dependencies. |

## Troubleshooting

- **No audio device / PortAudio error:** install PortAudio and make sure an
  output device is configured in your operating system.
- **PyAV installation fails:** install FFmpeg development libraries, then
  retry the dependency installation.
- **Playback stops or cannot open the stream:** check the network connection
  and confirm the embedded HLS URL is still reachable.

## Status

This repository is a focused audio-playback experiment rather than a polished
application. There is currently no command-line configuration, automated test
suite, or reconnect logic.
