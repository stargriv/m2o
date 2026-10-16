#!/usr/bin/env python

import av
import sys
import queue
import threading
from datetime import datetime
import requests
import sounddevice as sd

# Keep decoded audio ahead of the output callback to avoid playback gaps.
qlimit = 250
q = queue.Queue(qlimit)
buffersize = 3
# API endpoint that supplies both the stream URL and current song metadata.
url = "https://www.m2o.it/api/pub/v2/all/gdwc-audio-player/onair?format=json"
TRACK_POLL_INTERVAL = 20  # seconds

# Convert every decoded frame to the format expected by the audio device.
av_resample = av.AudioResampler(format='flt', layout='stereo', rate=48000)

def get_m2o_stream():
  """Resolve the current, redirectable M2O audio stream URL."""
  response = requests.get(url, timeout=10)
  response.raise_for_status()

  data = response.json()
  response = requests.get(data["audio_url"], timeout=10)
  m2o_stream_url = response.url
  return m2o_stream_url

def get_track(url):
  """Return the API title as a human-readable 'Artist - Song' string."""
  response = requests.get(url, timeout=10)
  response.raise_for_status()

  data = response.json()
  m2o_track_name = data["title"]
  words = m2o_track_name.split()

  # The artist is the uppercase block at the end.  Searching from the front
  # would mistake one-letter title words such as "I" for an artist.
  artist_start = len(words)
  for i in range(len(words) - 1, -1, -1):
    word = words[i]
    is_uppercase_word = word.isupper() and any(char.isalpha() for char in word)

    if is_uppercase_word or (word == "&" and artist_start < len(words)):
      artist_start = i
    else:
      break

  song = " ".join(words[:artist_start])
  artist = " ".join(words[artist_start:])
  track = artist.title()+" - "+song

  return(track)

def monitor_track(stop_event):
  """Print the current track only when it differs from the last result."""
  previous_track = None

  while not stop_event.is_set():
    try:
      track = get_track(url)
      if track != previous_track:
        timestamp = datetime.now().strftime("%H:%M %d.%m.%Y")
        print(f"{timestamp} {track}")
        previous_track = track
    except requests.RequestException as error:
      print(f"Could not fetch track information: {error}", file=sys.stderr)

    stop_event.wait(TRACK_POLL_INTERVAL)

def callback(outdata, frames, time, status):
  """Supply the audio device with the next frame prepared by the main loop."""
  assert not status
  if status.output_underflow:
    print('Output underflow: increase blocksize?', file=sys.stderr)
    raise sd.CallbackAbort
  try:
    data = q.get_nowait()
  except queue.Empty as e:
    print('Buffer is empty: increase buffersize?', file=sys.stderr)
    raise sd.CallbackAbort from e
  assert len(data) == len(outdata)
  outdata[:] = data

def main():
  """Start audio playback and the independent current-track monitor."""
  # Track polling runs separately so HTTP requests never block audio decoding.
  track_monitor_stop = threading.Event()
  track_monitor = threading.Thread(
    target=monitor_track,
    args=(track_monitor_stop,),
    daemon=True,
  )
  track_monitor.start()

  try:
    with av.open(get_m2o_stream(), options={"multiple_requests":"1"}) as container:
      av_stream = container.decode(audio=0)

      # Pre-fill the queue before opening the output device.
      for _,av_frame in enumerate(av_stream):
        if _ < buffersize:
          av_data = av_resample.resample(av_frame)[0].to_ndarray().reshape(-1,2)
          q.put_nowait(av_data)
        else:
          break

      # Decode subsequent frames continuously while the callback plays queued ones.
      with sd.OutputStream(samplerate=48000, blocksize=2048, channels=2, dtype='float32', latency='high', callback=callback):
        for av_frame in av_stream:
          av_data = av_resample.resample(av_frame)[0].to_ndarray().reshape(-1,2)
          q.put(av_data)

  except KeyboardInterrupt:
    print('\nInterrupted by user')
  finally:
    # Wake the monitor immediately instead of waiting for the next poll interval.
    track_monitor_stop.set()
    track_monitor.join(timeout=1)


if __name__ == "__main__":
  main()
