#!/usr/bin/env python

import av
import sys
import sounddevice as sd

av_resample = av.AudioResampler(format='flt', layout='stereo', rate=48000)
try:
  with av.open('https://4c4b867c89244861ac216426883d1ad0.msvdn.net/radiom2o/radiom2o/master_ma.m3u8', options={"multiple_requests":"1"}) as container:
    av_stream = container.decode(audio=0)
    with sd.OutputStream(samplerate=48000, channels=2, dtype='float32', latency='high') as sd_stream:
      for av_frame in av_stream:
        av_data = av_resample.resample(av_frame)[0].to_ndarray().reshape(-1,2)
        sd_stream.write(data=av_data)

except KeyboardInterrupt:
  sys.exit('\nInterrupted by user')